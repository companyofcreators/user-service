package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/companyofcreators/user-service/internal/app"
	"github.com/companyofcreators/user-service/internal/config"
	"github.com/companyofcreators/user-service/internal/infrastructure/db"
	httphandler "github.com/companyofcreators/user-service/internal/interfaces/http"
	"github.com/companyofcreators/user-service/internal/pkg"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize structured logger
	logger := pkg.NewLogger(cfg.LogLevel)
	logger.Info("starting user-service",
		slog.String("http_address", cfg.HTTPAddress),
		slog.String("log_level", cfg.LogLevel),
	)

	// Connect to PostgreSQL
	database, err := db.NewPostgresDB(cfg.DBDSN, logger)
	if err != nil {
		logger.Error("failed to connect to postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("failed to close database connection", slog.String("error", err.Error()))
		}
	}()

	// Build dependency container
	kafkaBrokers := cfg.KafkaBrokersList()
	container := app.NewContainer(database, logger, kafkaBrokers, cfg.OrderServiceURL)

	// Create HTTP handler and router
	handler := httphandler.NewUserHandler(
		container.GetProfileUseCase,
		container.UpdateProfileUseCase,
		container.GetMasterProfileUseCase,
		container.UpdateMasterProfileUseCase,
		container.SwitchRoleUseCase,
		container.UserRoleRepo,
		logger,
	)

	router := httphandler.NewRouter(handler)

	// Start Kafka consumers in goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := container.KafkaConsumer.Start(ctx); err != nil {
		logger.Error("failed to start kafka consumers", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Configure and start HTTP server
	server := &http.Server{
		Addr:         cfg.HTTPAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("http server starting", slog.String("address", cfg.HTTPAddress))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("received shutdown signal", slog.String("signal", sig.String()))

	// Shutdown Kafka consumers
	container.KafkaConsumer.Shutdown()
	cancel()

	// Shutdown HTTP server with graceful timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server forced to shutdown", slog.String("error", err.Error()))
	}

	logger.Info("user-service stopped")
}
