package main

import (
	"github.com/Cretezy/dSock/common"
	"google.golang.org/protobuf/proto"
	"sync"
)

func sendToWorkers(workerChannels []string, message proto.Message) *common.ApiError {
	rawMessage, err := proto.Marshal(message)

	if err != nil {
		return &common.ApiError{
			ErrorCode:  common.ErrorMarshallingMessage,
			StatusCode: 500,
		}
	}

	var workersWaitGroup sync.WaitGroup
	workersWaitGroup.Add(len(workerChannels))

	for _, workerId := range workerChannels {
		workerId := workerId
		go func() {
			defer workersWaitGroup.Done()

			redisClient.Publish(workerId, rawMessage)
		}()
	}

	workersWaitGroup.Wait()

	return nil
}
