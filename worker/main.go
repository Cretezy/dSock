package main

import (
	"context"
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

var connections = make(map[string]*SockConnection)
var connectionsLock sync.Mutex

var users = make(map[string][]string)
var usersLock sync.Mutex

var channels = make(map[string][]string)
var channelsLock sync.Mutex

var redisClient *redis.Client
var options common.DSockOptions

func init() {
	options = common.GetOptions()
}

func main() {
	log.Printf("Starting dSock worker %s\n", common.DSockVersion)

	// Setup application
	redisClient = redis.NewClient(options.RedisOptions)

	_, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}

	messageSubscription := redisClient.Subscribe(workerId)
	channelSubscription := redisClient.Subscribe(workerId + ":channel")

	if options.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.GET(common.PathConnect, connectHandler)

	// Start HTTP server
	srv := &http.Server{
		Addr:    options.Address,
		Handler: router,
	}

	signalQuit := make(chan os.Signal, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed listening: %s\n", err)
			options.QuitChannel <- struct{}{}
		}
	}()

	// Loop receiving messages from Redis
	go func() {
		for {
			redisMessage, err := messageSubscription.ReceiveMessage()
			if err != nil {
				// TODO: Possibly add better handling
				log.Printf("Error receiving from Redis: %s\n", err)
				break
			}

			var message protos.Message

			err = proto.Unmarshal([]byte(redisMessage.Payload), &message)

			if err != nil {
				// Couldn't parse message
				log.Printf("Invalid message from Redis: %s\n", err)
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
				log.Printf("Error receiving from Redis: %s\n", err)
				break
			}

			var channelAction protos.ChannelAction

			err = proto.Unmarshal([]byte(redisMessage.Payload), &channelAction)

			if err != nil {
				// Couldn't parse channel action
				log.Printf("Invalid channel action from Redis: %s\n", err)
				continue
			}

			go handleChannel(&channelAction)

			if signalQuit == nil {
				break
			}
		}
	}()

	log.Printf("Starting on: %s\n", options.Address)

	// Listen for signal or message in quit channel
	signal.Notify(signalQuit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-options.QuitChannel:
	case <-signalQuit:
	}

	signalQuit = nil

	// Server shutdown
	log.Print("Shutting server down...\n")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
	}

	// Cleanup
	_ = messageSubscription.Close()
	_ = channelSubscription.Close()

	// Disconnect all connections
	for _, connection := range connections {
		connection.CloseChannel <- struct{}{}
	}

	// Allow time to disconnect & clear from Redis
	time.Sleep(time.Second)

	log.Println("Server stopped")
}
