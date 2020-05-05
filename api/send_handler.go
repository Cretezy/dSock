package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"sync"
)

func sendHandler(c *gin.Context) {
	user := c.Query("user")
	connId := c.Query("id")

	workerIds, ok := resolveWorkers(c)

	if !ok {
		return
	}

	parsedMessageType := ParseMessageType(c.Query("type"))

	if parsedMessageType == -1 {
		c.AbortWithStatusJSON(400, map[string]interface{}{
			"success":   false,
			"error":     "Invalid message type, must be text or binary",
			"errorCode": common.ErrorInvalidMessageType,
		})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		c.AbortWithStatusJSON(500, map[string]interface{}{
			"success":   false,
			"error":     "Error reading message",
			"errorCode": common.ErrorReadingMessage,
		})
		return
	}

	message := &protos.Message{
		Body:       body,
		User:       user,
		Connection: connId,
		Session:    c.Query("session"),
		Type:       parsedMessageType,
	}

	rawMessage, err := proto.Marshal(message)

	if err != nil {
		c.AbortWithStatusJSON(500, map[string]interface{}{
			"success":   false,
			"error":     "Error marshalling message",
			"errorCode": common.ErrorMarshallingMessage,
		})
		return
	}

	var workersWaitGroup sync.WaitGroup

	workersWaitGroup.Add(len(workerIds))

	for _, workerId := range workerIds {
		workerId := workerId
		go func() {
			defer workersWaitGroup.Done()

			// Set: type, message, user, session
			redisClient.Publish(workerId, rawMessage)
		}()
	}

	workersWaitGroup.Wait()

	c.AbortWithStatusJSON(200, map[string]interface{}{
		"success": true,
	})
}

func ParseMessageType(messageType string) protos.Message_MessageType {
	switch messageType {
	case "text":
		fallthrough
	case "1":
		return protos.Message_TEXT
	case "binary":
		fallthrough
	case "2":
		return protos.Message_BINARY
	}

	return -1
}
