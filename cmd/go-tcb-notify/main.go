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

	// Initialize services with proper dependency injection
	fmspcService := services.NewFMSPCService(cfg, db)

	registryService, err := services.NewRegistryService(cfg, db)
	if err != nil {
		logrus.Fatal("Failed to initialize registry service:", err)
	}

	alertPublisher := services.NewAlertPublisher(cfg, db)
	tcbFetcher := services.NewTCBFetcher(cfg, db, fmspcService)
	quoteChecker := services.NewQuoteChecker(cfg, db, registryService, alertPublisher)

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Warm FMSPC list (FK for quotes)
	logrus.Info("Performing initial FMSPC fetch...")
	if err := fmspcService.FetchAndStoreAllFMSPCs(ctx); err != nil {
		logrus.WithError(err).Warn("Failed to fetch FMSPCs on startup, will retry during TCB checks")
	}

	// Start background services
	go tcbFetcher.Start(ctx)
	go quoteChecker.Start(ctx)

	// Start HTTP server with all services
	srv := server.New(cfg, db, tcbFetcher, quoteChecker, alertPublisher, fmspcService, registryService)

	// HTTP serve + shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Server failed to start:", err)
		}
	}()
	logrus.Info("Service started successfully")
	<-stop

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
