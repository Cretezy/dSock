package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
	"time"
)

func RefreshTtls() {
	_, err := redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
		redisWorker := map[string]interface{}{
			"lastPing": time.Now().Format(time.RFC3339),
		}

		pipeliner.HSet("worker:"+workerId, redisWorker)
		pipeliner.Expire("worker:"+workerId, options.TtlDuration+common.TtlBuffer)

		for connId := range connections.state {
			pipeliner.Expire("conn:"+connId, options.TtlDuration+common.TtlBuffer)
		}

		return nil
	})

	if err != nil {
		logger.Error("Could not refresh TTLs",
			zap.Error(err),
			zap.String("workerId", workerId),
		)
	}
}
