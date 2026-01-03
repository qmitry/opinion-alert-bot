package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	OpinionAPI OpinionAPIConfig
	Telegram   TelegramConfig
	Database   DatabaseConfig
	App        AppConfig
}

// OpinionAPIConfig holds Opinion API configuration
type OpinionAPIConfig struct {
	APIKey  string
	BaseURL string
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	Token string
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// AppConfig holds application-level configuration
type AppConfig struct {
	PollInterval int
	LogLevel     string
	Timezone     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	pollInterval, err := strconv.Atoi(getEnv("POLL_INTERVAL", "60"))
	if err != nil {
		return nil, fmt.Errorf("invalid POLL_INTERVAL: %w", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	cfg := &Config{
		OpinionAPI: OpinionAPIConfig{
			APIKey:  getEnv("OPINION_API_KEY", ""),
			BaseURL: getEnv("OPINION_API_BASE_URL", "https://openapi.opinion.trade"),
		},
		Telegram: TelegramConfig{
			Token: getEnv("TELEGRAM_TOKEN", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			Name:     getEnv("DB_NAME", "opinion_alerts"),
			User:     getEnv("DB_USER", "botuser"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		App: AppConfig{
			PollInterval: pollInterval,
			LogLevel:     getEnv("LOG_LEVEL", "info"),
			Timezone:     getEnv("TZ", "UTC"),
		},
	}

	// Validate required fields
	if cfg.OpinionAPI.APIKey == "" {
		return nil, fmt.Errorf("OPINION_API_KEY is required")
	}
	if cfg.Telegram.Token == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN is required")
	}
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// ConnectionString returns the PostgreSQL connection string
func (d *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}
