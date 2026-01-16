package store

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (s Store) CreateUserAnswers(ctx context.Context, ua []UserAnswer) error {
	if len(ua) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, answer := range ua {
		_, err := tx.Exec(ctx, `
			INSERT INTO user_answers (uuid, group_uuid, card_id, user_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
			answer.UUID,
			answer.GroupUUID,
			answer.CardID,
			answer.UserID,
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

func (s Store) GetUserAnswersByGroupUUID(ctx context.Context, uuid string) ([]FullUserAnswer, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			ua.id,
			ua.uuid,
			ua.group_uuid,
			ua.card_id,
			ua.user_id,
			ua.status,
			ua.created_at,
			ua.updated_at,
			c.answer,
			c.question,
			c.module_id,
			m.name
		FROM user_answers ua
		JOIN cards c ON c.id = ua.card_id
		JOIN modules m ON m.id = c.module_id
		WHERE ua.group_uuid = $1
		ORDER BY ua.id
	`, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	answers := make([]FullUserAnswer, 0)
	for rows.Next() {
		var answer FullUserAnswer
		err = rows.Scan(
			&answer.ID,
			&answer.UUID,
			&answer.GroupUUID,
			&answer.CardID,
			&answer.UserID,
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

func (s Store) GetTestSessions(ctx context.Context, userID int) ([]TestSessionSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			group_uuid,
			count(*) FILTER ( WHERE status = 'null' ) count_null,
			count(*) FILTER ( WHERE status = 'remember' ) count_remember,
			count(*) FILTER ( WHERE status = 'forgot' ) count_forget,
			created_at
		FROM user_answers
		WHERE user_id = $1
		GROUP BY group_uuid, created_at
		ORDER BY created_at DESC
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

func (s Store) GetDistinctUserAnswers(ctx context.Context, userID int) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT group_uuid
		FROM user_answers
		WHERE user_id = $1
		GROUP BY group_uuid
		ORDER BY group_uuid
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groupUUIDs := make([]string, 0)
	for rows.Next() {
		var groupUUID string
		if err := rows.Scan(&groupUUID); err != nil {
			return nil, err
		}
		groupUUIDs = append(groupUUIDs, groupUUID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groupUUIDs, nil
}

func (s Store) GetUserAnswerByUUID(ctx context.Context, uuid string) (*UserAnswer, error) {
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

func (s Store) UpdateUserAnswer(ctx context.Context, ua *UserAnswer) error {
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
