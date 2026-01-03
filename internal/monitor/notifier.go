package monitor

import (
	"context"

	"github.com/qmitry/opinion-alert-bot/internal/storage"
	"github.com/qmitry/opinion-alert-bot/internal/telegram"
	"github.com/sirupsen/logrus"
)

// Notifier handles sending alert notifications to users
type Notifier struct {
	bot     *telegram.Bot
	storage *storage.Storage
	log     *logrus.Logger
}

// NewNotifier creates a new notifier instance
func NewNotifier(bot *telegram.Bot, storage *storage.Storage, log *logrus.Logger) *Notifier {
	return &Notifier{
		bot:     bot,
		storage: storage,
		log:     log,
	}
}

// SendPriceAlert sends a price spike alert to a user
func (n *Notifier) SendPriceAlert(ctx context.Context, alert *storage.Alert, marketTitle string, previousPrice, currentPrice, changePct float64) error {
	// Get user to find their chat ID
	user, err := n.storage.GetUserByID(ctx, alert.UserID)
	if err != nil {
		n.log.Errorf("Failed to get user %d: %v", alert.UserID, err)
		return err
	}

	// Create alert history record
	history, err := n.storage.CreateAlertHistory(ctx, alert.ID, alert.MarketID, previousPrice, currentPrice, changePct)
	if err != nil {
		n.log.Errorf("Failed to create alert history: %v", err)
		return err
	}

	// Format and send notification
	message := telegram.FormatAlertNotification(
		marketTitle,
		alert.MarketID,
		previousPrice,
		currentPrice,
		changePct,
		alert.ThresholdPct,
	)

	err = n.bot.SendAlertNotification(user.TelegramID, message)
	if err != nil {
		n.log.Errorf("Failed to send notification to user %d: %v", user.TelegramID, err)
		return err
	}

	// Mark message as sent
	if err := n.storage.MarkMessageSent(ctx, history.ID); err != nil {
		n.log.Warnf("Failed to mark message as sent for history %d: %v", history.ID, err)
	}

	n.log.Infof("Sent price alert to user %d for market %s (%.2f%%)", user.TelegramID, alert.MarketID, changePct)
	return nil
}
