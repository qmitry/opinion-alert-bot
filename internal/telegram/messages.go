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
- Real-time notifications

Use the menu below to get started.`

	MsgHelp = `<b>Opinion.Trade Alert Bot - Help</b>

<b>How it works:</b>
The bot tracks market prices and alerts you when prices make changes and spikes. If the price change exceeds your threshold, you'll receive an alert.

<b>Commands:</b>
/start - Show main menu
/help - Show this help message

<b>Creating Alerts:</b>
1. Click "Create Alert"
2. Enter the market ID (you can find it in the URL as topicId when the market is open on Opinion.Trade)
3. Enter your minimum price change threshold (e.g., 20 for ±20%)

<b>Limits:</b>
- Maximum 10 markets per user
- Unlimited alerts per market`

	MsgSelectMarket      = "Select a market to create an alert, or enter a custom market ID:"
	MsgMarketIDPrompt    = "Please enter the Opinion.Trade market ID:\n\nTip: You can find the market ID in the URL as topicId when viewing a market on Opinion.Trade (e.g., app.opinion.trade/detail?<b>topicId=1098</b> → market ID is 1098)"
	MsgThresholdPrompt   = "Enter the minimum price change threshold percentage for 1 minute (e.g., 20 for ±20%):"
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

	marketURL := fmt.Sprintf("https://app.opinion.trade/detail?topicId=%s", marketID)

	return fmt.Sprintf(`📈 <b>Price Spike Alert!</b>

<b>Market:</b> <a href="%s">%s</a>
<b>Current Price:</b> $%.4f
<b>Price 1 minute ago:</b> $%.4f
<b>Change:</b> %s%.2f%% (threshold: ±%.1f%%)

<b>Triggered:</b> %s UTC`,
		marketURL,
		marketTitle,
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

	// Convert map to slice for sorting
	type marketAlert struct {
		marketID   string
		alertList  []AlertInfo
		marketName string
	}
	var sortedAlerts []marketAlert
	for marketID, alertList := range alerts {
		marketName := alertList[0].MarketName
		if marketName == "" {
			marketName = "Market"
		}
		sortedAlerts = append(sortedAlerts, marketAlert{marketID, alertList, marketName})
	}

	// Sort by market name for consistent ordering
	// (In Go, we'd need to import "sort" package, but for now we'll keep map iteration)

	var sb strings.Builder
	sb.WriteString("<b>Your Alerts</b>\n\n")

	marketNum := 1
	for _, ma := range sortedAlerts {
		// Create display name with market ID (only if not already included)
		var displayName string
		if strings.HasPrefix(ma.marketName, "Market #") {
			// Already has the ID from migration, use as-is
			displayName = ma.marketName
		} else if ma.marketName == "Market" {
			// Empty name, add ID
			displayName = fmt.Sprintf("Market #%s", ma.marketID)
		} else {
			// Real market name, append ID
			displayName = fmt.Sprintf("%s #%s", ma.marketName, ma.marketID)
		}

		// Create clickable link to the market (HTML format)
		marketURL := fmt.Sprintf("https://app.opinion.trade/detail?topicId=%s", ma.marketID)
		sb.WriteString(fmt.Sprintf("%d. <b><a href=\"%s\">%s</a></b>\n", marketNum, marketURL, displayName))

		// Only show first alert (since we'll limit to 1 per market)
		if len(ma.alertList) > 0 {
			sb.WriteString(fmt.Sprintf("Threshold: ±%.1f%%\n", ma.alertList[0].ThresholdPct))
		}
		sb.WriteString("\n")
		marketNum++
	}

	sb.WriteString(fmt.Sprintf("<i>Total markets tracked: %d/10</i>", len(alerts)))

	return sb.String()
}

// FormatMarketsList formats the list of tracked markets with names
func FormatMarketsList(markets []MarketInfo) string {
	if len(markets) == 0 {
		return MsgNoMarketsTracked
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Tracked Markets</b> (%d/10)\n\n", len(markets)))

	for i, market := range markets {
		name := market.MarketName
		if name == "" {
			name = "Market"
		}

		// Create display name with market ID (only if not already included)
		var displayName string
		if strings.HasPrefix(name, "Market #") {
			// Already has the ID from migration, use as-is
			displayName = name
		} else if name == "Market" {
			// Empty name, add ID
			displayName = fmt.Sprintf("Market #%s", market.MarketID)
		} else {
			// Real market name, append ID
			displayName = fmt.Sprintf("%s #%s", name, market.MarketID)
		}

		// Create clickable link to the market (HTML format)
		marketURL := fmt.Sprintf("https://app.opinion.trade/detail?topicId=%s", market.MarketID)
		sb.WriteString(fmt.Sprintf("%d. <a href=\"%s\">%s</a>\n", i+1, marketURL, displayName))
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

// FeaturedMarket holds information about a featured market
type FeaturedMarket struct {
	ID   string
	Name string
}

// FeaturedMarkets is a list of popular markets to suggest to users
var FeaturedMarkets = []FeaturedMarket{
	{ID: "2368", Name: "Tim Cook out as Apple CEO by March 31?"},
	{ID: "1098", Name: "First to 5k: Gold or ETH?"},
	{ID: "2366", Name: "Trump Executive Order on Bitcoin Strategic Reserve?"},
	{ID: "2357", Name: "Elon Musk posts on X 100+ times in a day?"},
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
