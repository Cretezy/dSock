package main

import (
	"context"
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true,
}

var workerId = uuid.New().String()

var users = usersState{
	state: make(map[string][]string),
}
var channels = channelsState{
	state: make(map[string][]string),
}
var connections = connectionsState{
	state: make(map[string]*SockConnection),
}

var options *common.DSockOptions
var logger *zap.Logger
var redisClient *redis.Client

func init() {
	var err error

	options, err = common.GetOptions()

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
	logger.Info("Starting dSock worker",
		zap.String("version", common.DSockVersion),
	)
	// Setup application
	redisClient = redis.NewClient(options.RedisOptions)

	_, err := redisClient.Ping().Result()
	if err != nil {
		logger.Error("Could not connect to Redis (ping)",
			zap.Error(err),
		)
	}

	messageSubscription := redisClient.Subscribe(workerId)
	channelSubscription := redisClient.Subscribe(workerId + ":channel")

	if options.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := common.NewGinEngine(logger, options)
	router.Use(common.RequestIdMiddleware)

	router.GET(common.PathConnect, connectHandler)

	// Start HTTP server
	srv := &http.Server{
		Addr:    options.Address,
		Handler: router,
	}

	signalQuit := make(chan os.Signal, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed listening",
				zap.Error(err),
			)
			options.QuitChannel <- struct{}{}
		}
	}()

	// Loop receiving messages from Redis
	go func() {
		for {
			redisMessage, err := messageSubscription.ReceiveMessage()
			if err != nil {
				// TODO: Possibly add better handling
				logger.Error("Error receiving message from Redis",
					zap.Error(err),
				)
				break
			}

			var message protos.Message

			err = proto.Unmarshal([]byte(redisMessage.Payload), &message)

			if err != nil {
				// Couldn't parse message
				logger.Error("Invalid message received from Redis",
					zap.Error(err),
				)
				continue
			}

			go handleSend(&message)

			if signalQuit == nil {
				break
			}
		}
	}()

	// Loop receiving channel actions from Redis
	go func() {
		for {
			redisMessage, err := channelSubscription.ReceiveMessage()
			if err != nil {
				// TODO: Possibly add better handling
				logger.Error("Error receiving message from Redis",
					zap.Error(err),
				)
				break
			}

			var channelAction protos.ChannelAction

			err = proto.Unmarshal([]byte(redisMessage.Payload), &channelAction)

			if err != nil {
				// Couldn't parse channel action
				logger.Error("Invalid message received from Redis",
					zap.Error(err),
				)
				continue
			}

			go handleChannel(&channelAction)

			if signalQuit == nil {
				break
			}
		}
	}()

	logger.Info("Listening",
		zap.String("address", options.Address),
	)

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

	// Cleanup
	_ = messageSubscription.Close()
	_ = channelSubscription.Close()

	// Disconnect all connections
	for _, connection := range connections.state {
		connection.CloseChannel <- struct{}{}
	}

	// Allow time to disconnect & clear from Redis
	time.Sleep(time.Second)

	logger.Info("Stopped")
	_ = logger.Sync()
}
