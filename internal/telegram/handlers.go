package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleMessage processes incoming text messages
func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	userID := message.From.ID
	username := message.From.UserName

	// Create or get user in database
	_, err := b.storage.CreateOrGetUser(ctx, userID, username)
	if err != nil {
		b.log.Errorf("Failed to create/get user: %v", err)
		b.SendMessage(message.Chat.ID, MsgErrorOccurred, nil)
		return
	}

	// Handle commands
	if message.IsCommand() {
		b.handleCommand(ctx, message)
		return
	}

	// Handle conversation flow based on user state
	state := b.getUserState(userID)

	switch state.Step {
	case "awaiting_market_id":
		b.handleMarketIDInput(ctx, message)
	case "awaiting_threshold":
		b.handleThresholdInput(ctx, message)
	default:
		// No active conversation, show unknown command message
		b.SendMessage(message.Chat.ID, MsgUnknownCommand, BuildMainMenu())
	}
}

// handleCommand processes bot commands
func (b *Bot) handleCommand(ctx context.Context, message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		b.handleStart(ctx, message)
	case "help":
		b.handleHelp(ctx, message)
	default:
		b.SendMessage(message.Chat.ID, MsgUnknownCommand, BuildMainMenu())
	}
}

// handleStart handles the /start command
func (b *Bot) handleStart(ctx context.Context, message *tgbotapi.Message) {
	b.clearUserState(message.From.ID)
	b.SendMessage(message.Chat.ID, MsgWelcome, BuildMainMenu())
}

// handleHelp handles the /help command
func (b *Bot) handleHelp(ctx context.Context, message *tgbotapi.Message) {
	b.SendMessage(message.Chat.ID, MsgHelp, BuildBackButton())
}

// handleMarketIDInput processes market ID input
func (b *Bot) handleMarketIDInput(ctx context.Context, message *tgbotapi.Message) {
	marketID := strings.TrimSpace(message.Text)

	if marketID == "" {
		b.SendMessage(message.Chat.ID, MsgInvalidMarketID, nil)
		return
	}

	// Validate market exists by fetching details
	// For multi-outcome markets, this will automatically select the first outcome token
	marketDetails, err := b.apiClient.GetMarketDetails(ctx, marketID)
	if err != nil {
		b.log.Warnf("Market %s not found: %v", marketID, err)
		b.SendMessage(message.Chat.ID, MsgMarketNotFound, BuildBackButton())
		b.clearUserState(message.From.ID)
		return
	}

	// Store market ID, name, and token ID (automatically set for both binary and multi-outcome)
	state := b.getUserState(message.From.ID)
	state.MarketID = marketID
	state.Data["market_name"] = marketDetails.MarketTitle
	state.Data["token_id"] = marketDetails.YesTokenID
	state.Step = "awaiting_threshold"

	// Confirm market and ask for threshold with quick-select buttons
	confirmMsg := fmt.Sprintf("✅ Market found: <b>%s</b>\n\n%s",
		marketDetails.MarketTitle, MsgThresholdPrompt)
	b.SendMessage(message.Chat.ID, confirmMsg, BuildThresholdSelectionMenu())
}

// handleThresholdInput processes threshold percentage input
func (b *Bot) handleThresholdInput(ctx context.Context, message *tgbotapi.Message) {
	thresholdStr := strings.TrimSpace(message.Text)

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil || threshold < 1 || threshold > 100 {
		b.SendMessage(message.Chat.ID, MsgInvalidThreshold, nil)
		return
	}

	state := b.getUserState(message.From.ID)
	b.createAlert(ctx, message.Chat.ID, message.From.ID, state, threshold)
}

// createAlert is a helper function to create an alert with the given threshold
func (b *Bot) createAlert(ctx context.Context, chatID int64, userID int64, state *UserState, threshold float64) {
	// Get user from database
	user, err := b.storage.GetUserByTelegramID(ctx, userID)
	if err != nil {
		b.log.Errorf("Failed to get user: %v", err)
		b.SendMessage(chatID, MsgErrorOccurred, BuildMainMenu())
		b.clearUserState(userID)
		return
	}

	// Get market name and token ID from state (was saved during market ID validation)
	marketName, _ := state.Data["market_name"].(string)
	if marketName == "" {
		marketName = "Market " + state.MarketID
	}

	var tokenID *string
	if tokenIDStr, ok := state.Data["token_id"].(string); ok && tokenIDStr != "" {
		tokenID = &tokenIDStr
	}

	// Create the alert
	_, err = b.storage.CreateAlert(ctx, user.ID, state.MarketID, marketName, tokenID, threshold)
	if err != nil {
		if strings.Contains(err.Error(), "cannot track more than") {
			b.SendMessage(chatID, MsgMaxMarketsReached, BuildMainMenu())
		} else {
			b.log.Errorf("Failed to create alert: %v", err)
			b.SendMessage(chatID, MsgErrorOccurred, BuildMainMenu())
		}
		b.clearUserState(userID)
		return
	}

	// Success - show market name
	successMsg := fmt.Sprintf("✅ Alert created successfully!\n\n<b>Market:</b> %s\n<b>Threshold:</b> ±%.1f%%\n\nYou'll be notified when the price changes by this amount.",
		marketName, threshold)
	b.SendMessage(chatID, successMsg, BuildMainMenu())
	b.clearUserState(userID)
}

