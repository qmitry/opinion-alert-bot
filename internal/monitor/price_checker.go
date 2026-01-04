package monitor

import (
	"context"
	"database/sql"
	"math"

	"github.com/qmitry/opinion-alert-bot/internal/api"
	"github.com/qmitry/opinion-alert-bot/internal/storage"
	"github.com/sirupsen/logrus"
)

// PriceChecker handles price spike detection logic
type PriceChecker struct {
	apiClient *api.Client
	storage   *storage.Storage
	notifier  *Notifier
	log       *logrus.Logger
}

// NewPriceChecker creates a new price checker instance
func NewPriceChecker(apiClient *api.Client, storage *storage.Storage, notifier *Notifier, log *logrus.Logger) *PriceChecker {
	return &PriceChecker{
		apiClient: apiClient,
		storage:   storage,
		notifier:  notifier,
		log:       log,
	}
}

// CheckMarketPrice checks a market for price spikes and triggers alerts
func (pc *PriceChecker) CheckMarketPrice(ctx context.Context, marketID string, alerts []storage.Alert) error {
	// Get market details for market name
	marketDetails, err := pc.apiClient.GetMarketDetails(ctx, marketID)
	if err != nil {
		pc.log.Warnf("Failed to get market details for %s: %v", marketID, err)
		return err
	}

	// Determine which token to track
	// For alerts with token_id set, use that; otherwise fall back to YesTokenID
	var tokenID string
	if len(alerts) > 0 && alerts[0].TokenID != nil && *alerts[0].TokenID != "" {
		tokenID = *alerts[0].TokenID
	} else {
		tokenID = marketDetails.YesTokenID
	}

	if tokenID == "" {
		pc.log.Warnf("No token ID available for market %s (may be multi-outcome market without token_id set)", marketID)
		return nil
	}

	// Get current token price
	tokenPrice, err := pc.apiClient.GetTokenPrice(ctx, tokenID)
	if err != nil {
		pc.log.Warnf("Failed to get token price for %s: %v", tokenID, err)
		return err
	}

	// Parse price
	currentPrice, err := api.ParseTokenPrice(tokenPrice.Price)
	if err != nil {
		pc.log.Errorf("Failed to parse token price: %v", err)
		return err
	}

	// Parse size
	size, err := api.ParseTokenSize(tokenPrice.Size)
	if err != nil {
		pc.log.Warnf("Failed to parse token size: %v", err)
		size = 0
	}

	// Store current price
	if err := pc.storage.StoreTokenPrice(ctx, tokenID, marketID, currentPrice, tokenPrice.Side, size); err != nil {
		pc.log.Errorf("Failed to store token price: %v", err)
		return err
	}

	// Get price from 1 minute ago
	previousTokenPrice, err := pc.storage.GetPriceOneMinuteAgo(ctx, marketID)
	if err != nil {
		if err == sql.ErrNoRows {
			pc.log.Debugf("No historical price available for market %s yet", marketID)
			return nil
		}
		pc.log.Errorf("Failed to get previous price: %v", err)
		return err
	}

	previousPrice := previousTokenPrice.Price

	// Calculate percentage change
	changePct := ((currentPrice - previousPrice) / previousPrice) * 100

	pc.log.Debugf("Market %s (token %s): current=%.4f, previous=%.4f, change=%.2f%%",
		marketID, tokenID, currentPrice, previousPrice, changePct)

	// Check each alert for this market
	for _, alert := range alerts {
		if alert.MarketID != marketID || !alert.IsActive {
			continue
		}

		// Check if change exceeds threshold (in either direction)
		if math.Abs(changePct) >= alert.ThresholdPct {
			pc.log.Infof("Alert triggered for market %s: %.2f%% change (threshold: %.1f%%)",
				marketID, changePct, alert.ThresholdPct)

			// Send notification
			if err := pc.notifier.SendPriceAlert(ctx, &alert, marketDetails.MarketTitle, previousPrice, currentPrice, changePct); err != nil {
				pc.log.Errorf("Failed to send price alert: %v", err)
			}
		}
	}

	return nil
}
