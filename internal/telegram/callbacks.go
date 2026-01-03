package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	state := b.getUserState(callback.From.ID)
	state.Step = "awaiting_market_id"

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, MsgMarketIDPrompt)
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
	msg.ParseMode = "Markdown"
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
