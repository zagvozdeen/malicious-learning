package store

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (s *Store) querier(ctx context.Context) querier {
	if tx, ok := ctx.Value("tx").(pgx.Tx); ok {
		return tx
	}
	return s.pool
}

func (s *Store) Begin(ctx context.Context) (context.Context, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, "tx", tx)
	return ctx, nil
}

func (s *Store) Commit(ctx context.Context) {
	if tx, ok := ctx.Value("tx").(pgx.Tx); ok {
		err := tx.Commit(ctx)
		if err != nil {
			s.log.Error("Failed to commit transaction", slog.Any("err", err))
		}
		return
	}
	s.log.Error("Commit failed: tx not found in context")
}

func (s *Store) Rollback(ctx context.Context) {
	if tx, ok := ctx.Value("tx").(pgx.Tx); ok {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			s.log.Error("Failed to rollback transaction", slog.Any("err", err))
		}
		return
	}
	s.log.Error("Rollback failed: tx not found in context")
}
