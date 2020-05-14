package main

import (
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"sync"
	"time"
)

func connectHandler(c *gin.Context) {
	// Authenticate client and get user/session
	authentication, apiError := authenticate(c)
	if apiError != nil {
		apiError.Send(c)
		return
	}

	authentication.Channels = append(authentication.Channels, options.DefaultChannels...)

	// Upgrade to a WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Could not upgrade:", err)
		return
	}

	// Generate connection ID (random UUIDv4, can't be guessed)
	connId := uuid.New().String()

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
		Channels:     authentication.Channels,
	}

	connections.Add(&connection)
	users.Add(connection.User, connId)

	for _, channel := range connection.Channels {
		channels.Add(channel, connId)

		redisClient.SAdd("channel:"+channel, connId)
	}

	// Add user/session to Redis
	redisConnection := map[string]interface{}{
		"user":     connection.User,
		"workerId": workerId,
		"lastPing": time.Now().Format(time.RFC3339),
		"channels":    strings.Join(connection.Channels, ","),
	}
	if connection.Session != "" {
		redisConnection["session"] = connection.Session
	}
	if len(connection.Channels) != 0 {
		redisConnection["channels"] = strings.Join(connection.Channels, ",")
	}
	redisClient.HSet("conn:"+connId, redisConnection)

	redisClient.SAdd("user:"+connection.User, connId)
	if connection.Session != "" {
		redisClient.SAdd("user-session:"+connection.User+"-"+connection.Session, connId)
	}

	// Send ping every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			if connection.CloseChannel == nil {
				break
			}

			_ = conn.WriteMessage(websocket.PingMessage, []byte{})
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
				redisClient.HSet(connId, "lastPing", time.Now().Format(time.RFC3339))
				break
			}

		}
	}()

	// Message sending loop (to client, from sending channel)
SendLoop:
	for {
		select {
		case message := <-sender:
			_ = conn.WriteMessage(int(message.Type), message.Body)
			break
		case <-connection.CloseChannel:
			connection.CloseChannel = nil
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

			for _, channel := range connection.Channels {
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
	Channels     []string
	lock         sync.Mutex
}

func (connection *SockConnection) SetChannels(channels []string) {
	connection.lock.Lock()
	defer connection.lock.Unlock()

	connection.Channels = channels
}
