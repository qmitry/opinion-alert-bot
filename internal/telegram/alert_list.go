package telegram

import (
	"context"
)

// showMyAlerts displays the user's alerts grouped by market
func (b *Bot) showMyAlerts(ctx context.Context, chatID, userID int64) {
	// Get user from database
	user, err := b.storage.GetUserByTelegramID(ctx, userID)
	if err != nil {
		b.log.Errorf("Failed to get user: %v", err)
		b.SendMessage(chatID, MsgErrorOccurred, BuildMainMenu())
		return
	}

	// Get all user's alerts
	alerts, err := b.storage.GetAlertsByUserID(ctx, user.ID)
	if err != nil {
		b.log.Errorf("Failed to get alerts: %v", err)
		b.SendMessage(chatID, MsgErrorOccurred, BuildMainMenu())
		return
	}

	// Group alerts by market
	alertsByMarket := make(map[string][]AlertInfo)
	for _, alert := range alerts {
		if !alert.IsActive {
			continue
		}

		alertInfo := AlertInfo{
			ID:           alert.ID,
			MarketID:     alert.MarketID,
			MarketName:   alert.MarketName,
			ThresholdPct: alert.ThresholdPct,
		}
		alertsByMarket[alert.MarketID] = append(alertsByMarket[alert.MarketID], alertInfo)
	}

	// Format and send the message
	message := FormatAlertsList(alertsByMarket)
	keyboard := BuildMainMenu()

	if len(alertsByMarket) > 0 {
		keyboard = BuildAlertListMenu(alertsByMarket)
	}

	b.SendMessage(chatID, message, keyboard)
}
