package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zagvozdeen/malicious-learning/internal/config"
	"github.com/zagvozdeen/malicious-learning/internal/store/models"
)

type Storage interface {
	GetAllCards(ctx context.Context) ([]models.Card, error)
	CreateTelegramUpdate(ctx context.Context, update *models.TelegramUpdate) error
	GetModuleByName(ctx context.Context, name string) (*models.Module, error)
	CreateModule(ctx context.Context, module *models.Module) error
	GetCardByUIDAndHash(ctx context.Context, uid int, hash string) (*models.Card, error)
	GetActiveCardByUID(ctx context.Context, uid int) (*models.Card, error)
	CreateCard(ctx context.Context, card *models.Card) error
	DeactivateCardByID(ctx context.Context, id int, updatedAt time.Time) error
	CreateUserAnswers(ctx context.Context, ua []models.UserAnswer) error
	GetUserAnswersByGroupUUID(ctx context.Context, uuid string) ([]models.FullUserAnswer, error)
	GetTestSessions(ctx context.Context, userID int) ([]models.TestSessionSummary, error)
	GetDistinctUserAnswers(ctx context.Context, userID int) ([]string, error)
	GetUserAnswerByUUID(ctx context.Context, uuid string) (*models.UserAnswer, error)
	UpdateUserAnswer(ctx context.Context, ua *models.UserAnswer) error
	GetUserByTID(ctx context.Context, tid int64) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
}

type User = models.User

type Store struct {
	cfg  *config.Config
	log  *slog.Logger
	pool *pgxpool.Pool
}

var _ Storage = (*Store)(nil)

func New(cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool) *Store {
	return &Store{cfg: cfg, log: log, pool: pool}
}

func (s Store) GetAllCards(ctx context.Context) ([]models.Card, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, uid, uuid, question, answer, module_id, is_active, hash, created_at, updated_at
		FROM cards
		WHERE is_active = true
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := make([]models.Card, 0)
	for rows.Next() {
		var card models.Card
		err = rows.Scan(
			&card.ID,
			&card.UID,
			&card.UUID,
			&card.Question,
			&card.Answer,
			&card.ModuleID,
			&card.IsActive,
			&card.Hash,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return cards, nil
}

func (s Store) CreateTelegramUpdate(ctx context.Context, update *models.TelegramUpdate) (err error) {
	_, err = s.pool.Exec(
		ctx,
		"INSERT INTO telegram_updates (id, update, date) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING",
		update.ID, update.Update, update.Date,
	)
	return err
}

func (s Store) GetModuleByName(ctx context.Context, name string) (*models.Module, error) {
	var module models.Module
	err := s.pool.QueryRow(ctx, `
		SELECT id, uuid, name, created_at, updated_at
		FROM modules
		WHERE name = $1
	`, name).Scan(
		&module.ID,
		&module.UUID,
		&module.Name,
		&module.CreatedAt,
		&module.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &module, nil
}

func (s Store) CreateModule(ctx context.Context, module *models.Module) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO modules (uuid, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, module.UUID, module.Name, module.CreatedAt, module.UpdatedAt).Scan(&module.ID)
}

func (s Store) GetCardByUIDAndHash(ctx context.Context, uid int, hash string) (*models.Card, error) {
	var card models.Card
	err := s.pool.QueryRow(ctx, `
		SELECT id, uid, uuid, question, answer, module_id, is_active, hash, created_at, updated_at
		FROM cards
		WHERE uid = $1 AND hash = $2
		LIMIT 1
	`, uid, hash).Scan(
		&card.ID,
		&card.UID,
		&card.UUID,
		&card.Question,
		&card.Answer,
		&card.ModuleID,
		&card.IsActive,
		&card.Hash,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (s Store) GetActiveCardByUID(ctx context.Context, uid int) (*models.Card, error) {
	var card models.Card
	err := s.pool.QueryRow(ctx, `
		SELECT id, uid, uuid, question, answer, module_id, is_active, hash, created_at, updated_at
		FROM cards
		WHERE uid = $1 AND is_active = true
		ORDER BY id DESC
		LIMIT 1
	`, uid).Scan(
		&card.ID,
		&card.UID,
		&card.UUID,
		&card.Question,
		&card.Answer,
		&card.ModuleID,
		&card.IsActive,
		&card.Hash,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (s Store) CreateCard(ctx context.Context, card *models.Card) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO cards (uid, uuid, question, answer, module_id, is_active, hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`,
		card.UID,
		card.UUID,
		card.Question,
		card.Answer,
		card.ModuleID,
		card.IsActive,
		card.Hash,
		card.CreatedAt,
		card.UpdatedAt,
	).Scan(&card.ID)
}

func (s Store) DeactivateCardByID(ctx context.Context, id int, updatedAt time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE cards
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`, updatedAt, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s Store) CreateUserAnswers(ctx context.Context, ua []models.UserAnswer) error {
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

func (s Store) GetUserAnswersByGroupUUID(ctx context.Context, uuid string) ([]models.FullUserAnswer, error) {
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

	answers := make([]models.FullUserAnswer, 0)
	for rows.Next() {
		var answer models.FullUserAnswer
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

func (s Store) GetTestSessions(ctx context.Context, userID int) ([]models.TestSessionSummary, error) {
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

	sessions := make([]models.TestSessionSummary, 0)
	for rows.Next() {
		var session models.TestSessionSummary
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

func (s Store) GetUserAnswerByUUID(ctx context.Context, uuid string) (*models.UserAnswer, error) {
	var answer models.UserAnswer
	err := s.pool.QueryRow(ctx, `
		SELECT id, uuid, group_uuid, card_id, user_id, status, created_at, updated_at
		FROM user_answers
		WHERE uuid = $1
	`, uuid).Scan(
		&answer.ID,
		&answer.UUID,
		&answer.GroupUUID,
		&answer.CardID,
		&answer.UserID,
		&answer.Status,
		&answer.CreatedAt,
		&answer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

func (s Store) UpdateUserAnswer(ctx context.Context, ua *models.UserAnswer) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE user_answers
		SET status = $1, updated_at = $2
		WHERE uuid = $3
	`, ua.Status, ua.UpdatedAt, ua.UUID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s Store) GetUserByTID(ctx context.Context, tid int64) (*models.User, error) {
	var user models.User
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
	return &user, nil
}

func (s Store) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
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
	return &user, nil
}

func (s Store) CreateUser(ctx context.Context, user *models.User) error {
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

func (s Store) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, tid, uuid, first_name, last_name, username, email, password, created_at, updated_at
		FROM users
		WHERE username = $1
	`, username).Scan(
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
	return &user, nil
}
