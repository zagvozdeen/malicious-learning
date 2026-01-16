package store

import "context"

func (s *Store) CreateTelegramUpdate(ctx context.Context, update *TelegramUpdate) (err error) {
	_, err = s.pool.Exec(
		ctx,
		"INSERT INTO telegram_updates (id, update, date) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING",
		update.ID, update.Update, update.Date,
	)
	return err
}
