package store

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zagvozdeen/malicious-learning/internal/db/null"
)

type TelegramUpdate struct {
	ID     int64
	Update json.RawMessage
	Date   time.Time
}

type User struct {
	ID        int
	TID       null.Int
	UUID      string
	FirstName string
	LastName  null.String
	Username  null.String
	Email     null.String
	Password  null.String
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LeaderboardEntry struct {
	ID              int         `json:"id"`
	Username        null.String `json:"username"`
	FirstName       string      `json:"first_name"`
	LastName        null.String `json:"last_name"`
	RememberCount   int         `json:"remember_count"`
	ForgotCount     int         `json:"forgot_count"`
	AnsweredCount   int         `json:"answered_count"`
	StartedSessions int         `json:"started_sessions"`
}

type Module struct {
	ID        int
	UUID      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Card struct {
	ID        int
	UID       int
	UUID      string
	Question  string
	Answer    string
	ModuleID  int
	IsActive  bool
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserAnswerStatus string

const (
	UserAnswerStatusNull     UserAnswerStatus = "null"
	UserAnswerStatusRemember UserAnswerStatus = "remember"
	UserAnswerStatusForgot   UserAnswerStatus = "forgot"
)

func ParseUserAnswerStatus(s string) (UserAnswerStatus, error) {
	switch s {
	case string(UserAnswerStatusNull):
		return UserAnswerStatusNull, nil
	case string(UserAnswerStatusRemember):
		return UserAnswerStatusRemember, nil
	case string(UserAnswerStatusForgot):
		return UserAnswerStatusForgot, nil
	default:
		return "", fmt.Errorf("invalid user answer status: %s", s)
	}
}

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

type UserAnswer struct {
	ID            int              `json:"id"`
	UUID          string           `json:"uuid"`
	CardID        int              `json:"card_id"`
	TestSessionID int              `json:"test_session_id"`
	Status        UserAnswerStatus `json:"status"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

type FullUserAnswer struct {
	UserAnswer

	Answer     string `json:"answer"`
	Question   string `json:"question"`
	ModuleID   int    `json:"module_id"`
	ModuleName string `json:"module_name"`
}

type TestSessionSummary struct {
	GroupUUID     string    `json:"group_uuid"`
	CountNull     int       `json:"count_null"`
	CountRemember int       `json:"count_remember"`
	CountForget   int       `json:"count_forget"`
	CreatedAt     time.Time `json:"created_at"`
}
