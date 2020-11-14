package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
)

func handleSend(message *protos.Message) {
	logger.Info("Received send message",
		zap.String("target.connection", message.Target.Connection),
		zap.String("target.user", message.Target.User),
		zap.String("target.session", message.Target.Session),
		zap.String("target.channel", message.Target.Channel),
		zap.String("type", message.Type.String()),
	)

	// Resolve all local connections for message target
	connections, ok := resolveConnections(common.ResolveOptions{
		Connection: message.Target.Connection,
		User:       message.Target.User,
		Session:    message.Target.Session,
		Channel:    message.Target.Channel,
	})

	if !ok {
		return
	}

	// Send to all connections for target
	for _, connection := range connections {
		if connection.Sender == nil || connection.CloseChannel == nil {
			continue
		}

		connection := connection

		go func() {
			if message.Type == protos.Message_DISCONNECT {
				connection.CloseChannel <- struct{}{}
			} else {
				connection.Sender <- message
			}
		}()
	}
}

func sendMessageHandler(c *gin.Context) {
	requestId := requestid.Get(c)

	if c.ContentType() != common.ProtobufContentType {
		apiError := &common.ApiError{
			ErrorCode:  common.ErrorInvalidContentType,
			StatusCode: 400,
			RequestId:  requestId,
		}
		apiError.Send(c)
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		apiError := &common.ApiError{
			InternalError: err,
			ErrorCode:     common.ErrorReadingBody,
			StatusCode:    400,
			RequestId:     requestId,
		}
		apiError.Send(c)
		return
	}

	var message protos.Message

	err = proto.Unmarshal(body, &message)

	if err != nil {
		// Couldn't parse message
		apiError := &common.ApiError{
			InternalError: err,
			ErrorCode:     common.ErrorReadingBody,
			StatusCode:    400,
			RequestId:     requestId,
		}
		apiError.Send(c)
		return
	}

	handleSend(&message)

	c.AbortWithStatusJSON(200, gin.H{
		"success": true,
	})
}
