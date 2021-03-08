package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/go-redis/redis/v7"
)

/// Resolves the workers holding the connection
func resolveWorkers(options common.ResolveOptions, requestId string) ([]string, *common.ApiError) {
	workerIds := make([]string, 0)

	if options.Connection != "" {
		// Get a connection by connection ID
		connection := redisClient.HGetAll("conn:" + options.Connection)

		if connection.Err() != nil {
			return nil, &common.ApiError{
				InternalError: connection.Err(),
				StatusCode:    500,
				ErrorCode:     common.ErrorGettingConnection,
			}
		}

		if len(connection.Val()) == 0 {
			// Connection doesn't exist
			return []string{}, nil
		}

		workerId, hasWorkerId := connection.Val()["workerId"]

		if !hasWorkerId {
			// Is missing worker ID, ignoring
			return []string{}, nil
		}

		workerIds = append(workerIds, workerId)
	} else if options.Channel != "" {
		// Get all connections for a channel
		channel := redisClient.SMembers("channel:" + options.Channel)

		if channel.Err() != nil {
			return nil, &common.ApiError{
				InternalError: channel.Err(),
				StatusCode:    500,
				ErrorCode:     common.ErrorGettingChannel,
				RequestId:     requestId,
			}
		}

		if len(channel.Val()) == 0 {
			// User doesn't exist
			return []string{}, nil
		}

		var connectionCmds = make([]*redis.StringStringMapCmd, len(channel.Val()))
		_, err := redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
			for index, connId := range channel.Val() {
				connectionCmds[index] = pipeliner.HGetAll("conn:" + connId)
			}

			return nil
		})

		if err != nil {
			return nil, &common.ApiError{
				InternalError: err,
				StatusCode:    500,
				ErrorCode:     common.ErrorGettingConnection,
				RequestId:     requestId,
			}
		}

		// Resolves connection for each user connection
		for index := range channel.Val() {
			connection := connectionCmds[index]

			if connection.Err() != nil {
				return nil, &common.ApiError{
					InternalError: connection.Err(),
					StatusCode:    500,
					ErrorCode:     common.ErrorGettingConnection,
					RequestId:     requestId,
				}
			}

			if len(connection.Val()) == 0 {
				// Connection doesn't exist
				continue
			}

			workerId, hasWorkerId := connection.Val()["workerId"]

			if !hasWorkerId {
				// Is missing worker ID, ignoring
				continue
			}

			workerIds = append(workerIds, workerId)
		}

	} else if options.User != "" {
		// Get all connections for a user (optionally filtered by session)
		user := redisClient.SMembers("user:" + options.User)

		if user.Err() != nil {
			return nil, &common.ApiError{
				InternalError: user.Err(),
				StatusCode:    500,
				ErrorCode:     common.ErrorGettingUser,
				RequestId:     requestId,
			}
		}

		if len(user.Val()) == 0 {
			// User doesn't exist
			return []string{}, nil
		}

		var connectionCmds = make([]*redis.StringStringMapCmd, len(user.Val()))
		_, err := redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
			for index, connId := range user.Val() {
				connectionCmds[index] = pipeliner.HGetAll("conn:" + connId)
			}

			return nil
		})

		if err != nil {
			return nil, &common.ApiError{
				InternalError: err,
				StatusCode:    500,
				ErrorCode:     common.ErrorGettingConnection,
				RequestId:     requestId,
			}
		}

		// Resolves connection for each user connection
		for index := range user.Val() {
			connection := connectionCmds[index]

			if connection.Err() != nil {
				return nil, &common.ApiError{
					InternalError: connection.Err(),
					StatusCode:    500,
					ErrorCode:     common.ErrorGettingConnection,
					RequestId:     requestId,
				}
			}

			if len(connection.Val()) == 0 {
				// Connection doesn't exist
				continue
			}

			workerId, hasWorkerId := connection.Val()["workerId"]

			if !hasWorkerId {
				// Is missing worker ID, ignoring
				continue
			}

			// Target specific session(s) for user if set
			if options.Session != "" && connection.Val()["session"] != options.Session {
				continue
			}

			workerIds = append(workerIds, workerId)
		}

	} else {
		// No targeting options where provided
		return nil, &common.ApiError{
			StatusCode: 400,
			ErrorCode:  common.ErrorTarget,
			RequestId:  requestId,
		}
	}

	return common.RemoveEmpty(common.UniqueString(workerIds)), nil
}
