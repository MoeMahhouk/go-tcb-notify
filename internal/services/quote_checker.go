package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/sirupsen/logrus"
)

type QuoteChecker struct {
	config *config.Config
	db     *sql.DB
}

func NewQuoteChecker(cfg *config.Config, db *sql.DB) *QuoteChecker {
	return &QuoteChecker{
		config: cfg,
		db:     db,
	}
}

func (c *QuoteChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(c.config.QuoteCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Quote Checker stopping...")
			return
		case <-ticker.C:
			c.checkQuotes(ctx)
		}
	}
}

func (c *QuoteChecker) checkQuotes(ctx context.Context) {
	logrus.Debug("Checking monitored quotes...")

	// Get quotes that need checking
	quotes, err := c.getQuotesNeedingCheck()
	if err != nil {
		logrus.WithError(err).Error("Failed to get quotes needing check")
		return
	}

	for _, quote := range quotes {
		if err := c.verifyQuote(quote); err != nil {
			logrus.WithError(err).WithField("address", quote).Error("Failed to verify quote")
		}
	}
}

func (c *QuoteChecker) getQuotesNeedingCheck() ([]string, error) {
	query := `
		SELECT address FROM monitored_tdx_quotes 
		WHERE needs_update = true OR last_checked < NOW() - INTERVAL '1 hour'`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []string
	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			return nil, err
		}
		quotes = append(quotes, address)
	}

	return quotes, nil
}

func (c *QuoteChecker) verifyQuote(address string) error {
	// This would integrate with go-tdx-guest library to verify the quote
	// against current TCB information
	logrus.WithField("address", address).Debug("Verifying quote")
	
	// Update last_checked timestamp
	_, err := c.db.Exec(
		"UPDATE monitored_tdx_quotes SET last_checked = NOW(), needs_update = false WHERE address = $1",
		address,
	)
	return err
}