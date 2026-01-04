package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Callback data constants
const (
	CallbackCreateAlert     = "create_alert"
	CallbackMyAlerts        = "my_alerts"
	CallbackMyMarkets       = "my_markets"
	CallbackHelp            = "help"
	CallbackDeleteAlert     = "delete_alert"
	CallbackConfirmDelete   = "confirm_delete"
	CallbackCancelDelete    = "cancel_delete"
	CallbackSelectMarket    = "select_market"
	CallbackCustomMarket    = "custom_market"
	CallbackSelectThreshold = "select_threshold"
	CallbackCustomThreshold = "custom_threshold"
)

// BuildMainMenu creates the main menu inline keyboard
func BuildMainMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Create Alert", CallbackCreateAlert),
			tgbotapi.NewInlineKeyboardButtonData("My Alerts", CallbackMyAlerts),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("My Markets", CallbackMyMarkets),
			tgbotapi.NewInlineKeyboardButtonData("Help", CallbackHelp),
		),
	)
}

// BuildAlertListMenu creates the alert list menu with delete buttons
func BuildAlertListMenu(alerts map[string][]AlertInfo) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for marketID, alertList := range alerts {
		// Market header (not clickable)
		for _, alert := range alertList {
			// Use market name if available
			displayName := alert.MarketName
			if displayName == "" || displayName == "Market #"+marketID {
				displayName = "Market #" + marketID
			} else {
				displayName = fmt.Sprintf("%s #%s", displayName, marketID)
			}

			// Truncate if too long for button (max 64 chars for Telegram)
			if len(displayName) > 35 {
				displayName = displayName[:32] + "..."
			}

			button := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("(Delete) %s - ±%.1f%%", displayName, alert.ThresholdPct),
				fmt.Sprintf("%s_%d", CallbackDeleteAlert, alert.ID),
			)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
		}
	}

	// Add back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Back to Menu", "back_to_menu"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildConfirmDeleteMenu creates a confirmation menu for alert deletion
func BuildConfirmDeleteMenu(alertID int64) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Yes, Delete", fmt.Sprintf("%s_%d", CallbackConfirmDelete, alertID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", CallbackCancelDelete),
		),
	)
}

// BuildBackButton creates a simple back button
func BuildBackButton() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back to Menu", "back_to_menu"),
		),
	)
}

// BuildMarketSelectionMenu creates a menu with featured markets
func BuildMarketSelectionMenu(featuredMarkets []FeaturedMarket) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Add featured markets
	for _, market := range featuredMarkets {
		// Truncate long market names for button display
		displayName := market.Name
		if len(displayName) > 45 {
			displayName = displayName[:42] + "..."
		}

		button := tgbotapi.NewInlineKeyboardButtonData(
			displayName,
			fmt.Sprintf("%s_%s", CallbackSelectMarket, market.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Add custom market ID button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📝 Enter Custom Market ID", CallbackCustomMarket),
	))

	// Add back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Back to Menu", "back_to_menu"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildThresholdSelectionMenu creates a menu with common threshold values
func BuildThresholdSelectionMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1%", fmt.Sprintf("%s_1", CallbackSelectThreshold)),
			tgbotapi.NewInlineKeyboardButtonData("2%", fmt.Sprintf("%s_2", CallbackSelectThreshold)),
			tgbotapi.NewInlineKeyboardButtonData("5%", fmt.Sprintf("%s_5", CallbackSelectThreshold)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("10%", fmt.Sprintf("%s_10", CallbackSelectThreshold)),
			tgbotapi.NewInlineKeyboardButtonData("25%", fmt.Sprintf("%s_25", CallbackSelectThreshold)),
			tgbotapi.NewInlineKeyboardButtonData("50%", fmt.Sprintf("%s_50", CallbackSelectThreshold)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Enter Custom Value", CallbackCustomThreshold),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back to Menu", "back_to_menu"),
		),
	)
}
