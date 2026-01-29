package api

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/zagvozdeen/malicious-learning/internal/db/null"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getUserRecommendationsByTestSessionID(user *store.User, id int) error {
	if _, ok := s.processingTS.Load(id); ok {
		return fmt.Errorf("processing ts %d already exists", id)
	}
	s.processingTS.Store(id, struct{}{})
	defer s.processingTS.Delete(id)

	s.events <- Event{
		ID:    user.ID,
		Event: "get-recommendations-start",
		Data:  "Program is getting recommendations",
	}
	ctx, err := s.store.Begin(s.ctx)
	if err != nil {
		return err
	}
	defer s.store.Rollback(ctx)
	ts, _, err := s.store.GetTestSessionByID(ctx, id)
	if err != nil {
		return err
	}
	if ts.Recommendations.Valid {
		return errors.New("recommendations exist")
	}
	ua, err := s.store.GetUserAnswersByTestSessionID(ctx, ts.ID)
	if err != nil {
		return err
	}
	if len(ua) == 0 {
		return errors.New("empty user answers")
	}
	var msgs = []string{
		strings.Join([]string{
			"Пользователь готовится к экзамену по машинному обучению.",
			"У него есть карточки, на которые он отвечает «Вспомнил» или «Забыл».",
			"Ниже будет вопрос с карточки и ответ пользователя.",
			"Твоя задача ― дать персонализированные рекомендации исходя из ответов: с чем сложности, что подучить, что повторить.",
			"Пиши кратко и просто, ответ должен уместиться в 2-3 параграфа текста.",
			"Итак, ниже будет номер вопроса, ответ пользователя и сам вопрос в формате HTML.",
		}, ""),
		"```",
	}
	slices.SortFunc(ua, func(a, b store.FullUserAnswer) int {
		return a.UID - b.UID
	})
	for _, answer := range ua {
		if answer.Status == store.UserAnswerStatusForgot || answer.Status == store.UserAnswerStatusRemember {
			var res string
			switch answer.Status {
			case store.UserAnswerStatusForgot:
				res = "Забыл"
			case store.UserAnswerStatusRemember:
				res = "Вспомнил"
			}
			msgs = append(msgs, fmt.Sprintf(
				"%d. %s. %s.",
				answer.UID, res, strings.TrimSpace(answer.Question),
			))
		}
	}
	msgs = append(msgs, "```")
	options := []option.RequestOption{
		option.WithBaseURL(s.cfg.NeuroAPI),
		option.WithAPIKey(s.cfg.NeuroToken),
	}
	if s.cfg.NeuroDebug {
		options = append(options, option.WithDebugLog(slog.NewLogLogger(s.log.Handler(), slog.LevelDebug)))
	}
	client := openai.NewClient(options...)
	res, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(strings.Join(msgs, "\n")),
		},
		Model: openai.ChatModelGPT5Mini,
	})
	if err != nil {
		return fmt.Errorf("failed to create chat model: %w", err)
	}
	uid, err := uuid.NewV7()
	if err != nil {
		return err
	}
	err = s.store.CreateChatCompletions(ctx, &store.ChatCompletions{
		UUID:             uid.String(),
		TestSessionID:    id,
		Model:            res.Model,
		CompletionTokens: res.Usage.CompletionTokens,
		PromptTokens:     res.Usage.PromptTokens,
		TotalTokens:      res.Usage.TotalTokens,
		Date:             res.Created,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to create chat model: %w", err)
	}
	if len(res.Choices) == 0 {
		return errors.New("empty chat model choices")
	}
	ts.Recommendations = null.WrapString(strings.ReplaceAll(res.Choices[0].Message.Content, "\n", "<br>"))
	ts.UpdatedAt = time.Now()
	err = s.store.UpdateTestSession(ctx, ts)
	if err != nil {
		return err
	}
	s.store.Commit(ctx)
	s.events <- Event{
		ID:    user.ID,
		Event: "get-recommendations-end",
		Data:  ts.Recommendations.V,
	}
	s.metrics.AppGeneratedRecommendationsCountInc()
	return nil
}
