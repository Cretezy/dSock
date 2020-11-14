package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"strings"
)

func handleChannel(channelAction *protos.ChannelAction) {
	// Resolve all local connections for message target
	connections, ok := resolveConnections(common.ResolveOptions{
		Connection: channelAction.Target.Connection,
		User:       channelAction.Target.User,
		Session:    channelAction.Target.Session,
		Channel:    channelAction.Target.Channel,
	})

	if !ok {
		return
	}

	// Apply to all connections for target
	for _, connection := range connections {
		connectionChannels := connection.GetChannels()
		if channelAction.Type == protos.ChannelAction_SUBSCRIBE && !common.IncludesString(connectionChannels, channelAction.Channel) {
			connection.SetChannels(append(connectionChannels, channelAction.Channel))

			channels.Add(channelAction.Channel, connection.Id)

			redisClient.SAdd("channel:"+channelAction.Channel, connection.Id)
		} else if channelAction.Type == protos.ChannelAction_UNSUBSCRIBE && common.IncludesString(connectionChannels, channelAction.Channel) {
			connection.SetChannels(common.RemoveString(connectionChannels, channelAction.Channel))

			channels.Remove(channelAction.Channel, connection.Id)

			redisClient.SRem("channel:"+channelAction.Channel, connection.Id)
		} else {
			// Don't set in Redis
			return
		}

		redisClient.HSet("conn:"+connection.Id, "channels", strings.Join(connection.GetChannels(), ","))
	}
}

func channelMessageHandler(c *gin.Context) {
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

	var message protos.ChannelAction

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

	handleChannel(&message)

	c.AbortWithStatusJSON(200, gin.H{
		"success": true,
	})
}
