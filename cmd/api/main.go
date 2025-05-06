package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thanhnguyen/product-api/internal/business/usecase"
	"github.com/thanhnguyen/product-api/internal/config"
	"github.com/thanhnguyen/product-api/internal/storage/cache"
	"github.com/thanhnguyen/product-api/internal/storage/postgres"
	transportHttp "github.com/thanhnguyen/product-api/internal/transport/http"
	"github.com/thanhnguyen/product-api/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewLogger(cfg.Logger.Level, cfg.Logger.Format, cfg.Logger.OutputPath)
	log.Info("Starting application")

	// Connect to database
	db, err := postgres.NewPostgresDB(cfg.GetDatabaseURL(),
		cfg.Database.MaxConns,
		cfg.Database.MinConns,
		cfg.Database.Timeout)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()
	log.Info("Connected to database")

	// Create repositories
	productRepo := postgres.NewProductRepository(db, log)
	categoryRepo := postgres.NewCategoryRepository(db, log)

	// Create caches
	statsCache := cache.NewStatsCache(log)

	// Create use cases
	productUseCase := usecase.NewProductUseCase(productRepo, categoryRepo, log, 5*time.Minute)
	statsUseCase := usecase.NewStatsUseCase(productRepo, categoryRepo, nil, nil, statsCache, log, 15*time.Minute)

	// Create HTTP server
	server := transportHttp.NewServer(cfg, log, productUseCase, statsUseCase)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.WithError(err).Fatal("Failed to start server")
		}
	}()
	log.Infof("Server started on port %d", cfg.Server.Port)

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server forced to shutdown")
	}

	log.Info("Server exiting")
}
