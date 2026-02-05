package api

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/zagvozdeen/malicious-learning/internal/analytics"
)

func (s *Service) startSendingMetrics() error {
	if s.cfg.TelegramBotGroup == 0 {
		return errors.New("telegram bot group must be set")
	}
	if !s.cfg.TelegramBotEnabled {
		return errors.New("telegram bot disabled")
	}
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	var old analytics.Snapshot
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-s.botStarted:
			snap := s.metrics.Snapshot()
			s.sendMetrics(snap, old)
			old = snap
		case <-ticker.C:
			snap := s.metrics.Snapshot()
			s.sendMetrics(snap, old)
			old = snap
		}
	}
}

func (s *Service) sendMetrics(new analytics.Snapshot, old analytics.Snapshot) {
	if new.Equal(old) {
		s.log.Info("Skip sending stats due to there is no changes")
		return
	}
	var lines = []string{
		"*Вот сводка за последний час\\:*\n",
		fmt.Sprintf("– *Создано пользователей\\:* %s", compare(new.AppUsersCreatedCount, old.AppUsersCreatedCount)),
		fmt.Sprintf("– *Отправлено не\\-сообщений\\:* %s", compare(new.AppNotMessageUpdateCount, old.AppNotMessageUpdateCount)),
		fmt.Sprintf("– *Создано рекомендаций\\:* %s", compare(new.AppGeneratedRecommendationsCount, old.AppGeneratedRecommendationsCount)),
		fmt.Sprintf("– *Обновлено ответов\\:* %s", compare(new.AppUpdatedUserAnswersCount, old.AppUpdatedUserAnswersCount)),
		fmt.Sprintf("– *Начато тестовых сессий\\:* %s", compare(new.AppCreatedTestSessionsCount, old.AppCreatedTestSessionsCount)),
		"", "*Статистика по страницам\\:*\n",
	}
	lines = append(lines, getAppResponsesTotalDiff(new, old)...)
	_, err := s.bot.SendMessage(s.ctx, &bot.SendMessageParams{
		ChatID:    s.cfg.TelegramBotGroup,
		Text:      strings.Join(lines, "\n"),
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		s.log.Error("Failed to send message to group", slog.Any("err", err))
	}
}

func compare(n int64, o int64) string {
	if diff := n - o; diff != 0 {
		return fmt.Sprintf("%d \\(\\+%d\\)", n, diff)
	}
	return strconv.Itoa(int(n))
}

func getAppResponsesTotalDiff(new analytics.Snapshot, old analytics.Snapshot) []string {
	r := make([]string, 0, len(new.AppResponsesTotal))
	for path, count := range new.AppResponsesTotal {
		if diff := count - old.AppResponsesTotal[path]; diff != 0 {
			r = append(r, fmt.Sprintf("– *%s\\:* %d \\(\\+%d\\)", bot.EscapeMarkdown(path), count, diff))
		} else {
			r = append(r, fmt.Sprintf("– *%s\\:* %d", bot.EscapeMarkdown(path), count))
		}
	}
	return r
}
