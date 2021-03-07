package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
)

func sendHandler(c *gin.Context) {
	logger.Info("Getting send request",
		zap.String("requestId", requestid.Get(c)),
		zap.String("id", c.Query("id")),
		zap.String("user", c.Query("user")),
		zap.String("session", c.Query("session")),
		zap.String("channel", c.Query("channel")),
	)

	resolveOptions := common.ResolveOptions{}

	err := c.BindQuery(&resolveOptions)
	if err != nil {
		apiError := &common.ApiError{
			InternalError: err,
			ErrorCode:     common.ErrorBindingQueryParams,
			StatusCode:    400,
			RequestId:     requestid.Get(c),
		}
		apiError.Send(c)
		return
	}

	// Get all worker IDs that the target(s) is connected to
	workerIds, apiError := resolveWorkers(resolveOptions, requestid.Get(c))
	if apiError != nil {
		apiError.Send(c)
		return
	}

	parsedMessageType := ParseMessageType(c.Query("type"))

	if parsedMessageType == -1 {
		apiError := common.ApiError{
			StatusCode: 400,
			ErrorCode:  common.ErrorInvalidMessageType,
			RequestId:  requestid.Get(c),
		}
		apiError.Send(c)
		return
	}

	// Read full body (message data)
	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		apiError := common.ApiError{
			InternalError: err,
			StatusCode:    500,
			ErrorCode:     common.ErrorReadingMessage,
			RequestId:     requestid.Get(c),
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
	apiError = sendToWorkers(workerIds, message, MessageMessageType, requestid.Get(c))
	if apiError != nil {
		apiError.Send(c)
		return
	}

	logger.Info("Sent message",
		zap.String("requestId", requestid.Get(c)),
		zap.Strings("workerIds", workerIds),
		zap.String("id", resolveOptions.Connection),
		zap.String("user", resolveOptions.User),
		zap.String("session", resolveOptions.Session),
		zap.String("channel", resolveOptions.Channel),
		zap.Int("bodyLength", len(body)),
	)

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
