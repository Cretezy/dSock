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
	"strconv"
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

	options, err = common.GetOptions(true)

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
		zap.String("workerId", workerId),
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

	router.Any(common.PathPing, common.PingHandler)
	router.GET(common.PathConnect, connectHandler)

	// Start HTTP server
	srv := &http.Server{
		Addr:    options.Address,
		Handler: router,
	}

	signalQuit := make(chan os.Signal, 1)

	// Register worker in Redis
	redisWorker := map[string]interface{}{
		"lastPing": time.Now().Format(time.RFC3339),
	}
	if options.MessagingMethod == common.MessageMethodDirect {
		redisWorker["ip"] = options.DirectHostname + ":" + strconv.Itoa(options.DirectPort)
	}

	redisClient.HSet("worker:"+workerId, redisWorker)
	redisClient.Expire("worker:"+workerId, options.TtlDuration*2)

	closeMessaging := func() {}

	go RefreshTtls()

	if options.MessagingMethod == common.MessageMethodRedis {
		logger.Info("Starting Redis messaging method",
			zap.String("workerId", workerId),
		)

		// Loop receiving messages from Redis
		messageSubscription := redisClient.Subscribe(workerId)
		go func() {
			for {
				redisMessage, err := messageSubscription.ReceiveMessage()
				if err != nil {
					// TODO: Possibly add better handling
					logger.Error("Error receiving message from Redis",
						zap.Error(err),
						zap.String("workerId", workerId),
					)
					break
				}

				go func() {
					var message protos.Message

					err = proto.Unmarshal([]byte(redisMessage.Payload), &message)

					if err != nil {
						// Couldn't parse message
						logger.Error("Invalid message received from Redis",
							zap.Error(err),
							zap.String("workerId", workerId),
						)
						return
					}

					handleSend(&message)
				}()

				if signalQuit == nil {
					break
				}
			}
		}()

		// Loop receiving channel actions from Redis
		channelSubscription := redisClient.Subscribe(workerId + ":channel")
		go func() {
			for {
				redisMessage, err := channelSubscription.ReceiveMessage()
				if err != nil {
					// TODO: Possibly add better handling
					logger.Error("Error receiving message from Redis",
						zap.Error(err),
						zap.String("workerId", workerId),
					)
					break
				}

				go func() {
					var channelAction protos.ChannelAction

					err = proto.Unmarshal([]byte(redisMessage.Payload), &channelAction)

					if err != nil {
						// Couldn't parse channel action
						logger.Error("Invalid message received from Redis",
							zap.Error(err),
							zap.String("workerId", workerId),
						)
						return
					}

					handleChannel(&channelAction)
				}()

				if signalQuit == nil {
					break
				}
			}
		}()

		closeMessaging = func() {
			_ = messageSubscription.Close()
			_ = channelSubscription.Close()
		}
	} else {
		logger.Info("Starting direct messaging method",
			zap.String("workerId", workerId),
			zap.String("directHostname", options.DirectHostname),
			zap.Int("directPort", options.DirectPort),
		)

		router.POST(common.PathReceiveMessage, sendMessageHandler)
		router.POST(common.PathReceiveChannelMessage, channelMessageHandler)
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed listening",
				zap.Error(err),
				zap.String("workerId", workerId),
			)
			options.QuitChannel <- struct{}{}
		}
	}()

	logger.Info("Listening",
		zap.String("address", options.Address),
		zap.String("workerId", workerId),
	)

	// Listen for signal or message in quit channel
	signal.Notify(signalQuit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-options.QuitChannel:
	case <-signalQuit:
	}

	// Server shutdown
	logger.Info("Shutting down",
		zap.String("workerId", workerId),
	)

	// Cleanup
	closeMessaging()
	redisClient.Del("worker:" + workerId)

	// Disconnect all connections
	for _, connection := range connections.state {
		connection.CloseChannel <- struct{}{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Error during server shutdown",
			zap.Error(err),
			zap.String("workerId", workerId),
		)
	}

	// Allow time to disconnect & clear from Redis
	time.Sleep(time.Second)

	logger.Info("Stopped",
		zap.String("workerId", workerId),
	)
	_ = logger.Sync()
}
