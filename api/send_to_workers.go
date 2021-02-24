package main

import (
	"bytes"
	"github.com/Cretezy/dSock/common"
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
			ErrorCode:  common.ErrorMarshallingMessage,
			StatusCode: 500,
			RequestId:  requestId,
		}
	}

	var workersWaitGroup sync.WaitGroup
	workersWaitGroup.Add(len(workerIds))

	errs := make([]error, 0)
	errsLock := sync.Mutex{}

	addError := func(err error) {
		errsLock.Lock()
		defer errsLock.Lock()

		errs = append(errs, err)
	}

	for _, workerId := range workerIds {
		workerId := workerId
		go func() {
			defer workersWaitGroup.Done()

			if options.MessagingMethod == common.MessageMethodRedis {
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

				redisClient.Publish(redisChannel, rawMessage)
			} else {
				worker := redisClient.HGetAll("worker:" + workerId)
				if worker.Err() != nil {
					logger.Error("Could not find worker in Redis",
						zap.String("requestId", requestId),
						zap.String("workerId", workerId),
					)
					addError(&common.ApiError{
						InternalError: worker.Err(),
						StatusCode:    500,
						ErrorCode:     common.ErrorGettingWorker,
					})
					return
				}

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
			}
		}()
	}

	workersWaitGroup.Wait()

	if len(errs) > 0 {
		return &common.ApiError{
			ErrorCode:  common.ErrorDeliveringMessage,
			StatusCode: 500,
		}
	}

	return nil
}
