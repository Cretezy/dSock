package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

func createClaimHandler(c *gin.Context) {
	user := c.Query("user")

	if user == "" {
		c.AbortWithStatusJSON(400, map[string]interface{}{
			"success":   false,
			"error":     "User ID is required",
			"errorCode": common.ErrorUserIdRequired,
		})
		return
	}

	session := c.Query("session")

	var expirationTime time.Time

	if c.Query("expiration") != "" {
		expiration, err := strconv.Atoi(c.Query("expiration"))

		if err != nil {
			c.AbortWithStatusJSON(400, map[string]interface{}{
				"success":   false,
				"error":     "Could not parse expiration (must be a integer)",
				"errorCode": common.ErrorInvalidExpiration,
			})
			return
		}

		if expiration < 1 {
			c.AbortWithStatusJSON(400, map[string]interface{}{
				"success":   false,
				"error":     "Can not use 0 or negative expiration",
				"errorCode": common.ErrorNegativeExpiration,
			})
			return
		}

		expirationTime = time.Unix(int64(expiration), 0)
	} else if c.Query("duration") != "" {
		duration, err := strconv.Atoi(c.Query("duration"))

		if err != nil {
			c.AbortWithStatusJSON(400, map[string]interface{}{
				"success":   false,
				"error":     "Could not parse duration (must be a integer)",
				"errorCode": common.ErrorInvalidDuration,
			})
			return
		}

		if duration < 1 {
			c.AbortWithStatusJSON(400, map[string]interface{}{
				"success": false,
				"error":   "Can not use 0 or negative duration",
			})
			return
		}

		expirationTime = time.Now().Add(time.Duration(duration) * time.Second)
	} else {
		expirationTime = time.Now().Add(time.Minute)
	}

	var id string

	if c.Query("id") != "" {
		id = c.Query("id")

		exists := redisClient.Exists("claim:" + id)

		if exists.Err() != nil {
			c.AbortWithStatusJSON(500, map[string]interface{}{
				"success":   false,
				"error":     "Error checking if claim already exists",
				"errorCode": common.ErrorNegativeDuration,
			})
			return
		}

		if exists.Val() == 1 {
			c.AbortWithStatusJSON(400, map[string]interface{}{
				"success":   false,
				"error":     "Claim ID is already used",
				"errorCode": common.ErrorClaimIdAlreadyUsed,
			})
			return
		}
	} else {
		id = common.RandomString(32)
	}

	claim := map[string]interface{}{
		"user":       user,
		"expiration": expirationTime.Format(time.RFC3339),
	}

	userKey := "claim-user:" + user
	redisClient.SAdd(userKey, id, 0)

	if session != "" {
		claim["session"] = session
		sessionKey := "claim-user-session:" + user + "-" + session
		redisClient.SAdd(sessionKey, id, 0)
	}

	claimKey := "claim:" + id
	redisClient.HSet(claimKey, claim)
	redisClient.ExpireAt(claimKey, expirationTime)

	c.AbortWithStatusJSON(200, map[string]interface{}{
		"success":    true,
		"id":         id,
		"expiration": expirationTime.Unix(),
	})
}
