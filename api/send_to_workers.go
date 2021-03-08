package main

import (
	"bytes"
	"errors"
	"github.com/Cretezy/dSock/common"
	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"net/http"
	"sync"
	"time"
)

const (
	MessageMessageType = "message"
	ChannelMessageType = "channel"
)

func sendToWorkers(workerIds []string, message proto.Message, messageType string, requestId string) *common.ApiError {
	rawMessage, err := proto.Marshal(message)

	if err != nil {
		return &common.ApiError{
			InternalError: err,
			ErrorCode:     common.ErrorMarshallingMessage,
			StatusCode:    500,
			RequestId:     requestId,
		}
	}

	errs := make([]error, 0)
	errsLock := sync.Mutex{}

	addError := func(err error) {
		errsLock.Lock()
		defer errsLock.Lock()

		errs = append(errs, err)
	}

	if options.MessagingMethod == common.MessageMethodRedis {
		_, err := redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
			for _, workerId := range workerIds {
				redisChannel := workerId
				if messageType == ChannelMessageType {
					redisChannel = redisChannel + ":channel"
				}

				logger.Info("Publishing to worker",
					zap.String("requestId", requestId),
					zap.String("workerId", workerId),
					zap.String("messageType", messageType),
					zap.String("redisChannel", redisChannel),
				)

				pipeliner.Publish(redisChannel, rawMessage)
			}

			return nil
		})

		if err != nil {
			return &common.ApiError{
				InternalError: err,
				ErrorCode:     common.ErrorDeliveringMessage,
				StatusCode:    500,
				RequestId:     requestId,
			}
		}

	} else {
		var workerCmds = make([]*redis.StringStringMapCmd, len(workerIds))
		_, err := redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
			for index, workerId := range workerIds {
				workerCmds[index] = pipeliner.HGetAll("worker:" + workerId)
			}

			return nil
		})

		if err != nil {
			return &common.ApiError{
				InternalError: err,
				ErrorCode:     common.ErrorDeliveringMessage,
				StatusCode:    500,
				RequestId:     requestId,
			}
		}

		var workersWaitGroup sync.WaitGroup
		workersWaitGroup.Add(len(workerIds))

		for index, workerId := range workerIds {
			index := index
			workerId := workerId

			go func() {
				defer workersWaitGroup.Done()

				worker := workerCmds[index]

				if len(worker.Val()) == 0 {
					logger.Error("Found empty worker in Redis",
						zap.String("requestId", requestId),
						zap.String("workerId", workerId),
					)
					return
				}

				ip := worker.Val()["ip"]

				if ip == "" {
					logger.Error("Found worker with no IP in Redis (is the worker not configured for direct access?)",
						zap.String("requestId", requestId),
						zap.String("workerId", workerId),
					)
					return
				}

				path := common.PathReceiveMessage
				if messageType == ChannelMessageType {
					path = common.PathReceiveChannelMessage
				}

				url := "http://" + ip + path

				logger.Info("Starting request to worker",
					zap.String("requestId", requestId),
					zap.String("workerId", workerId),
					zap.String("messageType", messageType),
					zap.String("url", url),
				)

				// Non-standardized Content-Type used here.
				// Future: Add HTTPS options? Maybe store full direct path in Redis instead of "ip",
				// which here is hostname + port.
				// Also, no error can be returned on the worker side, so only 200 can be handled.
				req, err := http.NewRequest("POST", url, bytes.NewReader(rawMessage))
				if err != nil {
					logger.Error("Could create request",
						zap.String("requestId", requestId),
						zap.String("workerId", workerId),
						zap.String("url", url),
						zap.String("messageType", messageType),
						zap.Error(err),
					)
					addError(&common.ApiError{
						InternalError: worker.Err(),
						StatusCode:    500,
						ErrorCode:     common.ErrorReachingWorker,
					})
					return
				}

				req.Header.Set("Content-Type", common.ProtobufContentType)
				req.Header.Set("X-Request-ID", requestId)

				beforeRequestTime := time.Now()
				resp, err := http.DefaultClient.Do(req)
				requestTime := time.Now().Sub(beforeRequestTime)

				if err != nil {
					logger.Error("Could not reach worker",
						zap.String("requestId", requestId),
						zap.String("workerId", workerId),
						zap.String("url", url),
						zap.String("messageType", messageType),
						zap.Duration("requestTime", requestTime),
						zap.Error(err),
					)
					addError(&common.ApiError{
						InternalError: worker.Err(),
						StatusCode:    500,
						ErrorCode:     common.ErrorReachingWorker,
					})
					return
				}

				if resp.StatusCode != 200 {
					logger.Error("Worker could not handle request",
						zap.String("requestId", requestId),
						zap.String("workerId", workerId),
						zap.String("url", url),
						zap.String("messageType", messageType),
						zap.Int("statusCode", resp.StatusCode),
						zap.String("status", resp.Status),
						zap.Duration("requestTime", requestTime),
					)
					addError(&common.ApiError{
						InternalError: worker.Err(),
						StatusCode:    500,
						ErrorCode:     common.ErrorReachingWorker,
					})
				}

				logger.Info("Finished request to worker",
					zap.String("requestId", requestId),
					zap.String("workerId", workerId),
					zap.String("url", url),
					zap.String("messageType", messageType),
					zap.Int("statusCode", resp.StatusCode),
					zap.String("status", resp.Status),
					zap.Duration("requestTime", requestTime),
				)
			}()
		}

		workersWaitGroup.Wait()

		if len(errs) > 0 {
			errorMessage := ""
			for index, err := range errs {
				errorMessage = errorMessage + err.Error()
				if index+1 != len(errs) {
					errorMessage = errorMessage + ", "
				}
			}

			return &common.ApiError{
				InternalError: errors.New(errorMessage),
				ErrorCode:     common.ErrorDeliveringMessage,
				StatusCode:    500,
				RequestId:     requestId,
			}
		}
	}

	return nil
}
