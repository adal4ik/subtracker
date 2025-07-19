package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"subtracker/internal/config"
	"subtracker/internal/handler"
	"subtracker/internal/repository"
	"subtracker/internal/service"
	"subtracker/pkg/loadenv"
	"subtracker/pkg/logger"

	"go.uber.org/zap"
)

// @title           Subscription Tracker API
// @version         1.0
// @description     This is a service for aggregating user online subscriptions, part of a test task for Effective Mobile.

// @contact.name   adal4ik
// @contact.url    https://github.com/adal4ik/subtracker

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /
// @schemes   http
func main() {
	ctx := context.Background()
	loadenv.LoadEnvFile(".env")
	logger := logger.New(os.Getenv("APP_ENV"))
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Error syncing logger: %v\n", err)
		}
	}()
	logger.Info("Starting Subtracker application", zap.String("environment", os.Getenv("APP_ENV")))
	// Initialize configuration
	cfg := config.LoadConfig()
	logger.Info("Configuration loaded", zap.Any("config", cfg))
	// Connect to the database
	db, err := repository.ConnectDB(ctx, cfg.Postgres, logger)
	if err != nil {
		logger.Error("Failed to connect to the database", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("Connected to the database successfully", zap.String("dsn", cfg.Postgres.PostgresDSN))

	// Initialize the all components
	repo := repository.NewRepository(db, logger)
	service := service.NewService(repo, logger)
	handlers := handler.NewHandlers(service, logger)
	logger.Info("All components initialized successfully")

	mux := handler.Router(*handlers)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		log.Println("Server is running on port: http://localhost" + httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("ListenAndServe error", zap.Error(err))
		}
	}()
	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	<-ctx.Done()
	logger.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("HTTP server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}
