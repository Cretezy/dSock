package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/gin-gonic/gin"
	"sync"
)

func resolveWorkers(c *gin.Context) ([]string, bool) {
	connId := c.Query("id")
	user := c.Query("user")

	workerIds := make([]string, 0)

	var workersLock sync.Mutex
	stop := false

	if connId != "" {
		connection := redisClient.HGetAll("conn:" + connId)

		workerId, hasWorkerId := connection.Val()["workerId"]

		if connection.Err() != nil {
			c.AbortWithStatusJSON(500, map[string]interface{}{
				"success":   false,
				"error":     "Error getting connection",
				"errorCode": common.ErrorGettingConnection,
			})
			return nil, false
		}

		if !hasWorkerId {
			// Connection doesn't exist
			return []string{}, true
		}

		workerIds = append(workerIds, workerId)
	} else if user != "" {
		user := redisClient.SMembers("user:" + user)

		if user.Err() != nil {
			c.AbortWithStatusJSON(500, map[string]interface{}{
				"success":   false,
				"error":     "Error getting user",
				"errorCode": common.ErrorGettingUser,
			})
			return nil, false
		}

		if len(user.Val()) == 0 {
			// User doesn't exist
			return []string{}, true
		}

		session := c.Query("session")

		var usersWaitGroup sync.WaitGroup

		usersWaitGroup.Add(len(user.Val()))

		for _, connId := range user.Val() {
			connId := connId
			go func() {
				defer usersWaitGroup.Done()

				connection := redisClient.HGetAll("conn:" + connId)

				workerId, hasWorkerId := connection.Val()["workerId"]

				// Stop if one of the Redis commands failed
				if stop {
					return
				}

				if connection.Err() != nil {
					stop = true
					c.AbortWithStatusJSON(500, map[string]interface{}{
						"success":   false,
						"error":     "Error getting connection",
						"errorCode": common.ErrorGettingConnection,
					})
					return
				}

				if !hasWorkerId {
					// Connection doesn't exist
					return
				}

				// Target specific session(s) for user
				if session != "" && connection.Val()["session"] != session {
					return
				}

				workersLock.Lock()
				workerIds = append(workerIds, workerId)
				workersLock.Unlock()
			}()
		}

		usersWaitGroup.Wait()
	} else {
		c.AbortWithStatusJSON(404, map[string]interface{}{
			"success":   false,
			"error":     "Connection ID or user ID is missing",
			"errorCode": common.ErrorMissingConnectionOrUser,
		})
		return nil, false
	}

	if stop {
		return nil, false
	}

	return common.UniqueString(workerIds), true
}
