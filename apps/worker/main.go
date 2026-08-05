package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"decree/internal/approval"
	"decree/internal/delivery"
	"decree/internal/platform/config"
	"decree/internal/platform/database"
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
	approvalService := approval.NewService(dbPool, nil, logger)

	logger.Info("worker started", "queues", cfg.WorkerQueues)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	processDue(ctx, logger, approvalService, deliveryService)

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker stopped")
			return
		case <-ticker.C:
			processDue(ctx, logger, approvalService, deliveryService)
		}
	}
}

func processDue(ctx context.Context, logger *slog.Logger, approvalService *approval.Service, deliveryService *delivery.Service) {
	if _, err := approvalService.ExpireDueApprovalRequests(ctx, 0); err != nil {
		logger.ErrorContext(ctx, "process due approval expiries", "error", err)
	}

	if err := deliveryService.ProcessDue(ctx); err != nil {
		logger.ErrorContext(ctx, "process due deliveries", "error", err)
	}
}
