package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"strings"
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

	log.Printf("Channels: %v", authentication.Channels)

	// Upgrade to a WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Could not upgrade:", err)
		return
	}

	// Generate connection ID (random UUIDv4, can't be guessed)
	connId := uuid.New().String()

	// Channel that will be used to send messages to the client
	sender := make(chan *protos.Message)

	// Add to memory cache
	connection := SockConnection{
		Conn:         conn,
		Id:           connId,
		User:         authentication.User,
		Session:      authentication.Session,
		Sender:       sender,
		CloseChannel: make(chan struct{}),
	}
	connections[connId] = connection

	usersEntry, userExists := users[authentication.User]
	if userExists {
		users[authentication.User] = append(usersEntry, connId)
	} else {
		users[authentication.User] = []string{connId}
	}

	for _, channel := range authentication.Channels {
		channelEntry, channelExists := channels[channel]
		if channelExists {
			channels[channel] = append(channelEntry, connId)
		} else {
			channels[channel] = []string{connId}
		}

		redisClient.SAdd("channel:"+channel, connId)
	}

	// Add user/session to Redis
	redisConnection := map[string]interface{}{
		"user":     authentication.User,
		"workerId": workerId,
		"lastPing": time.Now().Format(time.RFC3339),
	}
	if authentication.Session != "" {
		redisConnection["session"] = authentication.Session
	}
	if len(authentication.Channels) != 0 {
		redisConnection["channels"] = strings.Join(authentication.Channels, ",")
	}
	redisClient.HSet("conn:"+connId, redisConnection)

	redisClient.SAdd("user:"+authentication.User, connId)
	if authentication.Session != "" {
		redisClient.SAdd("user-session:"+authentication.User+"-"+authentication.Session, connId)
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
			redisClient.SRem("user:"+authentication.User, connId)
			if authentication.Session != "" {
				redisClient.SRem("user-session:"+authentication.User+"-"+authentication.Session, connId)
			}

			delete(connections, connId)
			users[authentication.User] = common.RemoveString(users[authentication.User], connId)

			for _, channel := range authentication.Channels {
				channels[channel] = common.RemoveString(channels[channel], connId)

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
}
