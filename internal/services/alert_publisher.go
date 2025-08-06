package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/sirupsen/logrus"
)

type AlertPublisher struct {
	config *config.Config
	db     *sql.DB
	client *http.Client
}

func NewAlertPublisher(cfg *config.Config, db *sql.DB) *AlertPublisher {
	return &AlertPublisher{
		config: cfg,
		db:     db,
		client: &http.Client{
			Timeout: cfg.WebhookTimeout,
		},
	}
}

func (a *AlertPublisher) SendAlert(alert *models.Alert) error {
	// Store alert in history
	if err := a.storeAlertHistory(alert); err != nil {
		logrus.WithError(err).Error("Failed to store alert history")
	}

	// Send webhook if configured
	if a.config.WebhookURL != "" {
		if err := a.sendWebhook(alert); err != nil {
			logrus.WithError(err).Error("Failed to send webhook")
			return err
		}
	}

	// Log the alert
	logrus.WithFields(logrus.Fields{
		"severity":    alert.Severity,
		"address":     alert.Quote.Address,
		"reason":      alert.Quote.Reason,
		"newStatus":   alert.Quote.NewStatus,
		"advisoryIds": alert.Quote.AdvisoryIDs,
	}).Warn("TCB Alert triggered")

	return nil
}

func (a *AlertPublisher) sendWebhook(alert *models.Alert) error {
	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.config.WebhookTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", a.config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-tcb-notify/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	logrus.WithFields(logrus.Fields{
		"url":     a.config.WebhookURL,
		"status":  resp.StatusCode,
		"address": alert.Quote.Address,
	}).Info("Webhook sent successfully")

	return nil
}

func (a *AlertPublisher) storeAlertHistory(alert *models.Alert) error {
	details, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert details: %w", err)
	}

	_, err = a.db.Exec(`
		INSERT INTO alert_history (quote_address, reason, details, sent_at)
		VALUES ($1, $2, $3, $4)`,
		alert.Quote.Address,
		alert.Quote.Reason,
		details,
		alert.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to store alert history: %w", err)
	}

	return nil
}

func (a *AlertPublisher) GetAlertHistory(address string, limit int) ([]*models.AlertHistory, error) {
	query := `
		SELECT id, quote_address, reason, details, sent_at, acknowledged
		FROM alert_history
		WHERE quote_address = $1
		ORDER BY sent_at DESC
		LIMIT $2`

	rows, err := a.db.Query(query, address, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*models.AlertHistory
	for rows.Next() {
		alert := &models.AlertHistory{}
		err := rows.Scan(
			&alert.ID,
			&alert.QuoteAddress,
			&alert.Reason,
			&alert.Details,
			&alert.SentAt,
			&alert.Acknowledged,
		)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (a *AlertPublisher) AcknowledgeAlert(id int) error {
	_, err := a.db.Exec(`
		UPDATE alert_history 
		SET acknowledged = true 
		WHERE id = $1`,
		id,
	)
	return err
}
