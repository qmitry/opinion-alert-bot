package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qmitry/opinion-alert-bot/internal/storage"
)

// handleCallbackQuery processes inline keyboard button clicks
func (b *Bot) handleCallbackQuery(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	// Answer the callback to remove the loading state
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	b.api.Request(callbackConfig)

	data := callback.Data

	// Route to appropriate handler based on callback data
	switch {
	case data == CallbackCreateAlert:
		b.handleCreateAlertCallback(ctx, callback)
	case data == CallbackMyAlerts:
		b.handleMyAlertsCallback(ctx, callback)
	case data == CallbackMyMarkets:
		b.handleMyMarketsCallback(ctx, callback)
	case data == CallbackHelp:
		b.handleHelpCallback(ctx, callback)
	case strings.HasPrefix(data, CallbackSelectMarket+"_"):
		b.handleSelectMarketCallback(ctx, callback)
	case data == CallbackCustomMarket:
		b.handleCustomMarketCallback(ctx, callback)
	case strings.HasPrefix(data, CallbackDeleteAlert+"_"):
		b.handleDeleteAlertCallback(ctx, callback)
	case strings.HasPrefix(data, CallbackConfirmDelete+"_"):
		b.handleConfirmDeleteCallback(ctx, callback)
	case data == CallbackCancelDelete:
		b.handleCancelDeleteCallback(ctx, callback)
	case data == "back_to_menu":
		b.handleBackToMenuCallback(ctx, callback)
	default:
		b.log.Warnf("Unknown callback data: %s", data)
	}
}

// handleCreateAlertCallback initiates the alert creation flow
func (b *Bot) handleCreateAlertCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	// Fetch random tracked markets from database
	randomMarkets, err := b.storage.GetRandomTrackedMarkets(ctx, 5)
	if err != nil {
		b.log.Errorf("Failed to get random markets: %v", err)
		// Fall back to empty list
		randomMarkets = []storage.MarketWithName{}
	}

	// Convert to FeaturedMarket format
	featuredMarkets := make([]FeaturedMarket, len(randomMarkets))
	for i, market := range randomMarkets {
		featuredMarkets[i] = FeaturedMarket{
			ID:   market.MarketID,
			Name: market.MarketName,
		}
	}

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		MsgSelectMarket,
	)
	keyboard := BuildMarketSelectionMenu(featuredMarkets)
	msg.ReplyMarkup = &keyboard
	b.api.Send(msg)
}

// handleSelectMarketCallback handles selection of a featured market
func (b *Bot) handleSelectMarketCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	// Extract market ID from callback data (format: "select_market_123")
	parts := strings.Split(callback.Data, "_")
	if len(parts) != 3 {
		b.log.Errorf("Invalid select market callback data: %s", callback.Data)
		return
	}

	marketID := parts[2]

	// Validate market exists by fetching details
	marketDetails, err := b.apiClient.GetMarketDetails(ctx, marketID)
	if err != nil {
		b.log.Warnf("Market %s not found: %v", marketID, err)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, MsgMarketNotFound)
		keyboard := BuildBackButton()
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
		return
	}

	// Store market ID and name
	state := b.getUserState(callback.From.ID)
	state.MarketID = marketID
	state.Data["market_name"] = marketDetails.MarketTitle

	// Delete the market selection message
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	b.api.Send(deleteMsg)

	// Check if it's a multi-outcome market (marketType 1 or no YesTokenID)
	if marketDetails.MarketType != 0 || marketDetails.YesTokenID == "" {
		// Multi-outcome market - ask for token ID
		state.Step = "awaiting_token_id"
		multiOutcomeMsg := fmt.Sprintf("Market selected: <b>%s</b>\n\n⚠️ This is a multi-outcome market. Please provide the specific token ID you want to track.\n\nYou can find token IDs by:\n1. Opening the market on Opinion.Trade\n2. Inspecting the specific outcome you want to track\n3. Looking for the token ID in the URL or market details", marketDetails.MarketTitle)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, multiOutcomeMsg)
		msg.ParseMode = "HTML"
		msg.DisableWebPagePreview = true
		b.api.Send(msg)
		return
	}

	// Binary market - use YesTokenID automatically
	state.Data["token_id"] = marketDetails.YesTokenID
	state.Step = "awaiting_threshold"

	// Send threshold prompt
	confirmMsg := fmt.Sprintf("Market selected: <b>%s</b>\n\n%s", marketDetails.MarketTitle, MsgThresholdPrompt)
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, confirmMsg)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	b.api.Send(msg)
}

