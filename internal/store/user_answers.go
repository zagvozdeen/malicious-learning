package store

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (s *Store) GetUserAnswersByTestSessionID(ctx context.Context, id int) ([]FullUserAnswer, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ua.id, ua.uuid, ua.card_id, ua.test_session_id, ua.status, ua.created_at, ua.updated_at, c.answer, c.question, c.module_id, m.name
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
	rows, err := s.pool.Query(ctx, `
		SELECT
			ts.uuid,
			count(ua.id) FILTER ( WHERE ua.status = 'null' ) count_null,
			count(ua.id) FILTER ( WHERE ua.status = 'remember' ) count_remember,
			count(ua.id) FILTER ( WHERE ua.status = 'forgot' ) count_forget,
			ts.created_at
		FROM test_sessions ts
		LEFT JOIN user_answers ua ON ua.test_session_id = ts.id
		WHERE ts.user_id = $1
		GROUP BY ts.id, ts.uuid, ts.created_at
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
			&session.GroupUUID,
			&session.CountNull,
			&session.CountRemember,
			&session.CountForget,
			&session.CreatedAt,
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
	err := s.pool.QueryRow(ctx, `
		SELECT ua.id, ua.uuid, ua.card_id, ua.test_session_id, ts.user_id, ua.status, ua.created_at, ua.updated_at
		FROM user_answers ua
		LEFT JOIN test_sessions ts ON ts.id = ua.test_session_id
		WHERE ua.uuid = $1
	`, uuid).Scan(
		&answer.ID,
		&answer.UUID,
		&answer.CardID,
		&answer.TestSessionID,
		&answer.UserID,
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
	tag, err := s.pool.Exec(
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
