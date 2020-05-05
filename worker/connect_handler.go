package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func connectHandler(c *gin.Context) {
	authentication := authenticate(c)
	if authentication == nil {
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	connId := uuid.New().String()

	// TODO: Add authentication

	sender := make(chan *protos.Message)

	// Add to cache
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

	// Add user/session to Redis
	redisConnection := map[string]interface{}{
		"user":     authentication.User,
		"workerId": workerId,
		"lastPing": time.Now().Format(time.RFC3339),
	}
	redisClient.SAdd("user:"+authentication.User, connId)
	if authentication.Session != "" {
		redisConnection["session"] = authentication.Session
		redisClient.SAdd("user-session:"+authentication.User+"-"+authentication.Session, connId)
	}
	redisClient.HSet("conn:"+connId, redisConnection)

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

	go func() {
	ReceiveLoop:
		for {
			messageType, _, err := conn.ReadMessage()

			if err != nil {
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

SendLoop:
	for {
		select {
		case message := <-sender:
			_ = conn.WriteMessage(int(message.Type), message.Body)
			break
		case <-connection.CloseChannel:
			connection.CloseChannel = nil
			_ = conn.Close()

			redisClient.Del("conn:" + connId)
			redisClient.SRem("user:"+authentication.User, connId)
			if authentication.Session != "" {
				redisClient.SRem("user-session:"+authentication.User+"-"+authentication.Session, connId)
			}

			delete(connections, connId)
			users[authentication.User] = common.RemoveString(users[authentication.User], connId)

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
