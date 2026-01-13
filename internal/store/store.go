package store

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zagvozdeen/malicious-learning/internal/config"
)

type Storage interface {
}

type Store struct {
	cfg  *config.Config
	log  *slog.Logger
	pool *pgxpool.Pool
}

var _ Storage = (*Store)(nil)

func New(cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool) *Store {
	return &Store{cfg: cfg, log: log, pool: pool}
}
