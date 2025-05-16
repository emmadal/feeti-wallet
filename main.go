package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/emmadal/feeti-wallet/controllers"
	"github.com/emmadal/feeti-wallet/helpers"
	"github.com/emmadal/feeti-wallet/middleware"
	"github.com/emmadal/feeti-wallet/models"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from the.env file if it exists,
	// This is now optional since we're using Docker env variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":4000"
	}

	// Initialize Gin server
	server := gin.Default()

	// middleware
	server.Use(
		cors.New(
			cors.Config{
				AllowMethods:     []string{"GET", "POST"},
				AllowOrigins:     []string{"*"},
				AllowFiles:       false,
				AllowWildcard:    false,
				AllowCredentials: true,
			},
		),
	)
	server.Use(middleware.Helmet())
	server.Use(gzip.Gzip(gzip.BestCompression))
	server.Use(middleware.Timeout(5 * time.Second))
	server.Use(middleware.Recover())

	// Set api version group
	v1 := server.Group("/v1/api")

	// initialize server
	s := &http.Server{
		Handler:        server,
		Addr:           port,
		WriteTimeout:   10 * time.Second,
		ReadTimeout:    10 * time.Second,
		IdleTimeout:    20 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// v1 routes
	v1.POST("/lock", controllers.LockWalletByUser)
	v1.GET("/balance/:userID", controllers.GetBalanceByUser)
	v1.POST("/deposit", controllers.TopupWallet)
	v1.POST("/withdraw", controllers.WithdrawWallet)
	v1.POST("/unlock", controllers.UnLockWalletByUser)
	v1.GET("/health", controllers.HealthCheck)

	// Subscription is now handled inside NatsConnect
	if err := helpers.NatsConnect(); err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	// start server
	go func() {
		// Database connection
		models.DBConnect()
		_, err := fmt.Fprintf(os.Stdout, "Server started on port %s\n", port)
		if err != nil {
			log.Fatalln("Error writing to stdout")
		}
		// service connections
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for the interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Drain NATS connection on shutdown
	if err := helpers.DrainNatsConnection(); err != nil {
		log.Printf("Error draining NATS connection: %v\n", err)
	} else {
		log.Println("NATS connection drained successfully")
	}

	if err := s.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	models.DB.Close()
	<-ctx.Done()
	log.Println("Server exiting")
}
