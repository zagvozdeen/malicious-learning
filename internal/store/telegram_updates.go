package store

import (
	"context"
	"encoding/json"
	"time"
)

type TelegramUpdate struct {
	ID     int64
	Update json.RawMessage
	Date   time.Time
}

func (s *Store) CreateTelegramUpdate(ctx context.Context, update *TelegramUpdate) (err error) {
	_, err = s.querier(ctx).Exec(
		ctx,
		"INSERT INTO telegram_updates (id, update, date) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING",
		update.ID, update.Update, update.Date,
	)
	return err
}
