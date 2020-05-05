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

var connections = make(map[string]SockConnection)
var users = make(map[string][]string)

var redisClient *redis.Client
var options common.DSockOptions

func init() {
	options = common.GetOptions()
}

func main() {
	log.Printf("Starting dSock worker %s\n", common.DSockVersion)

	redisClient = redis.NewClient(options.RedisOptions)

	_, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}

	subscription := redisClient.Subscribe(workerId)

	if options.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.GET(common.PathConnect, connectHandler)

	srv := &http.Server{
		Addr:    options.Address,
		Handler: router,
	}

	closed := false

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed listening: %s\n", err)
			options.QuitChannel <- struct{}{}
		}
	}()

	go func() {
		for {
			redisMessage, err := subscription.ReceiveMessage()
			if err != nil {
				// TODO: Add better handling
				println("redis receive error", err)
				break
			}

			var message protos.Message

			err = proto.Unmarshal([]byte(redisMessage.Payload), &message)

			if err != nil {
				// Couldn't parse message
				println("redis invalid message", err)
				continue
			}

			go send(&message)

			if closed {
				break
			}
		}
	}()

	log.Printf("Starting on: %s\n", options.Address)

	signalQuit := make(chan os.Signal, 1)

	signal.Notify(signalQuit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-options.QuitChannel:
	case <-signalQuit:
	}

	closed = true

	log.Print("Shutting server down...\n")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
	}

	_ = subscription.Close()

	var connectionsWaitGroup sync.WaitGroup

	connectionsWaitGroup.Add(len(connections))

	go func() {
		for _, connection := range connections {
			connection.CloseChannel <- struct{}{}
		}
	}()

	connectionsWaitGroup.Wait()

	log.Print("Server exited\n")
}
