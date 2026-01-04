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
	marketDetails, err := b.apiClient.GetMarketDetails(ctx, marketID)
	if err != nil {
		b.log.Warnf("Market %s not found: %v", marketID, err)
		b.SendMessage(message.Chat.ID, MsgMarketNotFound, BuildBackButton())
		b.clearUserState(message.From.ID)
		return
	}

	// Store market ID and name, move to next step
	state := b.getUserState(message.From.ID)
	state.MarketID = marketID
	state.Data["market_name"] = marketDetails.MarketTitle
	state.Step = "awaiting_threshold"

	// Confirm market and ask for threshold
	confirmMsg := fmt.Sprintf("Market found: *%s*\n\n%s", marketDetails.MarketTitle, MsgThresholdPrompt)
	b.SendMessage(message.Chat.ID, confirmMsg, nil)
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

	// Get user from database
	user, err := b.storage.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil {
		b.log.Errorf("Failed to get user: %v", err)
		b.SendMessage(message.Chat.ID, MsgErrorOccurred, BuildMainMenu())
		b.clearUserState(message.From.ID)
		return
	}

	// Get market name from state (was saved during market ID validation)
	marketName, _ := state.Data["market_name"].(string)
	if marketName == "" {
		marketName = "Market " + state.MarketID
	}

	// Create the alert
	_, err = b.storage.CreateAlert(ctx, user.ID, state.MarketID, marketName, threshold)
	if err != nil {
		if strings.Contains(err.Error(), "cannot track more than") {
			b.SendMessage(message.Chat.ID, MsgMaxMarketsReached, BuildMainMenu())
		} else {
			b.log.Errorf("Failed to create alert: %v", err)
			b.SendMessage(message.Chat.ID, MsgErrorOccurred, BuildMainMenu())
		}
		b.clearUserState(message.From.ID)
		return
	}

	// Success - show market name
	successMsg := fmt.Sprintf("✅ Alert created successfully!\n\n*Market:* %s\n*Threshold:* ±%.1f%%\n\nYou'll be notified when the price changes by this amount.",
		marketName, threshold)
	b.SendMessage(message.Chat.ID, successMsg, BuildMainMenu())
	b.clearUserState(message.From.ID)
}

// formatString is a helper to format strings with placeholders
func formatString(format string, args ...interface{}) string {
	return strings.TrimSpace(formatMessage(format, args...))
}

func formatMessage(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}

	// Simple sprintf-like formatting
	result := format
	for _, arg := range args {
		switch v := arg.(type) {
		case float64:
			result = strings.Replace(result, "%.1f", strconv.FormatFloat(v, 'f', 1, 64), 1)
		case int:
			result = strings.Replace(result, "%d", strconv.Itoa(v), 1)
		case string:
			result = strings.Replace(result, "%s", v, 1)
		}
	}
	return result
}
