package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strings"
	"sync"
)

var actionTypeName = map[protos.ChannelAction_ChannelActionType]string{
	protos.ChannelAction_SUBSCRIBE:   "subscribe",
	protos.ChannelAction_UNSUBSCRIBE: "unsubscribe",
}

func getChannelHandler(actionType protos.ChannelAction_ChannelActionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Getting channel request",
			zap.String("requestId", requestid.Get(c)),
			zap.String("action", actionTypeName[actionType]),
			zap.String("id", c.Query("id")),
			zap.String("user", c.Query("user")),
			zap.String("session", c.Query("session")),
			zap.String("channel", c.Query("channel")),
			zap.String("ignoreClaims", c.Query("ignoreClaims")),
			zap.String("channelChange", c.Param("channel")),
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

		channelChange := c.Param("channel")
		ignoreClaims := c.Query("ignoreClaims") == "true"

		// Get all worker IDs that the target(s) is connected to
		workerIds, apiError := resolveWorkers(resolveOptions, requestid.Get(c))
		if apiError != nil {
			apiError.Send(c)
			return
		}

		if !ignoreClaims {
			// Add channel to all claims for the target
			claimIds, apiError := resolveClaims(resolveOptions, requestid.Get(c))

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
							RequestId:     requestid.Get(c),
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
				Connection: resolveOptions.Connection,
				User:       resolveOptions.User,
				Session:    resolveOptions.Session,
				Channel:    resolveOptions.Channel,
			},
			Type: actionType,
		}

		// Send to all workers
		apiError = sendToWorkers(workerIds, message, ChannelMessageType, requestid.Get(c))
		if apiError != nil {
			apiError.Send(c)
			return
		}

		logger.Info("Set channel",
			zap.String("requestId", requestid.Get(c)),
			zap.String("action", actionTypeName[actionType]),
			zap.String("id", resolveOptions.Connection),
			zap.String("user", resolveOptions.User),
			zap.String("session", resolveOptions.Session),
			zap.String("channel", resolveOptions.Channel),
			zap.Bool("ignoreClaims", ignoreClaims),
			zap.String("channelChange", channelChange),
		)

		c.AbortWithStatusJSON(200, map[string]interface{}{
			"success": true,
		})
	}
}
