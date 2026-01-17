package store

import (
	"context"
	"time"

	"github.com/zagvozdeen/malicious-learning/internal/db/null"
)

type TestSession struct {
	ID              int         `json:"id"`
	UUID            string      `json:"uuid"`
	UserID          int         `json:"user_id"`
	ModuleIDs       []int       `json:"module_ids"`
	IsShuffled      bool        `json:"is_shuffled"`
	IsActive        bool        `json:"is_active"`
	Recommendations null.String `json:"recommendations"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

func (s *Store) GetTestSessionByUUID(ctx context.Context, uuid string) (*TestSession, error) {
	session := &TestSession{}
	err := s.querier(ctx).QueryRow(
		ctx,
		"SELECT id, uuid, user_id, module_ids, is_shuffled, is_active, recommendations, created_at, updated_at FROM test_sessions WHERE uuid = $1",
		uuid,
	).Scan(
		&session.ID,
		&session.UUID,
		&session.UserID,
		&session.ModuleIDs,
		&session.IsShuffled,
		&session.IsActive,
		&session.Recommendations,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetTestSessionByID кроме получения TestSession ещё считает оставшиеся вопросы.
func (s *Store) GetTestSessionByID(ctx context.Context, id int) (*TestSession, int, error) {
	session, n := &TestSession{}, 0
	err := s.querier(ctx).QueryRow(
		ctx,
		"SELECT id, uuid, user_id, module_ids, is_shuffled, is_active, recommendations, created_at, updated_at, (SELECT COUNT(*) FROM user_answers WHERE test_session_id = $1 AND status = $2) FROM test_sessions WHERE id = $1",
		id, UserAnswerStatusNull,
	).Scan(
		&session.ID,
		&session.UUID,
		&session.UserID,
		&session.ModuleIDs,
		&session.IsShuffled,
		&session.IsActive,
		&session.Recommendations,
		&session.CreatedAt,
		&session.UpdatedAt,
		&n,
	)
	if err != nil {
		return nil, 0, err
	}
	return session, n, nil
}

func (s *Store) CreateTestSession(ctx context.Context, session *TestSession, answers []UserAnswer) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(
		ctx,
		"INSERT INTO test_sessions (uuid, user_id, module_ids, is_shuffled, is_active, recommendations, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		session.UUID,
		session.UserID,
		session.ModuleIDs,
		session.IsShuffled,
		session.IsActive,
		session.Recommendations,
		session.CreatedAt,
		session.UpdatedAt,
	).Scan(&session.ID)
	if err != nil {
		return err
	}

	for _, answer := range answers {
		_, err = tx.Exec(
			ctx,
			"INSERT INTO user_answers (uuid, card_id, test_session_id, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
			answer.UUID,
			answer.CardID,
			session.ID,
			answer.Status,
			answer.CreatedAt,
			answer.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Store) UpdateTestSession(ctx context.Context, session *TestSession) error {
	_, err := s.querier(ctx).Exec(
		ctx,
		"UPDATE test_sessions SET is_active = $1, recommendations = $2, updated_at = $3 WHERE id = $4",
		session.IsActive,
		session.Recommendations,
		session.UpdatedAt,
		session.ID,
	)
	return err
}
