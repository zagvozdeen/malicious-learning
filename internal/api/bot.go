package api

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (s *Service) startBot() error {
	b, err := bot.New(s.cfg.TelegramBotToken, bot.WithDefaultHandler(s.defaultHandler))
	if err != nil {
		return err
	}
	b.Start(s.ctx)
	return nil
}

func (s *Service) defaultHandler(ctx context.Context, bot *bot.Bot, update *models.Update) {
	// write code here
}
