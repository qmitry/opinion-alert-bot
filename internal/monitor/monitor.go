package monitor

import (
	"context"
	"time"

	"github.com/qmitry/opinion-alert-bot/internal/api"
	"github.com/qmitry/opinion-alert-bot/internal/storage"
	"github.com/qmitry/opinion-alert-bot/internal/telegram"
	"github.com/sirupsen/logrus"
)

// Monitor represents the main monitoring service
type Monitor struct {
	storage      *storage.Storage
	apiClient    *api.Client
	priceChecker *PriceChecker
	pollInterval time.Duration
	log          *logrus.Logger
}

// NewMonitor creates a new monitor instance
func NewMonitor(storage *storage.Storage, apiClient *api.Client, bot *telegram.Bot, pollInterval int, log *logrus.Logger) *Monitor {
	notifier := NewNotifier(bot, storage, log)
	priceChecker := NewPriceChecker(apiClient, storage, notifier, log)

	return &Monitor{
		storage:      storage,
		apiClient:    apiClient,
		priceChecker: priceChecker,
		pollInterval: time.Duration(pollInterval) * time.Second,
		log:          log,
	}
}

// Start begins the monitoring loop
func (m *Monitor) Start(ctx context.Context) error {
	m.log.Infof("Starting market monitor (poll interval: %v)", m.pollInterval)

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	// Run initial check immediately
	m.runMonitoringCycle(ctx)

	for {
		select {
		case <-ctx.Done():
			m.log.Info("Stopping market monitor...")
			return nil
		case <-ticker.C:
			m.runMonitoringCycle(ctx)
		}
	}
}

// runMonitoringCycle performs one monitoring cycle
func (m *Monitor) runMonitoringCycle(ctx context.Context) {
	m.log.Debug("Starting monitoring cycle...")

	// Get all active alerts
	alerts, err := m.storage.GetActiveAlerts(ctx)
	if err != nil {
		m.log.Errorf("Failed to get active alerts: %v", err)
		return
	}

	if len(alerts) == 0 {
		m.log.Debug("No active alerts to monitor")
		return
	}

	// Get unique markets being monitored
	markets, err := m.storage.GetUniqueTrackedMarkets(ctx)
	if err != nil {
		m.log.Errorf("Failed to get tracked markets: %v", err)
		return
	}

	m.log.Debugf("Monitoring %d markets with %d alerts", len(markets), len(alerts))

	// Group alerts by market for efficient processing
	alertsByMarket := make(map[string][]storage.Alert)
	for _, alert := range alerts {
		alertsByMarket[alert.MarketID] = append(alertsByMarket[alert.MarketID], alert)
	}

	// Check each market
	for _, marketID := range markets {
		marketAlerts := alertsByMarket[marketID]
		if len(marketAlerts) == 0 {
			continue
		}

		// Check market price and trigger alerts if needed
		if err := m.priceChecker.CheckMarketPrice(ctx, marketID, marketAlerts); err != nil {
			m.log.Warnf("Error checking market %s: %v", marketID, err)
			continue
		}
	}

	// Cleanup old price data (older than 5 minutes)
	if err := m.storage.CleanupOldPrices(ctx, 5*time.Minute); err != nil {
		m.log.Warnf("Failed to cleanup old prices: %v", err)
	}

	m.log.Debug("Monitoring cycle completed")
}
