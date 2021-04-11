package main

import (
	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
	"time"
)

func RefreshTtls() {
	nextTime := time.Now()

	for {
		nextTime = nextTime.Add(options.TtlDuration)
		time.Sleep(time.Until(nextTime))

		_, err := redisClient.Pipelined(func(pipeliner redis.Pipeliner) error {
			redisWorker := map[string]interface{}{
				"lastPing": time.Now().Format(time.RFC3339),
			}

			pipeliner.HSet("worker:"+workerId, redisWorker)
			pipeliner.Expire("worker:"+workerId, options.TtlDuration*2)

			for connId := range connections.state {
				pipeliner.Expire("conn:"+connId, options.TtlDuration*2)
			}

			return nil
		})

		if err != nil {
			logger.Error("Could not refresh TTLs",
				zap.Error(err),
				zap.String("workerId", workerId),
			)

			continue
		}

		logger.Info("Refreshed TTLs",
			zap.String("workerId", workerId),
		)
	}
}