// handleCustomMarketCallback prompts user to enter custom market ID
func (b *Bot) handleCustomMarketCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	state := b.getUserState(callback.From.ID)
	state.Step = "awaiting_market_id"

	// Delete the market selection message
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	b.api.Send(deleteMsg)

	// Send market ID prompt
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, MsgMarketIDPrompt)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	b.api.Send(msg)
}

// handleMyAlertsCallback shows the user's alerts
func (b *Bot) handleMyAlertsCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	b.showMyAlerts(ctx, callback.Message.Chat.ID, callback.From.ID)
}

// handleMyMarketsCallback shows the user's tracked markets
func (b *Bot) handleMyMarketsCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	b.showMyMarkets(ctx, callback.Message.Chat.ID, callback.From.ID)
}

// handleHelpCallback shows the help message
func (b *Bot) handleHelpCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, MsgHelp)
	msg.ParseMode = "HTML"
	keyboard := BuildBackButton()
	msg.ReplyMarkup = &keyboard
	b.api.Send(msg)
}

// handleDeleteAlertCallback shows delete confirmation
func (b *Bot) handleDeleteAlertCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	// Extract alert ID from callback data (format: "delete_alert_123")
	parts := strings.Split(callback.Data, "_")
	if len(parts) != 3 {
		b.log.Errorf("Invalid delete callback data: %s", callback.Data)
		return
	}

	alertID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		b.log.Errorf("Invalid alert ID in callback: %s", parts[2])
		return
	}

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		MsgConfirmDelete,
	)
	keyboard := BuildConfirmDeleteMenu(alertID)
	msg.ReplyMarkup = &keyboard
	b.api.Send(msg)
}

// handleConfirmDeleteCallback deletes the alert
func (b *Bot) handleConfirmDeleteCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	// Extract alert ID from callback data (format: "confirm_delete_123")
	parts := strings.Split(callback.Data, "_")
	if len(parts) != 3 {
		b.log.Errorf("Invalid confirm delete callback data: %s", callback.Data)
		return
	}

	alertID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		b.log.Errorf("Invalid alert ID in callback: %s", parts[2])
		return
	}

	// Get user
	user, err := b.storage.GetUserByTelegramID(ctx, callback.From.ID)
	if err != nil {
		b.log.Errorf("Failed to get user: %v", err)
		b.SendMessage(callback.Message.Chat.ID, MsgErrorOccurred, BuildMainMenu())
		return
	}

	// Delete the alert
	err = b.storage.DeleteAlert(ctx, alertID, user.ID)
	if err != nil {
		b.log.Errorf("Failed to delete alert: %v", err)
		b.SendMessage(callback.Message.Chat.ID, MsgErrorOccurred, BuildMainMenu())
		return
	}

	// Show success message and return to alerts list
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, MsgAlertDeleted)
	msg.ReplyMarkup = BuildMainMenu()
	b.api.Send(msg)

	// Delete the confirmation message
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	b.api.Send(deleteMsg)
}

// handleCancelDeleteCallback cancels alert deletion
func (b *Bot) handleCancelDeleteCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		MsgCancelled,
	)
	keyboard := BuildBackButton()
	msg.ReplyMarkup = &keyboard
	b.api.Send(msg)
}

// handleBackToMenuCallback returns to the main menu
func (b *Bot) handleBackToMenuCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	b.clearUserState(callback.From.ID)

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		MsgWelcome,
	)
	keyboard := BuildMainMenu()
	msg.ReplyMarkup = &keyboard
	b.api.Send(msg)
}
