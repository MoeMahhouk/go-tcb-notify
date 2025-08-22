package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/services/evaluator"
	clickdb "github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
)

const serviceName = "evaluate-quotes"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logrus.Info("Shutting down evaluate-quotes service...")
		cancel()
	}()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	// Setup logging
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Connect to ClickHouse
	ch, err := clickdb.Open(ctx, &cfg.ClickHouse)
	if err != nil {
		logrus.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer ch.Close()

	logrus.WithFields(logrus.Fields{
		"service":           serviceName,
		"poll_interval":     cfg.EvaluateQuotes.PollInterval,
		"get_collateral":    cfg.EvaluateQuotes.GetCollateral,
		"check_revocations": cfg.EvaluateQuotes.CheckRevocations,
	}).Info("Starting quote evaluation service")

	// Create and run the evaluator service
	evaluatorService := evaluator.NewEvaluator(ch, &cfg.EvaluateQuotes)

	if err := evaluatorService.Run(ctx); err != nil {
		logrus.WithError(err).Error("Evaluator service failed")
	}
}
