package api

import (
	"context"
	"encoding/json/v2"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	tgbotmodels "github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zagvozdeen/malicious-learning/internal/db/null"
	storemodels "github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) startBot() error {
	if !s.cfg.TelegramBotEnabled {
		s.log.Info("Telegram bot disabled")
		return nil
	}
	var err error
	s.bot, err = bot.New(s.cfg.TelegramBotToken, bot.WithDefaultHandler(s.defaultHandler))
	if err != nil {
		return err
	}
	s.botStarted <- struct{}{}
	s.bot.Start(s.ctx)
	return nil
}

func (s *Service) defaultHandler(ctx context.Context, b *bot.Bot, update *tgbotmodels.Update) {
	if update == nil {
		return
	}

	if err := s.storeTelegramUpdate(ctx, update); err != nil {
		s.log.Warn("Failed to store telegram update", slog.Any("err", err))
	}

	if tgUser := extractTelegramUser(update); tgUser != nil {
		if err := s.ensureTelegramUser(ctx, tgUser); err != nil {
			s.log.Warn("Failed to ensure telegram user", slog.Any("err", err))
		}
	}

	if update.Message == nil {
		s.metrics.AppNotMessageUpdateCountInc()
		return
	}

	text := strings.TrimSpace(update.Message.Text)
	reply := []string{"Бот не поддерживает никаких команд, весь функционал находится в мини\\-приложении"}
	if text == "/start" {
		reply = []string{
			"Добро пожаловать в бот *Malicious Learning*\\!",
			"",
			"С помощью этого бота ты можешь подготовиться к экзамену по машинному обучению\\. Внутри MiniApp ты найдёшь карточки с вопросами и ответами\\. А также у тебя будет персональная статистика, рассчитанная из ответов:",
			"",
			"\\- жми «Вспомнил» если знаешь ответ",
			"\\- жми «Забыл» если не знаешь ответа",
			"",
			"Весь функционал находится в мини\\-приложении, открывай и готовься\\!",
			"",
			"Ещё сомневаешься или хочешь улучшить проект? [Код приложения](https://github.com/zagvozdeen/malicious-learning) публичный, доступен каждому\\. А если хочешь помочь улучшить ответы, то внутри есть инструкция, как это сделать, или можешь просто написать мне в личку\\.",
			"",
			"_[Связь с автором](https://t.me/denchik1170)_",
		}
	}

	disabledPreviewOptions := true
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		ParseMode: tgbotmodels.ParseModeMarkdown,
		Text:      strings.Join(reply, "\n"),
		LinkPreviewOptions: &tgbotmodels.LinkPreviewOptions{
			IsDisabled: &disabledPreviewOptions,
		},
	})
	if err != nil {
		s.log.Warn("Failed to send telegram reply", slog.Any("err", err))
	}
}

func (s *Service) storeTelegramUpdate(ctx context.Context, update *tgbotmodels.Update) error {
	payload, err := json.Marshal(update)
	if err != nil {
		return err
	}
	return s.store.CreateTelegramUpdate(ctx, &storemodels.TelegramUpdate{
		ID:     update.ID,
		Update: payload,
		Date:   updateTimestamp(update),
	})
}

func (s *Service) ensureTelegramUser(ctx context.Context, tgUser *tgbotmodels.User) error {
	_, err := s.store.GetUserByTID(ctx, tgUser.ID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return err
	}

	err = s.store.CreateUser(ctx, &storemodels.User{
		TID:       null.WrapInt(int(tgUser.ID)),
		UUID:      uid.String(),
		FirstName: strings.TrimSpace(tgUser.FirstName),
		LastName:  null.WrapString(strings.TrimSpace(tgUser.LastName)),
		Username:  null.WrapString(strings.TrimSpace(tgUser.Username)),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return err
	}
	s.metrics.AppUsersCreatedCountInc()
	return nil
}

func extractTelegramUser(update *tgbotmodels.Update) *tgbotmodels.User {
	switch {
	case update.Message != nil && update.Message.From != nil:
		return update.Message.From
	case update.EditedMessage != nil && update.EditedMessage.From != nil:
		return update.EditedMessage.From
	case update.ChannelPost != nil && update.ChannelPost.From != nil:
		return update.ChannelPost.From
	case update.EditedChannelPost != nil && update.EditedChannelPost.From != nil:
		return update.EditedChannelPost.From
	case update.BusinessMessage != nil && update.BusinessMessage.From != nil:
		return update.BusinessMessage.From
	case update.EditedBusinessMessage != nil && update.EditedBusinessMessage.From != nil:
		return update.EditedBusinessMessage.From
	case update.CallbackQuery != nil:
		return &update.CallbackQuery.From
	case update.InlineQuery != nil && update.InlineQuery.From != nil:
		return update.InlineQuery.From
	case update.ChosenInlineResult != nil:
		return &update.ChosenInlineResult.From
	case update.ShippingQuery != nil && update.ShippingQuery.From != nil:
		return update.ShippingQuery.From
	case update.PreCheckoutQuery != nil && update.PreCheckoutQuery.From != nil:
		return update.PreCheckoutQuery.From
	case update.PollAnswer != nil && update.PollAnswer.User != nil:
		return update.PollAnswer.User
	case update.MessageReaction != nil && update.MessageReaction.User != nil:
		return update.MessageReaction.User
	case update.MyChatMember != nil:
		return &update.MyChatMember.From
	case update.ChatMember != nil:
		return &update.ChatMember.From
	case update.ChatJoinRequest != nil:
		return &update.ChatJoinRequest.From
	default:
		return nil
	}
}

func updateTimestamp(update *tgbotmodels.Update) time.Time {
	switch {
	case update.Message != nil && update.Message.Date != 0:
		return time.Unix(int64(update.Message.Date), 0).UTC()
	case update.EditedMessage != nil && update.EditedMessage.Date != 0:
		return time.Unix(int64(update.EditedMessage.Date), 0).UTC()
	case update.ChannelPost != nil && update.ChannelPost.Date != 0:
		return time.Unix(int64(update.ChannelPost.Date), 0).UTC()
	case update.EditedChannelPost != nil && update.EditedChannelPost.Date != 0:
		return time.Unix(int64(update.EditedChannelPost.Date), 0).UTC()
	case update.BusinessMessage != nil && update.BusinessMessage.Date != 0:
		return time.Unix(int64(update.BusinessMessage.Date), 0).UTC()
	case update.EditedBusinessMessage != nil && update.EditedBusinessMessage.Date != 0:
		return time.Unix(int64(update.EditedBusinessMessage.Date), 0).UTC()
	case update.MessageReaction != nil && update.MessageReaction.Date != 0:
		return time.Unix(int64(update.MessageReaction.Date), 0).UTC()
	case update.MessageReactionCount != nil && update.MessageReactionCount.Date != 0:
		return time.Unix(int64(update.MessageReactionCount.Date), 0).UTC()
	default:
		return time.Now().UTC()
	}
}
