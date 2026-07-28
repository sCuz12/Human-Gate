package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"humangate/internal/delivery"
	"humangate/internal/platform/config"
	"humangate/internal/platform/database"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.LoadWorker()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	dbPool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	deliveryService := delivery.NewService(dbPool, nil, logger, delivery.Config{
		SigningKey: cfg.DecisionSigningKey,
	})

	logger.Info("worker started", "queues", cfg.WorkerQueues)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	if err := deliveryService.ProcessDue(ctx); err != nil {
		logger.ErrorContext(ctx, "process due deliveries", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker stopped")
			return
		case <-ticker.C:
			if err := deliveryService.ProcessDue(ctx); err != nil {
				logger.ErrorContext(ctx, "process due deliveries", "error", err)
			}
		}
	}
}
