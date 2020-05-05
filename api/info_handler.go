package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/gin-gonic/gin"
	"sync"
	"time"
)

func formatConnection(id string, connection map[string]string) gin.H {
	// Can safely ignore, will become 0
	lastPingTime, _ := time.Parse(time.RFC3339, connection["lastPing"])

	connectionMap := gin.H{
		"id":       id,
		"worker":   connection["workerId"],
		"lastPing": lastPingTime.Unix(),
		"user":     connection["user"],
	}

	if connection["session"] != "" {
		connectionMap["session"] = connection["session"]
	}

	return connectionMap
}

func formatClaim(id string, claim map[string]string) gin.H {
	// Can safely ignore, invalid times are already filtered out in infoHandler
	expirationTime, _ := time.Parse(time.RFC3339, claim["expiration"])

	claimMap := gin.H{
		"id":         id,
		"expiration": expirationTime.Unix(),
		"user":       claim["user"],
	}

	if claim["session"] != "" {
		claimMap["session"] = claim["session"]
	}

	return claimMap
}

func infoHandler(c *gin.Context) {
	connId := c.Query("id")
	user := c.Query("user")
	session := c.Query("session")

	var claimIds []string

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

		claimIds = userSessionClaims.Val()
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

		claimIds = userClaims.Val()
	}

	var claimsWaitGroup sync.WaitGroup
	claimsWaitGroup.Add(len(claimIds))

	var claimsLock sync.Mutex
	claims := make([]gin.H, 0)

	stop := false

	for _, claimId := range claimIds {
		claimId := claimId
		go func() {
			defer claimsWaitGroup.Done()

			claim := redisClient.HGetAll("claim:" + claimId)

			// Stop if one of the Redis commands failed
			if stop {
				return
			}

			if claim.Err() != nil {
				stop = true
				c.AbortWithStatusJSON(500, map[string]interface{}{
					"success":   false,
					"error":     "Error getting claim",
					"errorCode": common.ErrorGettingClaim,
				})
				return
			}

			if len(claim.Val()) == 0 {
				// Connection doesn't exist
				return
			}

			expirationTime, _ := time.Parse(time.RFC3339, claim.Val()["expiration"])

			if expirationTime.Before(time.Now()) {
				// Ignore invalid times (would become 0) or expired claims
				return
			}

			claimsLock.Lock()
			claims = append(claims, formatClaim(claimId, claim.Val()))
			claimsLock.Unlock()
		}()
	}

	claimsWaitGroup.Wait()

	if connId != "" {
		connection := redisClient.HGetAll("conn:" + connId)

		if connection.Err() != nil {
			c.AbortWithStatusJSON(500, map[string]interface{}{
				"success":   false,
				"error":     "Error getting connection",
				"errorCode": common.ErrorGettingConnection,
			})
			return
		}

		if len(connection.Val()) == 0 {
			// Connection doesn't exist
			c.AbortWithStatusJSON(200, map[string]interface{}{
				"success":     true,
				"connections": []interface{}{},
				"claims":      claims,
			})
			return
		}

		c.AbortWithStatusJSON(200, map[string]interface{}{
			"success":     true,
			"connections": []gin.H{formatConnection(connId, connection.Val())},
			"claims":      claims,
		})
		return
	} else if user != "" {
		user := redisClient.SMembers("user:" + user)

		if user.Err() != nil {
			c.AbortWithStatusJSON(500, map[string]interface{}{
				"success":   false,
				"error":     "Error getting user",
				"errorCode": common.ErrorGettingUser,
			})
			return
		}

		if len(user.Val()) == 0 {
			// User doesn't exist
			c.AbortWithStatusJSON(200, map[string]interface{}{
				"success":     true,
				"connections": []interface{}{},
				"claims":      claims,
			})
			return
		}

		var connectionsWaitGroup sync.WaitGroup
		connectionsWaitGroup.Add(len(user.Val()))

		var connectionsLock sync.Mutex
		connections := make([]gin.H, 0)

		stop := false

		for _, connId := range user.Val() {
			connId := connId
			go func() {
				defer connectionsWaitGroup.Done()

				connection := redisClient.HGetAll("conn:" + connId)

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

				if len(connection.Val()) == 0 {
					// Connection doesn't exist
					return
				}

				// Target specific session(s) for user
				if session != "" && connection.Val()["session"] != session {
					return
				}

				connectionsLock.Lock()
				connections = append(connections, formatConnection(connId, connection.Val()))
				connectionsLock.Unlock()
			}()
		}

		connectionsWaitGroup.Wait()

		c.AbortWithStatusJSON(200, map[string]interface{}{
			"success":     true,
			"connections": connections,
			"claims":      claims,
		})
		return
	} else {
		c.AbortWithStatusJSON(404, map[string]interface{}{
			"success":   false,
			"error":     "Connection ID or user ID is missing",
			"errorCode": common.ErrorMissingConnectionOrUser,
		})
		return
	}
}
