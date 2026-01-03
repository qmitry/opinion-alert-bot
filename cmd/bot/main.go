package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/qmitry/opinion-alert-bot/internal/api"
	"github.com/qmitry/opinion-alert-bot/internal/config"
	"github.com/qmitry/opinion-alert-bot/internal/monitor"
	"github.com/qmitry/opinion-alert-bot/internal/storage"
	"github.com/qmitry/opinion-alert-bot/internal/telegram"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	log.Info("Starting Opinion Alert Bot...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level
	level, err := logrus.ParseLevel(cfg.App.LogLevel)
	if err != nil {
		log.Warnf("Invalid log level '%s', using 'info'", cfg.App.LogLevel)
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// Initialize database
	log.Info("Connecting to database...")
	db, err := storage.NewStorage(cfg.Database.ConnectionString(), log)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Opinion API client
	log.Info("Initializing Opinion API client...")
	apiClient := api.NewClient(cfg.OpinionAPI.APIKey, cfg.OpinionAPI.BaseURL, log)

	// Initialize Telegram bot
	log.Info("Initializing Telegram bot...")
	bot, err := telegram.NewBot(cfg.Telegram.Token, db, apiClient, log)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	// Initialize monitor
	log.Info("Initializing market monitor...")
	mon := monitor.NewMonitor(db, apiClient, bot, cfg.App.PollInterval, log)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start Telegram bot in goroutine
	botErrChan := make(chan error, 1)
	go func() {
		log.Info("Starting Telegram bot...")
		if err := bot.Start(ctx); err != nil {
			botErrChan <- err
		}
	}()

	// Start monitor in goroutine
	monitorErrChan := make(chan error, 1)
	go func() {
		log.Info("Starting market monitor...")
		if err := mon.Start(ctx); err != nil {
			monitorErrChan <- err
		}
	}()

	log.Info("Opinion Alert Bot is now running. Press Ctrl+C to exit.")

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Infof("Received signal: %v. Shutting down gracefully...", sig)
		cancel()
	case err := <-botErrChan:
		log.Errorf("Telegram bot error: %v", err)
		cancel()
	case err := <-monitorErrChan:
		log.Errorf("Monitor error: %v", err)
		cancel()
	}

	log.Info("Opinion Alert Bot stopped.")
}
