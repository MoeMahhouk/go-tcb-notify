package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server configuration
	Port     string
	LogLevel string

	// Database configuration
	DatabaseURL string

	// Ethereum configuration
	RPCURL          string
	RegistryAddress string

	// Intel PCS configuration
	PCSBaseURL string

	// Monitoring intervals
	TCBCheckInterval   time.Duration
	QuoteCheckInterval time.Duration

	// Alerting configuration
	AlertWebhookURL string

	// Metrics configuration
	MetricsEnabled bool
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost/go_tcb_notify?sslmode=disable"),

		RPCURL:          getEnv("RPC_URL", "http://localhost:8545"),
		RegistryAddress: getEnv("REGISTRY_ADDRESS", ""),

		PCSBaseURL: getEnv("PCS_BASE_URL", "https://api.trustedservices.intel.com"),

		TCBCheckInterval:   getDurationEnv("TCB_CHECK_INTERVAL", time.Hour),
		QuoteCheckInterval: getDurationEnv("QUOTE_CHECK_INTERVAL", 5*time.Minute),

		AlertWebhookURL: getEnv("ALERT_WEBHOOK_URL", ""),

		MetricsEnabled: getBoolEnv("METRICS_ENABLED", true),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}