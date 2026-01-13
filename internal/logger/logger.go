package logger

import (
	"log/slog"
	"os"

	"github.com/zagvozdeen/malicious-learning/internal/config"
)

func New(config *config.Config) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
}
