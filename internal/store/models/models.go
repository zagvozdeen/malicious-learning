package models

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

type UserAnswer struct {
	ID        int
	UUID      string
	GroupUUID string
	CardID    int
	UserID    int
	Status    UserAnswerStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FullUserAnswer struct {
	UserAnswer

	Answer     string
	Question   string
	ModuleID   int
	ModuleName string
}
