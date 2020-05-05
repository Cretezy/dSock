package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
	"sync"
)

func disconnectHandler(c *gin.Context) {
	user := c.Query("user")
	session := c.Query("session")
	connId := c.Query("id")
	keepClaims := c.Query("keepClaims") == "true"

	workerIds, ok := resolveWorkers(c)

	if !ok {
		return
	}

	if !keepClaims {
		// Expire claims instantly
		var claims []string
		if session != "" {
			userSessionClaims := redisClient.SMembers("claim-user-session:" + user + "-" + session)

			if userSessionClaims.Err() != nil {
				c.AbortWithStatusJSON(500, map[string]interface{}{
					"success":   false,
					"error":     "Error getting claim",
					"errorCode": common.ErrorGettingClaim,
				})
				return
			}

			claims = userSessionClaims.Val()
		} else {
			userClaims := redisClient.SMembers("claim-user:" + user)

			if userClaims.Err() != nil {
				c.AbortWithStatusJSON(500, map[string]interface{}{
					"success":   false,
					"error":     "Error getting claim",
					"errorCode": common.ErrorGettingClaim,
				})
				return
			}

			claims = userClaims.Val()
		}

		claimKeys := make([]string, len(claims))
		for index, claim := range claims {
			claimKeys[index] = "claim:" + claim
		}

		redisClient.SRem("claim-user:"+user, claims)
		redisClient.SRem("claim-user-session:"+user+"-"+session, claims)

		redisClient.Del(claimKeys...)
	}

	message := &protos.Message{
		User:       user,
		Connection: connId,
		Session:    session,
		Type:       protos.Message_DISCONNECT,
	}

	rawMessage, err := proto.Marshal(message)

	if err != nil {
		c.AbortWithStatusJSON(500, map[string]interface{}{
			"success":   false,
			"error":     "Error marshalling message",
			"errorCode": common.ErrorMarshallingMessage,
		})
		return
	}

	var workersWaitGroup sync.WaitGroup

	workersWaitGroup.Add(len(workerIds))

	for _, workerId := range workerIds {
		workerId := workerId
		go func() {
			defer workersWaitGroup.Done()

			// Set: type, message, user, session
			redisClient.Publish(workerId, rawMessage)
		}()
	}

	workersWaitGroup.Wait()

	c.AbortWithStatusJSON(200, map[string]interface{}{
		"success": true,
	})
}
