package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
	"strings"
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
		"channels": strings.Split(connection["channels"], ","),
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
		"channels":   strings.Split(claim["channels"], ","),
	}

	if claim["session"] != "" {
		claimMap["session"] = claim["session"]
	}

	return claimMap
}

func infoHandler(c *gin.Context) {
	logger.Info("Getting info request",
		zap.String("requestId", requestid.Get(c)),
		zap.String("id", c.Query("id")),
		zap.String("user", c.Query("user")),
		zap.String("session", c.Query("session")),
		zap.String("channel", c.Query("channel")),
	)

	resolveOptions := common.ResolveOptions{}

	err := c.BindQuery(&resolveOptions)
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

	claimIds, apiError := resolveClaims(resolveOptions, requestid.Get(c))
	if apiError != nil {
		apiError.Send(c)
		return
	}

	claimCmds := make([]*redis.StringStringMapCmd, len(claimIds))
	_, err = redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
		for index, claimId := range claimIds {
			claimCmds[index] = pipeliner.HGetAll("claim:" + claimId)
		}

		return nil
	})

	if err != nil {
		apiError := &common.ApiError{
			InternalError: err,
			ErrorCode:     common.ErrorGettingClaim,
			StatusCode:    500,
			RequestId:     requestid.Get(c),
		}
		apiError.Send(c)
		return
	}

	claims := make([]gin.H, 0)

	for index, claimId := range claimIds {
		claim := claimCmds[index]

		if claim.Err() != nil {
			apiError := &common.ApiError{
				InternalError: claim.Err(),
				ErrorCode:     common.ErrorGettingClaim,
				StatusCode:    500,
				RequestId:     requestid.Get(c),
			}
			apiError.Send(c)
			return
		}

		if len(claim.Val()) == 0 {
			// Connection doesn't exist
			continue
		}

		expirationTime, _ := time.Parse(time.RFC3339, claim.Val()["expiration"])

		if expirationTime.Before(time.Now()) {
			// Ignore invalid times (would become 0) or expired claims
			return
		}

		claims = append(claims, formatClaim(claimId, claim.Val()))
	}

	// Get connection(s)
	if resolveOptions.Connection != "" {
		connection := redisClient.HGetAll("conn:" + resolveOptions.Connection)

		if connection.Err() != nil {
			apiError := common.ApiError{
				InternalError: connection.Err(),
				StatusCode:    500,
				ErrorCode:     common.ErrorGettingConnection,
				RequestId:     requestid.Get(c),
			}
			apiError.Send(c)
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
			"connections": []gin.H{formatConnection(resolveOptions.Connection, connection.Val())},
			"claims":      claims,
		})
	} else if resolveOptions.User != "" {
		user := redisClient.SMembers("user:" + resolveOptions.User)

		if user.Err() != nil {
			apiError := common.ApiError{
				InternalError: user.Err(),
				StatusCode:    500,
				ErrorCode:     common.ErrorGettingUser,
				RequestId:     requestid.Get(c),
			}
			apiError.Send(c)
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

		var connectionCmds = make([]*redis.StringStringMapCmd, len(user.Val()))
		_, err = redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
			for index, connId := range user.Val() {
				connectionCmds[index] = pipeliner.HGetAll("conn:" + connId)
			}

			return nil
		})

		if err != nil {
			apiError := &common.ApiError{
				InternalError: err,
				StatusCode:    500,
				ErrorCode:     common.ErrorGettingConnection,
				RequestId:     requestid.Get(c),
			}
			apiError.Send(c)
			return
		}

		connections := make([]gin.H, 0)

		for index, connId := range user.Val() {
			connId := connId

			connection := connectionCmds[index]

			if connection.Err() != nil {
				apiError = &common.ApiError{
					InternalError: connection.Err(),
					StatusCode:    500,
					ErrorCode:     common.ErrorGettingConnection,
					RequestId:     requestid.Get(c),
				}
				apiError.Send(c)

				return
			}

			if len(connection.Val()) == 0 {
				// Connection doesn't exist
				continue
			}

			// Target specific session(s) for user
			if resolveOptions.Session != "" && connection.Val()["session"] != resolveOptions.Session {
				return
			}

			connections = append(connections, formatConnection(connId, connection.Val()))
		}

		c.AbortWithStatusJSON(200, map[string]interface{}{
			"success":     true,
			"connections": connections,
			"claims":      claims,
		})
	} else if resolveOptions.Channel != "" {
		channel := redisClient.SMembers("channel:" + resolveOptions.Channel)

		if channel.Err() != nil {
			apiError := common.ApiError{
				InternalError: channel.Err(),
				StatusCode:    500,
				ErrorCode:     common.ErrorGettingChannel,
				RequestId:     requestid.Get(c),
			}
			apiError.Send(c)
			return
		}

		if len(channel.Val()) == 0 {
			// User doesn't exist
			c.AbortWithStatusJSON(200, map[string]interface{}{
				"success":     true,
				"connections": []interface{}{},
				"claims":      claims,
			})
			return
		}

		var connectionCmds = make([]*redis.StringStringMapCmd, len(channel.Val()))
		_, err = redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
			for index, connId := range channel.Val() {
				connectionCmds[index] = pipeliner.HGetAll("conn:" + connId)
			}

			return nil
		})

		connections := make([]gin.H, 0)

		for index, connId := range channel.Val() {
			connection := connectionCmds[index]

			if connection.Err() != nil {
				apiError = &common.ApiError{
					InternalError: connection.Err(),
					StatusCode:    500,
					ErrorCode:     common.ErrorGettingConnection,
					RequestId:     requestid.Get(c),
				}
				return
			}

			if len(connection.Val()) == 0 {
				// Connection doesn't exist
				return
			}

			connections = append(connections, formatConnection(connId, connection.Val()))
		}

		c.AbortWithStatusJSON(200, gin.H{
			"success":     true,
			"connections": connections,
			"claims":      claims,
		})
	} else {
		apiError := common.ApiError{
			StatusCode: 400,
			ErrorCode:  common.ErrorTarget,
			RequestId:  requestid.Get(c),
		}
		apiError.Send(c)
	}
}
