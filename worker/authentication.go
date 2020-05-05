package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"time"
)

type Authentication struct {
	User    string
	Session string
}

func authenticate(c *gin.Context) *Authentication {
	if claim := c.Query("claim"); claim != "" {

		claimData := redisClient.HGetAll("claim:" + claim)

		if claimData.Err() != nil {
			c.AbortWithStatusJSON(500, map[string]interface{}{
				"success":   false,
				"error":     "Error getting claim",
				"errorCode": common.ErrorGettingClaim,
			})
			return nil
		}

		user, hasUser := claimData.Val()["user"]

		if !hasUser {
			c.AbortWithStatusJSON(400, map[string]interface{}{
				"success":   false,
				"error":     "Could not find claim",
				"errorCode": common.ErrorMissingClaim,
			})
			return nil
		}

		expirationTime, err := time.Parse(time.RFC3339, claimData.Val()["expiration"])

		if err != nil {
			c.AbortWithStatusJSON(500, map[string]interface{}{
				"success":   false,
				"error":     "Error parsing expiration",
				"errorCode": common.ErrorInvalidExpiration,
			})
			return nil
		}

		// Double check that claim is not expired
		if expirationTime.Before(time.Now()) {
			c.AbortWithStatusJSON(400, map[string]interface{}{
				"success":   false,
				"error":     "Claim has expired",
				"errorCode": common.ErrorExpiredClaim,
			})
			return nil
		}

		session := claimData.Val()["session"]

		// Expire claim instantly
		redisClient.Del("claim:" + claim)
		redisClient.SRem("claim-user:"+user, claim)
		if session != "" {
			redisClient.SRem("claim-user-session:"+user+"-"+session, claim)
		}

		return &Authentication{
			User:    user,
			Session: session,
		}
	} else if jwtToken := c.Query("jwt"); jwtToken != "" && options.Jwt.JwtSecret != "" {
		token, err := jwt.ParseWithClaims(jwtToken, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(options.Jwt.JwtSecret), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(400, map[string]interface{}{
				"success":   false,
				"error":     "Could not validate JWT",
				"errorCode": common.ErrorInvalidJwt,
			})
			return nil
		}

		claims := token.Claims.(*JwtClaims)

		return &Authentication{
			User:    claims.Subject,
			Session: claims.Session,
		}
	} else {
		c.AbortWithStatusJSON(400, map[string]interface{}{
			"success":   false,
			"error":     "Did not provide an authentication method",
			"errorCode": common.ErrorMissingAuthentication,
		})
		return nil
	}
}
