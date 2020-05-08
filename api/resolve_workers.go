package main

import (
	"github.com/Cretezy/dSock/common"
	"sync"
)

/// Resolves the workers holding the connection
func resolveWorkers(options common.ResolveOptions) ([]string, *common.ApiError) {
	workerIds := make([]string, 0)

	var workersLock sync.Mutex
	var apiError *common.ApiError

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
			}
		}

		if len(channel.Val()) == 0 {
			// User doesn't exist
			return []string{}, nil
		}

		var channelWaitGroup sync.WaitGroup
		channelWaitGroup.Add(len(channel.Val()))

		// Resolves connection for each user connection
		for _, connId := range channel.Val() {
			connId := connId
			go func() {
				defer channelWaitGroup.Done()

				connection := redisClient.HGetAll("conn:" + connId)

				// Stop if one of the Redis commands failed
				if apiError != nil {
					return
				}

				if connection.Err() != nil {
					apiError = &common.ApiError{
						InternalError: connection.Err(),
						StatusCode:    500,
						ErrorCode:     common.ErrorGettingConnection,
					}
					return
				}

				if len(connection.Val()) == 0 {
					// Connection doesn't exist
					return
				}

				workerId, hasWorkerId := connection.Val()["workerId"]

				if !hasWorkerId {
					// Is missing worker ID, ignoring
					return
				}

				workersLock.Lock()
				workerIds = append(workerIds, workerId)
				workersLock.Unlock()
			}()
		}

		channelWaitGroup.Wait()
	} else if options.User != "" {
		// Get all connections for a user (optionally filtered by session)
		user := redisClient.SMembers("user:" + options.User)

		if user.Err() != nil {
			return nil, &common.ApiError{
				InternalError: user.Err(),
				StatusCode:    500,
				ErrorCode:     common.ErrorGettingUser,
			}
		}

		if len(user.Val()) == 0 {
			// User doesn't exist
			return []string{}, nil
		}

		var usersWaitGroup sync.WaitGroup
		usersWaitGroup.Add(len(user.Val()))

		// Resolves connection for each user connection
		for _, connId := range user.Val() {
			connId := connId
			go func() {
				defer usersWaitGroup.Done()

				connection := redisClient.HGetAll("conn:" + connId)

				// Stop if one of the Redis commands failed
				if apiError != nil {
					return
				}

				if connection.Err() != nil {
					apiError = &common.ApiError{
						InternalError: connection.Err(),
						StatusCode:    500,
						ErrorCode:     common.ErrorGettingConnection,
					}
					return
				}

				if len(connection.Val()) == 0 {
					// Connection doesn't exist
					return
				}

				workerId, hasWorkerId := connection.Val()["workerId"]

				if !hasWorkerId {
					// Is missing worker ID, ignoring
					return
				}

				// Target specific session(s) for user if set
				if options.Session != "" && connection.Val()["session"] != options.Session {
					return
				}

				workersLock.Lock()
				workerIds = append(workerIds, workerId)
				workersLock.Unlock()
			}()
		}

		usersWaitGroup.Wait()
	} else {
		// No targeting options where provided
		return nil, &common.ApiError{
			StatusCode: 400,
			ErrorCode:  common.ErrorTarget,
		}
	}

	if apiError != nil {
		return nil, apiError
	}

	return common.RemoveEmpty(common.UniqueString(workerIds)), nil
}
