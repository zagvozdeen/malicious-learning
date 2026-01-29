package api

import (
	"crypto/sha256"
	"encoding/hex"
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
	var old analytics.Metrics = nil
	var oldHash string
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-s.botStarted:
			old, oldHash = s.sendMetrics(s.metrics.Clone(), oldHash)
		case <-ticker.C:
			old, oldHash = s.sendMetrics(old, oldHash)
		}
	}
}

func (s *Service) sendMetrics(old analytics.Metrics, oldHash string) (analytics.Metrics, string) {
	var lines = []string{
		"*Вот сводка за последний час\\:*", "",
		fmt.Sprintf("– *Создано пользователей\\:* %s", s.compareTwoValues(old.GetAppUsersCreatedCount, s.metrics.GetAppUsersCreatedCount)),
		fmt.Sprintf("– *Отправлено не сообщений\\:* %s", s.compareTwoValues(old.GetAppNotMessageUpdateCount, s.metrics.GetAppNotMessageUpdateCount)),
		fmt.Sprintf("– *Сгенерировано рекомендаций\\:* %s", s.compareTwoValues(old.GetAppGeneratedRecommendationsCount, s.metrics.GetAppGeneratedRecommendationsCount)),
		fmt.Sprintf("– *Обновлено ответов пользователей\\:* %s", s.compareTwoValues(old.GetAppUpdatedUserAnswersCount, s.metrics.GetAppUpdatedUserAnswersCount)),
		fmt.Sprintf("– *Начато тестовых сессий\\:* %s", s.compareTwoValues(old.GetAppCreatedTestSessionsCount, s.metrics.GetAppCreatedTestSessionsCount)),
		"", "**Статистика по страницам\\:**", "",
	}
	lines = append(lines, s.getAppResponsesTotalDiff(old, s.metrics)...)
	text := strings.Join(lines, "\n")
	hasher := sha256.New()
	hasher.Write([]byte(text))
	hash := hex.EncodeToString(hasher.Sum(nil))
	if hash == oldHash {
		s.log.Info("Skip sending stats due to there is no changes")
		return s.metrics.Clone(), hash
	}
	_, err := s.bot.SendMessage(s.ctx, &bot.SendMessageParams{
		ChatID:    s.cfg.TelegramBotGroup,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		s.log.Error("Failed to send message to group", slog.Any("err", err))
	}
	return s.metrics.Clone(), hash
}

func (s *Service) compareTwoValues(old func() int64, new func() int64) string {
	o, n := old(), new()
	if n == o {
		return strconv.Itoa(int(n))
	}
	return fmt.Sprintf("%d \\(\\+%d\\)", n, n-o)
}

func (s *Service) getAppResponsesTotalDiff(old analytics.Metrics, new analytics.Metrics) (r []string) {
	o := old.GetAppResponsesTotal()
	for path, count := range new.GetAppResponsesTotal() {
		if o[path] == count {
			r = append(r, fmt.Sprintf("– *%s:* %d", bot.EscapeMarkdown(path), count))
		} else {
			r = append(r, fmt.Sprintf("– *%s:* %d \\(\\+%d\\)", bot.EscapeMarkdown(path), count, count-o[path]))
		}
	}
	return r
}
