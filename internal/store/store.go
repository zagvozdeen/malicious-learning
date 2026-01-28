package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zagvozdeen/malicious-learning/internal/config"
)

type Storage interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context)
	Rollback(ctx context.Context)

	CreateTelegramUpdate(ctx context.Context, update *TelegramUpdate) error

	GetModulesByCourseID(ctx context.Context, id int) ([]Module, error)
	GetModuleByName(ctx context.Context, name string) (*Module, error)
	CreateModule(ctx context.Context, module *Module) error

	GetCourses(ctx context.Context) ([]Course, error)
	CreateCourse(ctx context.Context, course *Course) error

	GetCards(ctx context.Context, moduleIDs []int) ([]Card, error)
	GetCardByUIDAndHash(ctx context.Context, uid int, hash string) (*Card, error)
	GetActiveCardByUID(ctx context.Context, uid int) (*Card, error)
	CreateCard(ctx context.Context, card *Card) error
	DeactivateCardByID(ctx context.Context, id int, updatedAt time.Time) error

	CreateTestSession(ctx context.Context, session *TestSession, answers []UserAnswer) error
	UpdateTestSession(ctx context.Context, session *TestSession) error
	GetTestSessionByID(ctx context.Context, id int) (*TestSession, int, error)
	GetTestSessionByUUID(ctx context.Context, uuid string) (*TestSession, error)
	GetUserAnswersByTestSessionID(ctx context.Context, id int) ([]FullUserAnswer, error)
	GetTestSessions(ctx context.Context, userID int) ([]TestSessionSummary, error)
	GetUserAnswerByUUID(ctx context.Context, uuid string) (*UserAnswer, error)
	UpdateUserAnswer(ctx context.Context, ua *UserAnswer) error

	GetLeaderboard(ctx context.Context) ([]LeaderboardEntry, error)
	GetUserByID(ctx context.Context, id int) (*User, error)
	GetUserByTID(ctx context.Context, tid int64) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	CreateUser(ctx context.Context, user *User) error

	CreateChatCompletions(ctx context.Context, cc *ChatCompletions) error
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
