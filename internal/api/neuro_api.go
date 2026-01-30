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
	"github.com/zagvozdeen/malicious-learning/internal/store/enum"
)

const channelSize = 100

func (s *Service) getUserRecommendationsByTestSessionID(user *store.User, id int) error {
	if _, ok := s.processingTS.Load(user.ID); ok {
		return fmt.Errorf("processing ts %d already exists", id)
	}
	ch := make(chan []byte, channelSize)
	defer close(ch)
	s.processingTS.Store(user.ID, ch)
	defer s.processingTS.Delete(user.ID)

	ch <- []byte("<start>")
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
		if answer.Status == enum.UserAnswerStatusForgot || answer.Status == enum.UserAnswerStatusRemember {
			msgs = append(msgs, fmt.Sprintf(
				"%d. %s. %s.",
				answer.UID, answer.Status.Condition(), strings.TrimSpace(answer.Question),
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
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(strings.Join(msgs, "\n")),
		},
		Model: openai.ChatModelGPT5Mini,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(true),
		},
	})

	acc := openai.ChatCompletionAccumulator{}
	var content strings.Builder
	var counter int
	for stream.Next() {
		chunk := stream.Current()
		if !acc.AddChunk(chunk) {
			return errors.New("failed to accumulate chat model stream")
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		content.WriteString(delta)
		if counter < channelSize-3 {
			ch <- []byte(strings.ReplaceAll(content.String(), "\n", "<br>"))
			counter++
		}
	}
	err = stream.Err()
	if err != nil {
		return fmt.Errorf("failed to stream chat model: %w", err)
	}
	uid, err := uuid.NewV7()
	if err != nil {
		return err
	}
	err = s.store.CreateChatCompletions(ctx, &store.ChatCompletions{
		UUID:             uid.String(),
		TestSessionID:    id,
		Model:            acc.Model,
		CompletionTokens: acc.Usage.CompletionTokens,
		PromptTokens:     acc.Usage.PromptTokens,
		TotalTokens:      acc.Usage.TotalTokens,
		Date:             acc.Created,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to create chat model: %w", err)
	}
	if len(acc.Choices) == 0 {
		return errors.New("empty chat model choices")
	}
	finalRecommendations := strings.ReplaceAll(content.String(), "\n", "<br>")
	ch <- []byte(finalRecommendations)
	ts.Recommendations = null.WrapString(finalRecommendations)
	ts.UpdatedAt = time.Now()
	err = s.store.UpdateTestSession(ctx, ts)
	if err != nil {
		return err
	}
	s.store.Commit(ctx)
	ch <- []byte("</start>")
	s.metrics.AppGeneratedRecommendationsCountInc()
	return nil
}
