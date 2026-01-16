package logger

import (
	"log/slog"
	"os"

	"github.com/zagvozdeen/malicious-learning/internal/config"
)

func New(cfg *config.Config) (*slog.Logger, func()) {
	if cfg.IsProduction {
		file, err := os.OpenFile("ml.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			slog.Error("Failed to open log file", slog.Any("err", err))
			os.Exit(1)
		}
		log := slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug}))
		return log, func() {
			closeErr := file.Close()
			if closeErr != nil {
				slog.Error("Failed to close file", slog.Any("err", closeErr))
			}
			slog.Info("Logs file has been closed")
		}
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})), func() {}
}
