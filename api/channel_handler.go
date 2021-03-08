package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
	"strings"
)

var actionTypeName = map[protos.ChannelAction_ChannelActionType]string{
	protos.ChannelAction_SUBSCRIBE:   "subscribe",
	protos.ChannelAction_UNSUBSCRIBE: "unsubscribe",
}

func getChannelHandler(actionType protos.ChannelAction_ChannelActionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := requestid.Get(c)

		logger.Info("Getting channel request",
			zap.String("requestId", requestId),
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
				RequestId:     requestId,
			}
			apiError.Send(c)
			return
		}

		channelChange := c.Param("channel")
		ignoreClaims := c.Query("ignoreClaims") == "true"

		// Get all worker IDs that the target(s) is connected to
		workerIds, apiError := resolveWorkers(resolveOptions, requestId)
		if apiError != nil {
			apiError.Send(c)
			return
		}

		if !ignoreClaims {
			// Add channel to all claims for the target
			claimIds, apiError := resolveClaims(resolveOptions, requestId)

			if apiError != nil {
				apiError.Send(c)
				return
			}

			var claimCmds = make([]*redis.StringStringMapCmd, len(claimIds))
			_, err := redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
				for index, claimId := range claimIds {
					// HGetAll instead of HGet to be able to check if claim exist
					claimCmds[index] = pipeliner.HGetAll("claim:" + claimId)
				}

				return nil
			})

			if err != nil {
				apiError = &common.ApiError{
					InternalError: err,
					ErrorCode:     common.ErrorGettingClaim,
					StatusCode:    500,
					RequestId:     requestId,
				}
				apiError.Send(c)

				return
			}

			_, err = redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
				// Update all resolved claims
				for index, claimId := range claimIds {
					claimKey := "claim:" + claimId
					claim := claimCmds[index]

					if len(claim.Val()) == 0 {
						// Claim doesn't exist
						continue
					}

					channels := common.RemoveEmpty(strings.Split(claim.Val()["channels"], ","))

					if actionType == protos.ChannelAction_SUBSCRIBE && !common.IncludesString(channels, channelChange) {
						channels = append(channels, channelChange)
						pipeliner.SAdd("claim-channel:"+channelChange, claimId)
					} else if actionType != protos.ChannelAction_SUBSCRIBE && common.IncludesString(channels, channelChange) {
						channels = common.RemoveString(channels, channelChange)
						pipeliner.SRem("claim-channel:"+channelChange, claimId)
					} else {
						continue
					}

					pipeliner.HSet(claimKey, "channels", strings.Join(channels, ","))
				}

				return nil
			})

			if err != nil {
				apiError = &common.ApiError{
					InternalError: err,
					// TODO: Improve error
					ErrorCode:  common.ErrorGettingChannel,
					StatusCode: 500,
					RequestId:  requestId,
				}
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
		apiError = sendToWorkers(workerIds, message, ChannelMessageType, requestId)
		if apiError != nil {
			apiError.Send(c)
			return
		}

		logger.Info("Set channel",
			zap.String("requestId", requestId),
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
