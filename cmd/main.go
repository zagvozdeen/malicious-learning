package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/zagvozdeen/malicious-learning/internal/analytics"
	"github.com/zagvozdeen/malicious-learning/internal/api"
	"github.com/zagvozdeen/malicious-learning/internal/config"
	"github.com/zagvozdeen/malicious-learning/internal/db"
	"github.com/zagvozdeen/malicious-learning/internal/logger"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.New()
	log, stop := logger.New(cfg)
	defer stop()
	metrics, enough := analytics.New(log)
	defer enough()
	pool := db.New(ctx, cfg, log)
	defer pool.Close()
	storage := store.New(cfg, log, pool)

	api.New(ctx, cfg, log, storage, metrics).Run()
}
