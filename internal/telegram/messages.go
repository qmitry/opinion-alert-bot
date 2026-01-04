package telegram

import (
	"fmt"
	"strings"
	"time"
)

const (
	MsgWelcome = `Welcome to Opinion.Trade Price Alert Bot!

This bot monitors token prices on Opinion.Trade markets and sends you alerts when prices change by your specified threshold.

Features:
- Price spike detection
- Set unlimited price alerts per market
- Real-time notifications
- Track up to 10 markets

Use the menu below to get started.`

	MsgHelp = `*Opinion.Trade Alert Bot - Help*

*How it works:*
The bot tracks market prices and alerts you when prices make changes and spikes. If the price change exceeds your threshold, you'll receive an alert.

*Commands:*
/start - Show main menu
/help - Show this help message

*Creating Alerts:*
1. Click "Create Alert"
2. Enter the market ID (you can find it in the URL when the market is open on Opinion.Trade)
3. Enter your minimum price change threshold (e.g., 20 for ±20%)

*Limits:*
- Maximum 10 markets per user
- Unlimited alerts per market`

	MsgMarketIDPrompt    = "Please enter the Opinion.Trade market ID:\n\nTip: You can find the market ID in the URL when viewing a market on Opinion.Trade (e.g., app.opinion.trade/detail?topicId=1098 → market ID is 1098)"
	MsgThresholdPrompt   = "Enter the minimum price change threshold percentage (e.g., 20 for ±20%):"
	MsgAlertCreated      = "Alert created successfully! You'll be notified when the price changes by ±%.1f%% within 1 minute."
	MsgAlertDeleted      = "Alert deleted successfully."
	MsgInvalidMarketID   = "Invalid market ID. Please enter a valid market ID."
	MsgInvalidThreshold  = "Invalid threshold. Please enter a number between 1 and 100."
	MsgMaxMarketsReached = "You've reached the maximum of 10 tracked markets. Delete an alert for a market you no longer want to track."
	MsgNoAlerts          = "You don't have any alerts set up yet. Click 'Create Alert' to get started!"
	MsgNoMarketsTracked  = "You're not tracking any markets yet."
	MsgConfirmDelete     = "Are you sure you want to delete this alert?"
	MsgCancelled         = "Operation cancelled."
	MsgUnknownCommand    = "Unknown command. Use /start to see available options."
	MsgErrorOccurred     = "An error occurred. Please try again later."
	MsgMarketNotFound    = "Market not found. Please check the market ID and try again."
)

// FormatAlertNotification formats a price spike alert message
func FormatAlertNotification(marketTitle, marketID string, previousPrice, currentPrice, changePct, threshold float64) string {
	direction := "+"
	if changePct < 0 {
		direction = ""
	}

	return fmt.Sprintf(`📈 *Price Spike Alert!*

*Market:* %s (#%s)
*Current Price:* $%.4f
*1 min ago:* $%.4f
*Change:* %s%.2f%% (threshold: ±%.1f%%)

*Triggered:* %s UTC`,
		escapeMarkdown(marketTitle),
		marketID,
		currentPrice,
		previousPrice,
		direction,
		changePct,
		threshold,
		time.Now().UTC().Format("2006-01-02 15:04:05"),
	)
}

// FormatAlertsList formats the list of user's alerts
func FormatAlertsList(alerts map[string][]AlertInfo) string {
	if len(alerts) == 0 {
		return MsgNoAlerts
	}

	var sb strings.Builder
	sb.WriteString("*Your Alerts*\n\n")

	for marketID, alertList := range alerts {
		// Use market name if available
		marketName := alertList[0].MarketName
		if marketName == "" {
			marketName = "Market #" + marketID
		}

		sb.WriteString(fmt.Sprintf("*%s*\n", marketName))
		for i, alert := range alertList {
			sb.WriteString(fmt.Sprintf("%d. Threshold: ±%.1f%%\n", i+1, alert.ThresholdPct))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("_Total markets tracked: %d/10_", len(alerts)))

	return sb.String()
}

// FormatMarketsList formats the list of tracked markets with names
func FormatMarketsList(markets []MarketInfo) string {
	if len(markets) == 0 {
		return MsgNoMarketsTracked
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*Tracked Markets* (%d/10)\n\n", len(markets)))

	for i, market := range markets {
		name := market.MarketName
		if name == "" {
			name = "Market #" + market.MarketID
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, name))
	}

	return sb.String()
}

// AlertInfo holds alert display information
type AlertInfo struct {
	ID           int64
	MarketID     string
	MarketName   string
	ThresholdPct float64
}

// MarketInfo holds market display information
type MarketInfo struct {
	MarketID   string
	MarketName string
}

// escapeMarkdown escapes special characters for Telegram MarkdownV2
func escapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}
