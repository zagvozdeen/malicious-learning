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
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	<-s.botStarted
	old := s.sendMetrics(s.metrics.Clone())
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-ticker.C:
			old = s.sendMetrics(old)
		}
	}
}

func (s *Service) sendMetrics(old analytics.Metrics) analytics.Metrics {
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
	_, err := s.bot.SendMessage(s.ctx, &bot.SendMessageParams{
		ChatID:    s.cfg.TelegramBotGroup,
		Text:      strings.Join(lines, "\n"),
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		s.log.Error("Failed to send message to group", slog.Any("err", err))
	}
	return s.metrics.Clone()
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
	for path, codes := range new.GetAppResponsesTotal() {
		for code, counter := range codes {
			oldCodes, ok := o[path]
			if !ok {
				oldCodes = codes
			}
			oldCounter, ok := oldCodes[code]
			if !ok {
				oldCounter = counter
			}
			if oldCounter == counter {
				r = append(r, fmt.Sprintf("– *%s \\[%d\\]:* %d", bot.EscapeMarkdown(path), code, counter))
			} else {
				r = append(r, fmt.Sprintf("– *%s \\[%d\\]:* %d \\(\\+%d\\)", bot.EscapeMarkdown(path), code, counter, counter-oldCounter))
			}
		}
	}
	return r
}
