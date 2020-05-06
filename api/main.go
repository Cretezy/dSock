package main

import (
	"context"
	"github.com/Cretezy/dSock/common"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var redisClient *redis.Client

var options common.DSockOptions

func init() {
	options = common.GetOptions()
}

func main() {
	log.Printf("Starting dSock API %s\n", common.DSockVersion)

	// Setup application
	redisClient = redis.NewClient(options.RedisOptions)

	_, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}

	if options.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(common.TokenMiddleware(options.Token))

	router.POST(common.PathSend, sendHandler)
	router.POST(common.PathDisconnect, disconnectHandler)
	router.POST(common.PathClaim, createClaimHandler)
	router.GET(common.PathInfo, infoHandler)

	// Start HTTP server
	srv := &http.Server{
		Addr:    options.Address,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed listening: %s\n", err)
		}
	}()

	log.Printf("Starting on: %s\n", options.Address)

	signalQuit := make(chan os.Signal, 1)

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
		log.Fatalf("Error during server shutdown: %s\n", err)
	}

	log.Println("Server stopped")
}
