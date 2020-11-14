package main

import (
	"context"
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var redisClient *redis.Client

var options *common.DSockOptions
var logger *zap.Logger

func init() {
	var err error

	options, err = common.GetOptions(false)

	if err != nil {
		println("Could not get options. Make sure your config is valid!")
		panic(err)
	}

	if options.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		println("Could not create logger")
		panic(err)
	}
}

func main() {
	logger.Info("Starting dSock API",
		zap.String("version", common.DSockVersion),
		zap.Int("port", options.Port),
		zap.String("DEPRECATED.address", options.Address),
	)

	// Setup application
	redisClient = redis.NewClient(options.RedisOptions)

	_, err := redisClient.Ping().Result()
	if err != nil {
		logger.Error("Could not connect to Redis (ping)",
			zap.Error(err),
		)
	}

	if options.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := common.NewGinEngine(logger, options)
	router.Use(common.RequestIdMiddleware)
	router.Use(common.TokenMiddleware(options.Token))

	router.Any(common.PathPing, common.PingHandler)
	router.POST(common.PathSend, sendHandler)
	router.POST(common.PathDisconnect, disconnectHandler)
	router.POST(common.PathClaim, createClaimHandler)
	router.GET(common.PathInfo, infoHandler)
	router.POST(common.PathChannelSubscribe, getChannelHandler(protos.ChannelAction_SUBSCRIBE))
	router.POST(common.PathChannelUnsubscribe, getChannelHandler(protos.ChannelAction_UNSUBSCRIBE))

	// Start HTTP server
	srv := &http.Server{
		Addr:    options.Address,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed listening",
				zap.Error(err),
			)
			options.QuitChannel <- struct{}{}
		}
	}()

	logger.Info("Listening",
		zap.String("address", options.Address),
	)

	signalQuit := make(chan os.Signal, 1)

	// Listen for signal or message in quit channel
	signal.Notify(signalQuit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-options.QuitChannel:
	case <-signalQuit:
	}

	signalQuit = nil

	// Server shutdown
	logger.Info("Shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Error during server shutdown",
			zap.Error(err),
		)
	}

	logger.Info("Stopped")
	_ = logger.Sync()
}
