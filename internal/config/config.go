package config

import (
	"log/slog"
	"os"
	"strconv"

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
	TelegramBotGroup   int
	RootUserName       string
	RootUserPassword   string
	NeuroAPI           string
	NeuroToken         string
	NeuroDebug         bool
}

func New() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Failed to load .env file", slog.Any("err", err))
	}
	return &Config{
		AppSecret:          os.Getenv("APP_SECRET"),
		IsProduction:       os.Getenv("IS_PRODUCTION") == "true",
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
		TelegramBotGroup:   parseInt("TG_BOT_GROUP", 0),
		RootUserName:       os.Getenv("ROOT_USER_NAME"),
		RootUserPassword:   os.Getenv("ROOT_USER_PASSWORD"),
		NeuroAPI:           os.Getenv("NEURO_API"),
		NeuroToken:         os.Getenv("NEURO_TOKEN"),
		NeuroDebug:         false,
	}
}

func parseInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
