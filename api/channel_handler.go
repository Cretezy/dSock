package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
)

func getChannelHandler(actionType protos.ChannelAction_ChannelActionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		connId := c.Query("id")
		user := c.Query("user")
		session := c.Query("session")
		channel := c.Query("channel")
		channelChange := c.Param("channel")

		// Get all worker IDs that the target(s) is connected to
		workerIds, apiError := resolveWorkers(common.ResolveOptions{
			Connection: connId,
			User:       user,
			Session:    session,
			Channel:    channel,
		})
		if apiError != nil {
			apiError.Send(c)
			return
		}

		// Prepare message for worker
		message := &protos.ChannelAction{
			Channel: channelChange,
			Target: &protos.Target{
				Connection: connId,
				User:       user,
				Session:    session,
				Channel:    channel,
			},
			Type: actionType,
		}

		// Build Redis channel names for $workerId:channel
		workerChannels := make([]string, len(workerIds))
		for index, workerId := range workerIds {
			workerChannels[index] = workerId + ":channel"
		}

		// Send to all workers
		apiError = sendToWorkers(workerChannels, message)
		if apiError != nil {
			apiError.Send(c)
			return
		}

		c.AbortWithStatusJSON(200, map[string]interface{}{
			"success": true,
		})
	}
}
