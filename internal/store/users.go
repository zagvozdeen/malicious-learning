package store

import "context"

func (s *Store) GetUserByTID(ctx context.Context, tid int64) (*User, error) {
	user := &User{}
	err := s.pool.QueryRow(ctx, `
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
	err := s.pool.QueryRow(ctx, `
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
	return s.pool.QueryRow(ctx, `
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
	err := s.pool.QueryRow(
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
