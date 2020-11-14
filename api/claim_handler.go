package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type claimOptions struct {
	Id         string `form:"id"`
	User       string `form:"user"`
	Channels   string `form:"channels"`
	Session    string `form:"session"`
	Expiration string `form:"expiration"`
	Duration   string `form:"duration"`
}

func createClaimHandler(c *gin.Context) {
	logger.Info("Getting new claim request",
		zap.String("requestId", requestid.Get(c)),
		zap.String("id", c.Query("id")),
		zap.String("user", c.Query("user")),
		zap.String("session", c.Query("session")),
		zap.String("channels", c.Query("channels")),
		zap.String("expiration", c.Query("expiration")),
		zap.String("duration", c.Query("duration")),
	)

	claimOptions := claimOptions{}

	err := c.BindQuery(&claimOptions)
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

	channels := common.UniqueString(common.RemoveEmpty(
		strings.Split(claimOptions.Channels, ","),
	))

	if claimOptions.User == "" {
		apiError := common.ApiError{
			ErrorCode:  common.ErrorUserIdRequired,
			StatusCode: 400,
			RequestId:  requestid.Get(c),
		}
		apiError.Send(c)
		return
	}

	// Parses expiration time from expiration or duration
	var expirationTime time.Time

	if claimOptions.Expiration != "" {
		expiration, err := strconv.Atoi(claimOptions.Expiration)

		if err != nil {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorInvalidExpiration,
				StatusCode: 400,
				RequestId:  requestid.Get(c),
			}
			apiError.Send(c)
			return
		}

		if expiration < 1 {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorNegativeExpiration,
				StatusCode: 400,
				RequestId:  requestid.Get(c),
			}
			apiError.Send(c)
			return
		}

		expirationTime = time.Unix(int64(expiration), 0)

		if expirationTime.Before(time.Now()) {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorInvalidExpiration,
				StatusCode: 400,
				RequestId:  requestid.Get(c),
			}
			apiError.Send(c)
			return
		}
	} else if claimOptions.Duration != "" {
		duration, err := strconv.Atoi(claimOptions.Duration)

		if err != nil {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorInvalidDuration,
				StatusCode: 400,
				RequestId:  requestid.Get(c),
			}
			apiError.Send(c)
			return
		}

		if duration < 1 {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorNegativeDuration,
				StatusCode: 400,
				RequestId:  requestid.Get(c),
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

	if claimOptions.Id != "" {
		exists := redisClient.Exists("claim:" + claimOptions.Id)

		if exists.Err() != nil {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorCheckingClaim,
				StatusCode: 500,
				RequestId:  requestid.Get(c),
			}
			apiError.Send(c)
			return
		}

		if exists.Val() == 1 {
			apiError := common.ApiError{
				ErrorCode:  common.ErrorClaimIdAlreadyUsed,
				StatusCode: 400,
				RequestId:  requestid.Get(c),
			}
			apiError.Send(c)
			return
		}
	} else {
		id = common.RandomString(32)
	}

	// Creates claim in Redis
	claim := map[string]interface{}{
		"user":       claimOptions.User,
		"expiration": expirationTime.Format(time.RFC3339),
	}

	if claimOptions.Session != "" {
		claim["session"] = claimOptions.Session
	}

	if len(channels) != 0 {
		claim["channels"] = strings.Join(channels, ",")
	}

	claimKey := "claim:" + id
	redisClient.HSet(claimKey, claim)
	redisClient.ExpireAt(claimKey, expirationTime)

	// Create user/session claim
	userKey := "claim-user:" + claimOptions.User
	redisClient.SAdd(userKey, id, 0)

	if claimOptions.Session != "" {
		userSessionKey := "claim-user-session:" + claimOptions.User + "-" + claimOptions.Session
		redisClient.SAdd(userSessionKey, id, 0)
	}
	for _, channel := range channels {
		channelKey := "claim-channel:" + channel
		redisClient.SAdd(channelKey, id, 0)
	}

	logger.Info("Created new claim",
		zap.String("requestId", requestid.Get(c)),
		zap.String("id", id),
		zap.String("user", claimOptions.User),
		zap.Strings("channels", channels),
		zap.String("session", claimOptions.Session),
		zap.Time("expiration", expirationTime),
	)

	claimResponse := gin.H{
		"id":         id,
		"expiration": expirationTime.Unix(),
		"user":       claimOptions.User,
		"channels":   channels,
	}

	if claimOptions.Session != "" {
		claimResponse["session"] = claimOptions.Session
	}

	if len(channels) != 0 {
		claimResponse["channels"] = channels
	}

	c.AbortWithStatusJSON(200, map[string]interface{}{
		"success": true,
		"claim":   claimResponse,
	})
}
