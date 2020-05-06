package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
	"sync"
)

func disconnectHandler(c *gin.Context) {
	connId := c.Query("id")
	user := c.Query("user")
	session := c.Query("session")
	keepClaims := c.Query("keepClaims") == "true"

	resolveOptions := common.ResolveOptions{
		Connection: connId,
		User:       user,
		Session:    session,
	}

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

		redisClient.SRem("claim-user:"+user, claimIds)
		redisClient.SRem("claim-user-session:"+user+"-"+session, claimIds)

		redisClient.Del(claimKeys...)
	}

	// Prepare message for worker
	message := &protos.Message{
		User:       user,
		Connection: connId,
		Session:    session,
		Type:       protos.Message_DISCONNECT,
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

	c.AbortWithStatusJSON(200, map[string]interface{}{
		"success": true,
	})
}
