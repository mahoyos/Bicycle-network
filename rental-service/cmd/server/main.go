package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/config"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/database"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/dependencies"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/handlers"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/messaging"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/repositories"
	"github.com/mahoyos/Bicycle-network/rental-service/internal/services"
)

var mqClient *messaging.RabbitMQClient

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	// Initialize database
	if err := database.InitDB(cfg.DatabaseURL, cfg.RunMigrations); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Initialize RabbitMQ
	var err error
	mqClient, err = messaging.NewRabbitMQClient(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("WARNING: RabbitMQ not available: %v", err)
	} else {
		defer mqClient.Close()
	}

	// Build dependency chain
	repo := repositories.NewRentalsRepository(database.GetDB())
	var publisher services.EventPublisher
	if mqClient != nil {
		publisher = mqClient
	}
	svc := services.NewRentalsService(repo, publisher)
	handler := handlers.NewRentalsHandler(svc)

	// Start RabbitMQ consumer for bike lifecycle events (DELETED)
	if mqClient != nil {
		err := mqClient.StartConsumer(func(event messaging.BicycleEvent) {
			bikeID, parseErr := uuid.Parse(event.BikeID)
			if parseErr != nil {
				log.Printf("Invalid bike_id in event: %s", event.BikeID)
				return
			}
			svc.HandleBicycleDeleted(bikeID)
		})
		if err != nil {
			log.Printf("WARNING: Failed to start RabbitMQ consumer: %v", err)
		}
	}

	// Setup Gin router
	router := gin.Default()

	// CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health & readiness
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "rental-service"})
	})
	router.GET("/ready", func(c *gin.Context) {
		dbOk := database.CheckDB()
		mqOk := mqClient != nil && mqClient.Check()
		if !dbOk || !mqOk {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "not ready",
				"database": dbOk,
				"rabbitmq": mqOk,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready", "database": true, "rabbitmq": true})
	})

	// Protected rental routes
	rentals := router.Group("/rentals")
	rentals.Use(dependencies.AuthMiddleware(cfg.JWTSecretKey, cfg.DisableAuth))
	handler.RegisterRoutes(rentals)

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		log.Printf("Rental service starting on port %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
