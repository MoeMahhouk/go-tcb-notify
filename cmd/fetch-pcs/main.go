package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/services/pcs"
	"github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
)

func main() {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	setupShutdownHandler(cancel)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logging
	setupLogging(cfg.LogLevel)

	// Connect to database
	db, err := clickhouse.Open(ctx, &cfg.ClickHouse)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer db.Close()

	// Create storage implementations
	tcbStore := clickhouse.NewTCBStore(db)
	quoteStore := clickhouse.NewQuoteStore(db)
	alertStore := clickhouse.NewAlertStore(db)

	// Create and run the PCS fetcher service with dependency injection
	fetcher := pcs.NewFetcher(
		tcbStore,
		quoteStore,
		alertStore,
		&cfg.PCS,
	)
	if err := fetcher.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("PCS fetcher failed: %v", err)
	}
}

func setupShutdownHandler(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logrus.Info("Shutting down fetch-pcs service...")
		cancel()
	}()
}

func setupLogging(logLevel string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}
