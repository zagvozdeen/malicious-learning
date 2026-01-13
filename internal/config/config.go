package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppSecret          string
	IsProduction       bool
	DBHost             string
	DBPort             string
	DBDatabase         string
	DBUsername         string
	DBPassword         string
	DBDownMigrations   bool
	APIHost            string
	APIPort            string
	TelegramBotToken   string
	TelegramBotEnabled bool
	RootUserName       string
	RootUserPassword   string
}

func New() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Failed to load .env file", slog.Any("err", err))
	}
	return &Config{
		AppSecret:          os.Getenv("APP_SECRET"),
		IsProduction:       false,
		DBHost:             os.Getenv("DB_HOST"),
		DBPort:             os.Getenv("DB_PORT"),
		DBDatabase:         os.Getenv("DB_DATABASE"),
		DBUsername:         os.Getenv("DB_USERNAME"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBDownMigrations:   false,
		APIHost:            "127.0.0.1",
		APIPort:            "8081",
		TelegramBotToken:   os.Getenv("TG_BOT_TOKEN"),
		TelegramBotEnabled: os.Getenv("TG_BOT_ENABLED") == "true",
		RootUserName:       "root",
		RootUserPassword:   "password",
	}
}
