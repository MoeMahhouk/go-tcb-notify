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

	// Intel PCS
	PCSBaseURL string

	// Ethereum / Registry
	EthereumRPCURL    string
	RegistryAddress   string
	RegistryABIPath   string
	RegistryEventName string
	StartBlock        uint64
	BatchSize         uint64
	Prod              bool

	// Webhook
	WebhookURL     string
	WebhookTimeout time.Duration

	// Service intervals
	TCBFetchInterval   time.Duration
	QuoteCheckInterval time.Duration

	// Alerts
	AlertCooldown time.Duration
}

func Load() (*Config, error) {
	// durations
	tcbFetchInterval, _ := time.ParseDuration(envFirst("TCB_FETCH_INTERVAL", "TCB_CHECK_INTERVAL", "1h"))
	quoteCheckInterval, _ := time.ParseDuration(envFirst("QUOTE_CHECK_INTERVAL", "30m"))
	webhookTimeout, _ := time.ParseDuration(envFirst("WEBHOOK_TIMEOUT", "30s"))
	alertCooldown, _ := time.ParseDuration(envFirst("ALERT_COOLDOWN", "1h"))

	// numeric
	startBlock, _ := strconv.ParseUint(envFirst("START_BLOCK", "0"), 10, 64)
	batchSize, _ := strconv.ParseUint(envFirst("BATCH_SIZE", "1000"), 10, 64)

	// bools
	metricsEnabled, _ := strconv.ParseBool(envFirst("METRICS_ENABLED", "true"))
	prodFlag, _ := strconv.ParseBool(envFirst("PROD", "false"))

	return &Config{
		Port:           envFirst("PORT", "8080"),
		LogLevel:       envFirst("LOG_LEVEL", "info"),
		DatabaseURL:    envFirst("DATABASE_URL", "postgres://localhost/tcb_notify?sslmode=disable"),
		MetricsEnabled: metricsEnabled,

		PCSBaseURL: envFirst("PCS_BASE_URL", "https://api.trustedservices.intel.com"),

		// accept ETHEREUM_RPC_URL or RPC_URL
		EthereumRPCURL:    envFirst("ETHEREUM_RPC_URL", "RPC_URL", ""),
		RegistryAddress:   envFirst("REGISTRY_ADDRESS", ""),
		RegistryABIPath:   envFirst("REGISTRY_ABI_PATH", ""),
		RegistryEventName: envFirst("REGISTRY_EVENT_NAME", "TEEServiceRegistered"),
		StartBlock:        startBlock,
		BatchSize:         batchSize,
		Prod:              prodFlag,

		// accept WEBHOOK_URL or ALERT_WEBHOOK_URL
		WebhookURL:     envFirst("WEBHOOK_URL", "ALERT_WEBHOOK_URL", ""),
		WebhookTimeout: webhookTimeout,

		TCBFetchInterval:   tcbFetchInterval,
		QuoteCheckInterval: quoteCheckInterval,
		AlertCooldown:      alertCooldown,
	}, nil
}

func envFirst(keys ...string) string {
	for i := 0; i < len(keys); i++ {
		if val := os.Getenv(keys[i]); val != "" {
			return val
		}
	}
	return keys[len(keys)-1] // last is default/fallback literal
}
