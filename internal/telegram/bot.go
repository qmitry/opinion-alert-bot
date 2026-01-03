package telegram

import (
	"context"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qmitry/opinion-alert-bot/internal/api"
	"github.com/qmitry/opinion-alert-bot/internal/storage"
	"github.com/sirupsen/logrus"
)

// Bot represents the Telegram bot
type Bot struct {
	api       *tgbotapi.BotAPI
	storage   *storage.Storage
	apiClient *api.Client
	log       *logrus.Logger

	// User conversation states
	userStates map[int64]*UserState
	stateMu    sync.RWMutex
}

// UserState tracks the current conversation state for a user
type UserState struct {
	Step      string // Current step in the conversation
	Data      map[string]interface{}
	MarketID  string
	Threshold float64
}

// NewBot creates a new Telegram bot instance
func NewBot(token string, storage *storage.Storage, apiClient *api.Client, log *logrus.Logger) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Infof("Authorized on account %s", botAPI.Self.UserName)

	return &Bot{
		api:        botAPI,
		storage:    storage,
		apiClient:  apiClient,
		log:        log,
		userStates: make(map[int64]*UserState),
	}, nil
}

// Start begins polling for updates
func (b *Bot) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	b.log.Info("Telegram bot started, waiting for updates...")

	for {
		select {
		case <-ctx.Done():
			b.log.Info("Stopping Telegram bot...")
			b.api.StopReceivingUpdates()
			return nil
		case update := <-updates:
			go b.handleUpdate(update)
		}
	}
}

// handleUpdate processes incoming updates
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	ctx := context.Background()

	// Handle callback queries (button clicks)
	if update.CallbackQuery != nil {
		b.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}

	// Handle regular messages
	if update.Message != nil {
		b.handleMessage(ctx, update.Message)
		return
	}
}

// getUserState retrieves or creates a user state
func (b *Bot) getUserState(userID int64) *UserState {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()

	if state, exists := b.userStates[userID]; exists {
		return state
	}

	state := &UserState{
		Step: "",
		Data: make(map[string]interface{}),
	}
	b.userStates[userID] = state
	return state
}

// clearUserState clears a user's conversation state
func (b *Bot) clearUserState(userID int64) {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	delete(b.userStates, userID)
}

// SendMessage sends a message to a user
func (b *Bot) SendMessage(chatID int64, text string, keyboard interface{}) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}

	_, err := b.api.Send(msg)
	return err
}

// SendAlertNotification sends a price alert notification to a user
func (b *Bot) SendAlertNotification(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	_, err := b.api.Send(msg)
	if err != nil {
		b.log.Errorf("Failed to send alert notification to chat %d: %v", chatID, err)
		return err
	}

	b.log.Debugf("Sent alert notification to chat %d", chatID)
	return nil
}
