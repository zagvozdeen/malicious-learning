package store

import "context"

func (s *Store) GetUserByTID(ctx context.Context, tid int64) (*User, error) {
	user := &User{}
	err := s.querier(ctx).QueryRow(ctx, `
		SELECT id, tid, uuid, first_name, last_name, username, email, password, created_at, updated_at
		FROM users
		WHERE tid = $1
	`, tid).Scan(
		&user.ID,
		&user.TID,
		&user.UUID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int) (*User, error) {
	user := &User{}
	err := s.querier(ctx).QueryRow(ctx, `
		SELECT id, tid, uuid, first_name, last_name, username, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.TID,
		&user.UUID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) CreateUser(ctx context.Context, user *User) error {
	return s.querier(ctx).QueryRow(ctx, `
		INSERT INTO users (tid, uuid, first_name, last_name, username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`,
		user.TID,
		user.UUID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Email,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	err := s.querier(ctx).QueryRow(
		ctx,
		"SELECT id, tid, uuid, first_name, last_name, username, email, password, created_at, updated_at FROM users WHERE username = $1 LIMIT 1",
		username,
	).Scan(
		&user.ID,
		&user.TID,
		&user.UUID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) GetLeaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	rows, err := s.querier(ctx).Query(ctx, `
		SELECT
			u.id,
			u.username,
			u.first_name,
			u.last_name,
			COUNT(ua.id) FILTER (WHERE ua.status = 'remember') AS remember_count,
			COUNT(ua.id) FILTER (WHERE ua.status = 'forgot') AS forgot_count,
			COUNT(ua.id) FILTER (WHERE ua.status in ('remember', 'forgot')) AS answered_count,
			COUNT(DISTINCT ts.id) AS started_sessions
		FROM users u
		LEFT JOIN test_sessions ts ON ts.user_id = u.id
		LEFT JOIN user_answers ua ON ua.test_session_id = ts.id
		GROUP BY u.id, u.username, u.first_name, u.last_name
		ORDER BY answered_count DESC, u.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]LeaderboardEntry, 0)
	for rows.Next() {
		var entry LeaderboardEntry
		err = rows.Scan(
			&entry.ID,
			&entry.Username,
			&entry.FirstName,
			&entry.LastName,
			&entry.RememberCount,
			&entry.ForgotCount,
			&entry.AnsweredCount,
			&entry.StartedSessions,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
