package telegram

import (
	"context"
)

// showMyMarkets displays the markets being tracked by the user
func (b *Bot) showMyMarkets(ctx context.Context, chatID, userID int64) {
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

	// Extract unique markets with names
	marketMap := make(map[string]string) // marketID -> marketName
	for _, alert := range alerts {
		if !alert.IsActive {
			continue
		}
		if _, exists := marketMap[alert.MarketID]; !exists {
			marketMap[alert.MarketID] = alert.MarketName
		}
	}

	// Build MarketInfo slice
	var markets []MarketInfo
	for marketID, marketName := range marketMap {
		markets = append(markets, MarketInfo{
			MarketID:   marketID,
			MarketName: marketName,
		})
	}

	// Format and send the message
	message := FormatMarketsList(markets)
	b.SendMessage(chatID, message, BuildBackButton())
}
