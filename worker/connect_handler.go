package main

import (
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

func connectHandler(c *gin.Context) {
	logger.Info("Getting new connection request",
		zap.String("requestId", requestid.Get(c)),
		zap.String("claim", c.Query("claim")),
		zap.String("jwt", c.Query("jwt")),
	)

	// Authenticate client and get user/session
	authentication, apiError := authenticate(c)
	if apiError != nil {
		apiError.Send(c)
		return
	}

	logger.Info("Authenticated connection request",
		zap.String("requestId", requestid.Get(c)),
		zap.String("user", authentication.User),
		zap.String("session", authentication.Session),
		zap.Strings("channels", authentication.Channels),
	)

	authentication.Channels = append(authentication.Channels, options.DefaultChannels...)

	// Upgrade to a WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warn("Could not upgrade request to WebSocket",
			zap.String("requestId", requestid.Get(c)),
			zap.Error(err),
		)
		return
	}

	// Generate connection ID (random UUIDv4, can't be guessed)
	connId := uuid.New().String()

	logger.Info("Upgraded connection request",
		zap.String("requestId", requestid.Get(c)),
		zap.String("id", connId),
	)

	// Channel that will be used to handleSend messages to the client
	sender := make(chan *protos.Message)

	// Add to memory cache
	connection := SockConnection{
		Conn:         conn,
		Id:           connId,
		User:         authentication.User,
		Session:      authentication.Session,
		Sender:       sender,
		CloseChannel: make(chan struct{}),
		channels:     authentication.Channels,
		lastPing:     time.Now(),
	}

	connections.Add(&connection)
	users.Add(connection.User, connId)
	for _, channel := range connection.channels {
		channels.Add(channel, connId)
	}

	connection.Refresh(redisClient)

	sendMutex := sync.Mutex{}

	// Send ping every minute
	go func() {
		for {
			time.Sleep(time.Second * 30)

			if connection.CloseChannel == nil {
				break
			}

			sendMutex.Lock()
			_ = conn.WriteMessage(websocket.PingMessage, []byte{})
			sendMutex.Unlock()
		}
	}()

	// Message receiving loop (from client)
	go func() {
	ReceiveLoop:
		for {
			messageType, _, err := conn.ReadMessage()

			if err != nil {
				// Disconnect on error
				if connection.CloseChannel != nil {
					connection.CloseChannel <- struct{}{}
				}
				break
			}

			switch messageType {
			case websocket.CloseMessage:
				if connection.CloseChannel != nil {
					connection.CloseChannel <- struct{}{}
				}
				break ReceiveLoop
			// Handling receiving ping/pong
			case websocket.PingMessage:
				fallthrough
			case websocket.PongMessage:
				connection.lock.Lock()
				connection.lastPing = time.Now()
				connection.lock.Unlock()

				connection.Refresh(redisClient)
				break
			}

		}
	}()

	// Message sending loop (to client, from sending channel)
SendLoop:
	for {
		select {
		case message := <-sender:
			sendMutex.Lock()
			_ = conn.WriteMessage(int(message.Type), message.Body)
			sendMutex.Unlock()
			break
		case <-connection.CloseChannel:
			logger.Info("Disconnecting user",
				zap.String("requestId", requestid.Get(c)),
				zap.String("id", connId),
			)

			connection.CloseChannel = nil

			sendMutex.Lock()

			// Send close message with 1000
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			// Sleep a tiny bit to allow message to be sent before closing connection
			time.Sleep(time.Millisecond)
			_ = conn.Close()

			redisClient.Del("conn:" + connId)
			redisClient.SRem("user:"+connection.User, connId)
			if connection.Session != "" {
				redisClient.SRem("user-session:"+connection.User+"-"+connection.Session, connId)
			}

			connections.Remove(connId)

			users.Remove(connection.User, connId)

			for _, channel := range connection.GetChannels() {
				channels.Remove(channel, connId)

				redisClient.SRem("channel:"+channel, connId)
			}

			break SendLoop
		}
	}
}

type SockConnection struct {
	/// WebSocket connection
	Conn    *websocket.Conn
	Id      string
	User    string
	Session string
	/// Message sending channel. Messages sent to it will be sent to the connection
	Sender chan *protos.Message
	/// Channel to close the connect. nil when connection is closed/closing
	CloseChannel chan struct{}
	channels     []string
	lastPing     time.Time
	lock         sync.RWMutex
}

func (connection *SockConnection) SetChannels(channels []string) {
	connection.lock.Lock()
	defer connection.lock.Unlock()

	connection.channels = channels
}

func (connection *SockConnection) GetChannels() []string {
	connection.lock.RLock()
	defer connection.lock.RUnlock()

	return connection.channels
}

func (connection *SockConnection) Refresh(redisCmdable redis.Cmdable) {
	connection.lock.RLock()
	defer connection.lock.RUnlock()

	redisConnection := map[string]interface{}{
		"user":     connection.User,
		"workerId": workerId,
		"lastPing": connection.lastPing.Format(time.RFC3339),
		"channels": strings.Join(connection.channels, ","),
	}
	if connection.Session != "" {
		redisConnection["session"] = connection.Session
	}

	redisCmdable.HSet("conn:"+connection.Id, redisConnection)
	redisCmdable.Expire("conn:"+connection.Id, options.TtlDuration*2)

	// Add user/session to Redis
	for _, channel := range connection.channels {
		redisCmdable.SAdd("channel:"+channel, connection.Id)
	}
	redisCmdable.SAdd("user:"+connection.User, connection.Id)
	if connection.Session != "" {
		redisCmdable.SAdd("user-session:"+connection.User+"-"+connection.Session, connection.Id)
	}
}
