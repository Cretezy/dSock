package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func sendHandler(c *gin.Context) {
	resolveOptions := common.ResolveOptions{}

	err := c.BindQuery(&resolveOptions)
	if err != nil {
		apiError := &common.ApiError{
			InternalError: err,
			ErrorCode:     common.ErrorBindingQueryParams,
			StatusCode:    400,
		}
		apiError.Send(c)
		return
	}

	// Get all worker IDs that the target(s) is connected to
	workerIds, apiError := resolveWorkers(resolveOptions)
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
		Type: parsedMessageType,
		Body: body,
		Target: &protos.Target{
			Connection: resolveOptions.Connection,
			User:       resolveOptions.User,
			Session:    resolveOptions.Session,
			Channel:    resolveOptions.Channel,
		},
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
