package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"strings"
	"sync"
)

func getChannelHandler(actionType protos.ChannelAction_ChannelActionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		connId := c.Query("id")
		user := c.Query("user")
		session := c.Query("session")
		channel := c.Query("channel")
		channelChange := c.Param("channel")
		ignoreClaims := c.Query("ignoreClaims") == "true"

		resolveOptions := common.ResolveOptions{
			Connection: connId,
			User:       user,
			Session:    session,
			Channel:    channel,
		}

		// Get all worker IDs that the target(s) is connected to
		workerIds, apiError := resolveWorkers(resolveOptions)
		if apiError != nil {
			apiError.Send(c)
			return
		}

		if !ignoreClaims {
			// Add channel to all claims for the target
			claimIds, apiError := resolveClaims(resolveOptions)

			if apiError != nil {
				apiError.Send(c)
				return
			}

			// Update all resolved claims
			var claimWaitGroup sync.WaitGroup
			claimWaitGroup.Add(len(claimIds))

			for _, claimId := range claimIds {
				claimId := claimId
				go func() {
					defer claimWaitGroup.Done()
					claimKey := "claim:" + claimId
					// HGetAll instead of HGet to be able to check if claim exist
					claim := redisClient.HGetAll(claimKey)
					if apiError != nil {
						return
					}

					if claim.Err() != nil {
						apiError = &common.ApiError{
							InternalError: claim.Err(),
							ErrorCode:     common.ErrorGettingClaim,
							StatusCode:    500,
						}
					}

					if len(claim.Val()) == 0 {
						// Claim doesn't exist
						return
					}

					channels := common.RemoveEmpty(strings.Split(claim.Val()["channels"], ","))

					if actionType == protos.ChannelAction_SUBSCRIBE && !common.IncludesString(channels, channelChange) {
						channels = append(channels, channelChange)
						redisClient.SAdd("claim-channel:"+channelChange, claimId)
					} else if actionType != protos.ChannelAction_SUBSCRIBE && common.IncludesString(channels, channelChange) {
						channels = common.RemoveString(channels, channelChange)
						redisClient.SRem("claim-channel:"+channelChange, claimId)
					} else {
						return
					}

					redisClient.HSet(claimKey, "channels", strings.Join(channels, ","))
				}()
			}

			claimWaitGroup.Wait()

			if apiError != nil {
				apiError.Send(c)
				return
			}
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
