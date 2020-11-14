package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func disconnectHandler(c *gin.Context) {
	logger.Info("Getting disconnect request",
		zap.String("requestId", requestid.Get(c)),
		zap.String("id", c.Query("id")),
		zap.String("user", c.Query("user")),
		zap.String("session", c.Query("session")),
		zap.String("channel", c.Query("channel")),
		zap.String("keepClaims", c.Query("keepClaims")),
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

	keepClaims := c.Query("keepClaims") == "true"

	// Get all worker IDs that the target is connected to
	workerIds, apiError := resolveWorkers(resolveOptions, requestid.Get(c))
	if apiError != nil {
		apiError.Send(c)
		return
	}

	if !keepClaims {
		// Expire claims instantly, must resolve all claims for target
		claimIds, apiError := resolveClaims(resolveOptions, requestid.Get(c))

		if apiError != nil {
			apiError.Send(c)
			return
		}

		// Delete all resolved claims
		claimKeys := make([]string, len(claimIds))
		for index, claim := range claimIds {
			claimKeys[index] = "claim:" + claim
		}

		redisClient.SRem("claim-user:"+resolveOptions.User, claimIds)
		if resolveOptions.Session != "" {
			redisClient.SRem("claim-user-session:"+resolveOptions.User+"-"+resolveOptions.Session, claimIds)
		}
		if resolveOptions.Channel != "" {
			redisClient.SRem("claim-channel:"+resolveOptions.Channel, claimIds)
		}

		redisClient.Del(claimKeys...)
	}

	// Prepare message for worker
	message := &protos.Message{
		Type: protos.Message_DISCONNECT,
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

	logger.Info("Disconnected",
		zap.String("requestId", requestid.Get(c)),
		zap.Strings("workerIds", workerIds),
		zap.String("id", resolveOptions.Connection),
		zap.String("user", resolveOptions.User),
		zap.String("session", resolveOptions.Session),
		zap.String("channel", resolveOptions.Channel),
		zap.Bool("keepClaims", keepClaims),
	)

	c.AbortWithStatusJSON(200, map[string]interface{}{
		"success": true,
	})
}
