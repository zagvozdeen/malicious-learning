package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/zagvozdeen/malicious-learning/internal/config"
)

//go:embed migrations
var fs embed.FS

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) *pgxpool.Pool {
	conn, err := connect(ctx, cfg, log)
	if err != nil {
		log.Error("Failed to create db connection", slog.Any("err", err))
		os.Exit(1)
	}
	return conn
}

func connect(ctx context.Context, cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, &pgxpool.Config{
		ConnConfig: &pgx.ConnConfig{
			Config: pgconn.Config{
				Host:     cfg.DBHost,
				Port:     cfg.DBPort,
				Database: cfg.DBDatabase,
				User:     cfg.DBUsername,
				Password: cfg.DBPassword,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}
	err = migrate(ctx, cfg, &gooseLogger{log: log}, pool)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

type gooseLogger struct {
	log *slog.Logger
}

func (l *gooseLogger) Fatalf(format string, v ...any) {
	l.log.Error(fmt.Sprintf(format, v...))
}

func (l *gooseLogger) Printf(format string, v ...any) {
	l.log.Info(fmt.Sprintf(format, v...))
}

func migrate(ctx context.Context, cfg *config.Config, log goose.Logger, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	err := db.PingContext(ctx)
	if err != nil {
		return err
	}

	goose.SetBaseFS(fs)
	goose.SetLogger(log)

	err = goose.SetDialect("clickhouse")
	if err != nil {
		return err
	}

	if cfg.DBDownMigrations {
		err = goose.DownContext(ctx, db, "migrations")
		if err != nil {
			return err
		}
	}

	err = goose.UpContext(ctx, db, "migrations")
	if err != nil {
		return err
	}

	return nil
}
