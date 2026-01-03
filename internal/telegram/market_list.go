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

	// Get tracked markets
	markets, err := b.storage.GetTrackedMarketsByUserID(ctx, user.ID)
	if err != nil {
		b.log.Errorf("Failed to get tracked markets: %v", err)
		b.SendMessage(chatID, MsgErrorOccurred, BuildMainMenu())
		return
	}

	// Format and send the message
	message := FormatMarketsList(markets)
	b.SendMessage(chatID, message, BuildBackButton())
}
