package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

type Authentication struct {
	User     string
	Session  string
	Channels []string
}

func authenticate(c *gin.Context) (*Authentication, *common.ApiError) {
	if claim := c.Query("claim"); claim != "" {
		// Validate claim
		claimData := redisClient.HGetAll("claim:" + claim)

		if claimData.Err() != nil {
			return nil, &common.ApiError{
				InternalError: claimData.Err(),
				ErrorCode:     common.ErrorGettingClaim,
				StatusCode:    500,
				RequestId:     requestid.Get(c),
			}
		}

		if len(claimData.Val()) == 0 {
			// Claim doesn't exist
			return nil, &common.ApiError{
				ErrorCode:  common.ErrorMissingClaim,
				StatusCode: 400,
			}
		}

		user, hasUser := claimData.Val()["user"]
		if !hasUser {
			// Invalid claim (missing user)
			return nil, &common.ApiError{
				ErrorCode:  common.ErrorMissingClaim,
				StatusCode: 400,
				RequestId:  requestid.Get(c),
			}
		}

		expirationTime, err := time.Parse(time.RFC3339, claimData.Val()["expiration"])
		if err != nil {
			// Invalid expiration (can't parse)
			return nil, &common.ApiError{
				InternalError: err,
				ErrorCode:     common.ErrorInvalidExpiration,
				StatusCode:    500,
				RequestId:     requestid.Get(c),
			}
		}

		// Double check that claim is not expired
		if expirationTime.Before(time.Now()) {
			return nil, &common.ApiError{
				ErrorCode:  common.ErrorExpiredClaim,
				StatusCode: 400,
				RequestId:  requestid.Get(c),
			}
		}

		session := claimData.Val()["session"]

		// Expire claim instantly
		redisClient.Del("claim:" + claim)
		redisClient.SRem("claim-user:"+user, claim)
		if session != "" {
			redisClient.SRem("claim-user-session:"+user+"-"+session, claim)
		}
		for _, channel := range strings.Split(claimData.Val()["channels"], ",") {
			redisClient.SRem("claim-channel:"+channel, claim)
		}

		return &Authentication{
			User:     user,
			Session:  session,
			Channels: common.RemoveEmpty(strings.Split(claimData.Val()["channels"], ",")),
		}, nil
	} else if jwtToken := c.Query("jwt"); jwtToken != "" && options.Jwt.JwtSecret != "" {
		// Valid JWT (only enabled if `jwt_secret` is set)
		token, err := jwt.ParseWithClaims(jwtToken, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(options.Jwt.JwtSecret), nil
		})
		if err != nil {
			return nil, &common.ApiError{
				InternalError: err,
				ErrorCode:     common.ErrorInvalidJwt,
				StatusCode:    400,
				RequestId:     requestid.Get(c),
			}
		}

		// JWT claims, not "claim" as above
		claims := token.Claims.(*JwtClaims)

		return &Authentication{
			User:    claims.Subject,
			Session: claims.Session,
			Channels: common.UniqueString(common.RemoveEmpty(
				claims.Channels,
			)),
		}, nil
	} else {
		return nil, &common.ApiError{
			ErrorCode:  common.ErrorMissingAuthentication,
			StatusCode: 400,
			RequestId:  requestid.Get(c),
		}
	}
}
