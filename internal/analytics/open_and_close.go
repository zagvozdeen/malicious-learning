package analytics

import (
	"encoding/json/v2"
	"log/slog"
	"os"
)

func (a *Analytics) open() {
	b, err := os.ReadFile("metrics.json")
	if err != nil {
		a.log.Warn("Failed to open and read metrics file", slog.Any("err", err))
		return
	}
	err = json.Unmarshal(b, &a.snapshot)
	if err != nil {
		a.log.Warn("Failed to parse metrics file", slog.Any("err", err))
	}
}

func (a *Analytics) close() {
	b, err := json.Marshal(a.snapshot)
	if err != nil {
		a.log.Warn("Failed to marshal metrics to file", slog.Any("err", err))
		return
	}
	err = os.WriteFile("metrics.json", b, 0o644)
	if err != nil {
		a.log.Warn("Failed to open metrics file", slog.Any("err", err))
		return
	}
}
