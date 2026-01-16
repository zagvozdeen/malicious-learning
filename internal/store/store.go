package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zagvozdeen/malicious-learning/internal/config"
)

type Storage interface {
	CreateTelegramUpdate(ctx context.Context, update *TelegramUpdate) error

	GetModuleByName(ctx context.Context, name string) (*Module, error)
	CreateModule(ctx context.Context, module *Module) error

	GetAllCards(ctx context.Context) ([]Card, error)
	GetCardByUIDAndHash(ctx context.Context, uid int, hash string) (*Card, error)
	GetActiveCardByUID(ctx context.Context, uid int) (*Card, error)
	CreateCard(ctx context.Context, card *Card) error
	DeactivateCardByID(ctx context.Context, id int, updatedAt time.Time) error

	CreateTestSession(ctx context.Context, session *TestSession, answers []UserAnswer) error
	GetTestSessionByUUID(ctx context.Context, uuid string) (*TestSession, error)
	GetUserAnswersByTestSessionID(ctx context.Context, id int) ([]FullUserAnswer, error)
	GetTestSessions(ctx context.Context, userID int) ([]TestSessionSummary, error)
	GetUserAnswerByUUID(ctx context.Context, uuid string) (*UserAnswer, error)
	UpdateUserAnswer(ctx context.Context, ua *UserAnswer) error

	GetUserByID(ctx context.Context, id int) (*User, error)
	GetUserByTID(ctx context.Context, tid int64) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
}

type Store struct {
	cfg  *config.Config
	log  *slog.Logger
	pool *pgxpool.Pool
}

var _ Storage = (*Store)(nil)

func New(cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool) *Store {
	return &Store{cfg: cfg, log: log, pool: pool}
}
