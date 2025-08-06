package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/database"
	"github.com/MoeMahhouk/go-tcb-notify/internal/server"
	"github.com/MoeMahhouk/go-tcb-notify/internal/services"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatal("Failed to load configuration:", err)
	}

	// Setup logging
	setupLogging(cfg.LogLevel)

	logrus.Info("Starting go-tcb-notify service")

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logrus.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		logrus.Fatal("Failed to run database migrations:", err)
	}

	// Initialize services
	tcbFetcher := services.NewTCBFetcher(cfg, db)
	quoteChecker := services.NewQuoteChecker(cfg, db)
	alertPublisher := services.NewAlertPublisher(cfg)

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go tcbFetcher.Start(ctx)
	go quoteChecker.Start(ctx)

	// Start HTTP server
	srv := server.New(cfg, db, tcbFetcher, quoteChecker, alertPublisher)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Server failed to start:", err)
		}
	}()

	logrus.Info("Service started successfully")
	<-c

	logrus.Info("Shutting down service...")
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Error("Server forced to shutdown:", err)
	}

	logrus.Info("Service stopped")
}

func setupLogging(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
