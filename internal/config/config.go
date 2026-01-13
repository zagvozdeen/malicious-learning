package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost           string
	DBPort           uint16
	DBDatabase       string
	DBUsername       string
	DBPassword       string
	DBDownMigrations bool
	APIHost          string
	APIPort          string
}

func New() *Config {
	return &Config{
		DBHost:           os.Getenv("DB_HOST"),
		DBPort:           uint16(getEnvInt("DB_PORT", 5432)),
		DBDatabase:       os.Getenv("DB_DATABASE"),
		DBUsername:       os.Getenv("DB_USERNAME"),
		DBPassword:       os.Getenv("DB_PASSWORD"),
		DBDownMigrations: false,
		APIHost:          "127.0.0.1",
		APIPort:          "8081",
	}
}

func getEnvInt(key string, fallback int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return v
}
