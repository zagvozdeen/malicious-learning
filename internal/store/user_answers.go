package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/zagvozdeen/malicious-learning/internal/store/enum"
)

type UserAnswer struct {
	ID            int                   `json:"id"`
	UUID          string                `json:"uuid"`
	CardID        int                   `json:"card_id"`
	TestSessionID int                   `json:"test_session_id"`
	Status        enum.UserAnswerStatus `json:"status"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

type FullUserAnswer struct {
	UserAnswer

	UID        int    `json:"uid"`
	Answer     string `json:"answer"`
	Question   string `json:"question"`
	ModuleID   int    `json:"module_id"`
	ModuleName string `json:"module_name"`
}

type TestSessionSummary struct {
	UUID               string    `json:"uuid"`
	IsActive           bool      `json:"is_active"`
	IsShuffled         bool      `json:"is_shuffled"`
	ModuleIDs          []int     `json:"module_ids"`
	HasRecommendations bool      `json:"has_recommendations"`
	CountNull          int       `json:"count_null"`
	CountRemember      int       `json:"count_remember"`
	CountForget        int       `json:"count_forget"`
	CreatedAt          time.Time `json:"created_at"`
	CourseName         string    `json:"course_name"`
}

func (s *Store) GetUserAnswersByTestSessionID(ctx context.Context, id int) ([]FullUserAnswer, error) {
	rows, err := s.querier(ctx).Query(ctx, `
		SELECT ua.id, ua.uuid, ua.card_id, ua.test_session_id, ua.status, ua.created_at, ua.updated_at, c.uid, c.answer, c.question, c.module_id, m.name
		FROM user_answers ua
		JOIN cards c ON c.id = ua.card_id
		JOIN modules m ON m.id = c.module_id
		WHERE ua.test_session_id = $1
		ORDER BY ua.id
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []FullUserAnswer
	for rows.Next() {
		var answer FullUserAnswer
		err = rows.Scan(
			&answer.ID,
			&answer.UUID,
			&answer.CardID,
			&answer.TestSessionID,
			&answer.Status,
			&answer.CreatedAt,
			&answer.UpdatedAt,
			&answer.UID,
			&answer.Answer,
			&answer.Question,
			&answer.ModuleID,
			&answer.ModuleName,
		)
		if err != nil {
			return nil, err
		}
		answers = append(answers, answer)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return answers, nil
}

func (s *Store) GetTestSessions(ctx context.Context, userID int) ([]TestSessionSummary, error) {
	rows, err := s.querier(ctx).Query(ctx, `
		SELECT
			ts.uuid,
			ts.is_active,
			ts.is_shuffled,
			ts.module_ids,
			ts.recommendations IS NOT NULL,
			count(ua.id) FILTER ( WHERE ua.status = 'null' ) count_null,
			count(ua.id) FILTER ( WHERE ua.status = 'remember' ) count_remember,
			count(ua.id) FILTER ( WHERE ua.status = 'forgot' ) count_forget,
			ts.created_at,
			co.name
		FROM test_sessions ts
		JOIN courses co ON co.id = ts.course_id
		LEFT JOIN user_answers ua ON ua.test_session_id = ts.id
		WHERE ts.user_id = $1
		GROUP BY ts.id, ts.uuid, ts.is_active, ts.is_shuffled, ts.module_ids, ts.recommendations IS NOT NULL, ts.created_at, co.name
		ORDER BY ts.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]TestSessionSummary, 0)
	for rows.Next() {
		var session TestSessionSummary
		err = rows.Scan(
			&session.UUID,
			&session.IsActive,
			&session.IsShuffled,
			&session.ModuleIDs,
			&session.HasRecommendations,
			&session.CountNull,
			&session.CountRemember,
			&session.CountForget,
			&session.CreatedAt,
			&session.CourseName,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *Store) GetUserAnswerByUUID(ctx context.Context, uuid string) (*UserAnswer, error) {
	answer := &UserAnswer{}
	err := s.querier(ctx).QueryRow(
		ctx,
		"SELECT id, uuid, card_id, test_session_id, status, created_at, updated_at FROM user_answers WHERE uuid = $1",
		uuid,
	).Scan(
		&answer.ID,
		&answer.UUID,
		&answer.CardID,
		&answer.TestSessionID,
		&answer.Status,
		&answer.CreatedAt,
		&answer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func (s *Store) UpdateUserAnswer(ctx context.Context, ua *UserAnswer) error {
	tag, err := s.querier(ctx).Exec(
		ctx,
		"UPDATE user_answers SET status = $1, updated_at = $2 WHERE id = $3",
		ua.Status, ua.UpdatedAt, ua.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
