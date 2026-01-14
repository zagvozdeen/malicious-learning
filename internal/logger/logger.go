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
			panic(err)
		}
		log := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug}))
		return log, func() {
			closeErr := file.Close()
			if closeErr != nil {
				log.Error("Failed to close file", slog.Any("err", closeErr))
			}
			log.Info("Logs file has been closed")
		}
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})), func() {}
}
