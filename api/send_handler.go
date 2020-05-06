package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func sendHandler(c *gin.Context) {
	connId := c.Query("id")
	user := c.Query("user")
	session := c.Query("session")

	// Get all worker IDs that the target is connected to
	workerIds, apiError := resolveWorkers(common.ResolveOptions{
		Connection: connId,
		User:       user,
		Session:    session,
	})
	if apiError != nil {
		apiError.Send(c)
		return
	}

	parsedMessageType := ParseMessageType(c.Query("type"))

	if parsedMessageType == -1 {
		apiError := common.ApiError{
			StatusCode: 400,
			ErrorCode:  common.ErrorInvalidMessageType,
		}
		apiError.Send(c)
		return
	}

	// Read full body (message data)
	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		apiError := common.ApiError{
			StatusCode: 500,
			ErrorCode:  common.ErrorReadingMessage,
		}
		apiError.Send(c)
		return
	}

	// Prepare message for worker
	message := &protos.Message{
		Body:       body,
		User:       user,
		Connection: connId,
		Session:    c.Query("session"),
		Type:       parsedMessageType,
	}

	// Send to all workers
	apiError = sendToWorkers(workerIds, message)
	if apiError != nil {
		apiError.Send(c)
		return
	}

	c.AbortWithStatusJSON(200, map[string]interface{}{
		"success": true,
	})
}

/// Parse message type, allowing for WebSocket frame type ID
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
