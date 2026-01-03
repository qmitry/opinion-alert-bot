package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Callback data constants
const (
	CallbackCreateAlert  = "create_alert"
	CallbackMyAlerts     = "my_alerts"
	CallbackMyMarkets    = "my_markets"
	CallbackHelp         = "help"
	CallbackDeleteAlert  = "delete_alert"
	CallbackConfirmDelete = "confirm_delete"
	CallbackCancelDelete = "cancel_delete"
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
			button := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("Market #%s - ±%.1f%% (Delete)", marketID, alert.ThresholdPct),
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
