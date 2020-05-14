package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"sync"
)

func disconnectHandler(c *gin.Context) {
	logger.Debug("Getting disconnect request",
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
		}
		apiError.Send(c)
		return
	}

	keepClaims := c.Query("keepClaims") == "true"

	// Get all worker IDs that the target is connected to
	workerIds, apiError := resolveWorkers(resolveOptions)
	if apiError != nil {
		apiError.Send(c)
		return
	}

	if !keepClaims {
		// Expire claims instantly, must resolve all claims for target
		claimIds, apiError := resolveClaims(resolveOptions)

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

	rawMessage, err := proto.Marshal(message)

	if err != nil {
		apiError := common.ApiError{
			ErrorCode:  common.ErrorMarshallingMessage,
			StatusCode: 500,
		}
		apiError.Send(c)
		return
	}

	// Send message to all resolved workers
	var workersWaitGroup sync.WaitGroup
	workersWaitGroup.Add(len(workerIds))

	for _, workerId := range workerIds {
		workerId := workerId
		go func() {
			defer workersWaitGroup.Done()

			redisClient.Publish(workerId, rawMessage)
		}()
	}

	workersWaitGroup.Wait()

	logger.Debug("Disconnected",
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
