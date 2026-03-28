package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bicycle-network/rental-service/internal/config"
	"github.com/bicycle-network/rental-service/internal/handler"
	"github.com/bicycle-network/rental-service/internal/messaging"
	"github.com/bicycle-network/rental-service/internal/repository"
	"github.com/bicycle-network/rental-service/internal/router"
	"github.com/bicycle-network/rental-service/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	// --- Database ---
	db, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected")

	if cfg.RunMigrations {
		if err := runMigrations(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations applied")
	}

	// --- RabbitMQ ---
	mqManager := messaging.NewRabbitMQManager(cfg.RabbitMQURL)
	if err := mqManager.Connect(); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	// --- Repositories ---
	bikeRepo := repository.NewBikeRepository(db)
	rentalRepo := repository.NewRentalRepository(db)

	// --- Consumer ---
	consumer := messaging.NewConsumer(mqManager.Channel(), bikeRepo)
	if err := consumer.Setup(); err != nil {
		log.Fatalf("Failed to setup consumer: %v", err)
	}

	// --- Service ---
	rentalSvc := service.NewRentalService(rentalRepo, bikeRepo)

	// --- Handlers ---
	rentalHandler := handler.NewRentalHandler(rentalSvc)
	healthHandler := handler.NewHealthHandler(db, mqManager.IsConnected)

	// --- Router ---
	r := router.Setup(rentalHandler, healthHandler, cfg.JWTSecret)

	// --- Start consumer goroutine ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// --- Start HTTP server ---
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("rental-service started on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// --- Graceful shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	mqManager.Close()

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	log.Println("Shutdown complete")
}

func connectDB(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, _ := db.DB()
			if sqlDB.Ping() == nil {
				return db, nil
			}
		}
		log.Printf("Database connection attempt %d/5 failed: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("failed to connect to database after 5 attempts: %w", err)
}

func runMigrations(db *gorm.DB) error {
	migration, err := os.ReadFile("migrations/000001_init.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}
	return db.Exec(string(migration)).Error
}
