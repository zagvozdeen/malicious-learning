package store

import (
	"context"
	"time"
)

type ChatCompletions struct {
	ID               int       `json:"id"`
	UUID             string    `json:"uuid"`
	TestSessionID    int       `json:"test_session_id"`
	Model            string    `json:"model"`
	CompletionTokens int64     `json:"completion_tokens"`
	PromptTokens     int64     `json:"prompt_tokens"`
	TotalTokens      int64     `json:"total_tokens"`
	Date             int64     `json:"date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (s *Store) CreateChatCompletions(ctx context.Context, cc *ChatCompletions) error {
	return s.querier(ctx).QueryRow(
		ctx,
		"INSERT INTO chat_completions (uuid, test_session_id, model, completion_tokens, prompt_tokens, total_tokens, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		cc.UUID,
		cc.TestSessionID,
		cc.Model,
		cc.CompletionTokens,
		cc.PromptTokens,
		cc.TotalTokens,
		cc.Date,
		cc.CreatedAt,
		cc.UpdatedAt,
	).Scan(&cc.ID)
}
