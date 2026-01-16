package store

import "context"

func (s *Store) GetTestSessionByUUID(ctx context.Context, uuid string) (*TestSession, error) {
	session := &TestSession{}
	err := s.pool.QueryRow(
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
