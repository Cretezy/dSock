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
		apiError := common.ApiError{
			ErrorCode:  common.ErrorUserIdRequired,
			StatusCode: 400,
		}
		apiError.Send(c)
		return
	}

	session := c.Query("session")

	// Parses expiration time from expiration or duration
	var expirationTime time.Time

	if c.Query("expiration") != "" {
		expiration, err := strconv.Atoi(c.Query("expiration"))

		if err != nil {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorInvalidExpiration,
				StatusCode: 400,
			}
			apiError.Send(c)
			return
		}

		if expiration < 1 {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorNegativeExpiration,
				StatusCode: 400,
			}
			apiError.Send(c)
			return
		}

		expirationTime = time.Unix(int64(expiration), 0)
	} else if c.Query("duration") != "" {
		duration, err := strconv.Atoi(c.Query("duration"))

		if err != nil {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorInvalidDuration,
				StatusCode: 400,
			}
			apiError.Send(c)
			return
		}

		if duration < 1 {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorNegativeDuration,
				StatusCode: 400,
			}
			apiError.Send(c)
			return
		}

		expirationTime = time.Now().Add(time.Duration(duration) * time.Second)
	} else {
		expirationTime = time.Now().Add(time.Minute)
	}

	// Gets or generates claim ID
	var id string

	if c.Query("id") != "" {
		id = c.Query("id")

		exists := redisClient.Exists("claim:" + id)

		if exists.Err() != nil {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorCheckingClaim,
				StatusCode: 500,
			}
			apiError.Send(c)
			return
		}

		if exists.Val() == 1 {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorClaimIdAlreadyUsed,
				StatusCode: 400,
			}
			apiError.Send(c)
			return
		}
	} else {
		id = common.RandomString(32)
	}

	// Creates claim in Redis
	claim := map[string]interface{}{
		"user":       user,
		"expiration": expirationTime.Format(time.RFC3339),
	}
	if session != "" {
		claim["session"] = session

	}
	claimKey := "claim:" + id
	redisClient.HSet(claimKey, claim)
	redisClient.ExpireAt(claimKey, expirationTime)

	// Create user/session claim
	userKey := "claim-user:" + user
	redisClient.SAdd(userKey, id, 0)

	if session != "" {
		userSessionKey := "claim-user-session:" + user + "-" + session
		redisClient.SAdd(userSessionKey, id, 0)
	}

	c.AbortWithStatusJSON(200, map[string]interface{}{
		"success":    true,
		"id":         id,
		"expiration": expirationTime.Unix(),
	})
}
