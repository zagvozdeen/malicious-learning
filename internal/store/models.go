package store

import (
	"encoding/json"
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
	ID            int
	UUID          string
	CardID        int
	TestSessionID int
	Status        UserAnswerStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type FullUserAnswer struct {
	UserAnswer

	Answer     string
	Question   string
	ModuleID   int
	ModuleName string
}

type TestSessionSummary struct {
	GroupUUID     string    `json:"group_uuid"`
	CountNull     int       `json:"count_null"`
	CountRemember int       `json:"count_remember"`
	CountForget   int       `json:"count_forget"`
	CreatedAt     time.Time `json:"created_at"`
}
