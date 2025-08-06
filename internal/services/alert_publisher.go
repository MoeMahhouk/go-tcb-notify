package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/sirupsen/logrus"
)

type AlertPublisher struct {
	config *config.Config
	client *http.Client
}

func NewAlertPublisher(cfg *config.Config) *AlertPublisher {
	return &AlertPublisher{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *AlertPublisher) SendAlert(alert models.Alert) error {
	if p.config.AlertWebhookURL == "" {
		logrus.Warn("No webhook URL configured, skipping alert")
		return nil
	}

	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	resp, err := p.client.Post(p.config.AlertWebhookURL, "application/json", bytes.NewBuffer(alertJSON))
	if err != nil {
		return fmt.Errorf("failed to send alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	logrus.WithFields(logrus.Fields{
		"address":  alert.Quote.Address,
		"severity": alert.Severity,
		"reason":   alert.Quote.Reason,
	}).Info("Alert sent successfully")

	return nil
}
