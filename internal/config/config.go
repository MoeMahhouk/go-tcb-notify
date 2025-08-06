package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server configuration
	Port           string
	LogLevel       string
	DatabaseURL    string
	MetricsEnabled bool

	// Intel PCS configuration
	PCSBaseURL string

	// Ethereum configuration
	EthereumRPCURL  string
	RegistryAddress string
	StartBlock      uint64
	BatchSize       uint64

	// Webhook configuration
	WebhookURL     string
	WebhookTimeout time.Duration

	// Service intervals
	TCBFetchInterval   time.Duration
	QuoteCheckInterval time.Duration

	// Alert configuration
	AlertCooldown time.Duration
}

func Load() (*Config, error) {
	tcbFetchInterval, _ := time.ParseDuration(getEnvOrDefault("TCB_FETCH_INTERVAL", "1h"))
	quoteCheckInterval, _ := time.ParseDuration(getEnvOrDefault("QUOTE_CHECK_INTERVAL", "30m"))
	webhookTimeout, _ := time.ParseDuration(getEnvOrDefault("WEBHOOK_TIMEOUT", "30s"))
	alertCooldown, _ := time.ParseDuration(getEnvOrDefault("ALERT_COOLDOWN", "1h"))

	startBlock, _ := strconv.ParseUint(getEnvOrDefault("START_BLOCK", "0"), 10, 64)
	batchSize, _ := strconv.ParseUint(getEnvOrDefault("BATCH_SIZE", "1000"), 10, 64)

	metricsEnabled, _ := strconv.ParseBool(getEnvOrDefault("METRICS_ENABLED", "true"))

	return &Config{
		Port:           getEnvOrDefault("PORT", "8080"),
		LogLevel:       getEnvOrDefault("LOG_LEVEL", "info"),
		DatabaseURL:    getEnvOrDefault("DATABASE_URL", "postgres://localhost/tcb_notify?sslmode=disable"),
		MetricsEnabled: metricsEnabled,

		PCSBaseURL: getEnvOrDefault("PCS_BASE_URL", "https://api.trustedservices.intel.com"),

		EthereumRPCURL:  getEnvOrDefault("ETHEREUM_RPC_URL", ""),
		RegistryAddress: getEnvOrDefault("REGISTRY_ADDRESS", ""),
		StartBlock:      startBlock,
		BatchSize:       batchSize,

		WebhookURL:     getEnvOrDefault("WEBHOOK_URL", ""),
		WebhookTimeout: webhookTimeout,

		TCBFetchInterval:   tcbFetchInterval,
		QuoteCheckInterval: quoteCheckInterval,
		AlertCooldown:      alertCooldown,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
