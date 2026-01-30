package enum

import (
	"database/sql/driver"
	"encoding/json/jsontext"
	"errors"
	"fmt"
)

type UserAnswerStatus struct {
	slug      string
	condition string
}

func NewUserAnswerStatus(s string) (UserAnswerStatus, error) {
	switch s {
	case UserAnswerStatusNull.slug:
		return UserAnswerStatusNull, nil
	case UserAnswerStatusRemember.slug:
		return UserAnswerStatusRemember, nil
	case UserAnswerStatusForgot.slug:
		return UserAnswerStatusForgot, nil
	default:
		return UserAnswerStatus{}, fmt.Errorf("unknown user answer status: %s", s)
	}
}

var (
	UserAnswerStatusNull     = UserAnswerStatus{"null", "Не ответил"}
	UserAnswerStatusRemember = UserAnswerStatus{"remember", "Вспомнил"}
	UserAnswerStatusForgot   = UserAnswerStatus{"forgot", "Забыл"}
)

func (u UserAnswerStatus) String() string {
	return u.slug
}

func (u UserAnswerStatus) Condition() string {
	return u.condition
}

func (u *UserAnswerStatus) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return errors.New("can not assert user answer status to string")
	}
	r, err := NewUserAnswerStatus(s)
	if err != nil {
		return err
	}
	*u = r
	return nil
}

func (u UserAnswerStatus) Value() (driver.Value, error) {
	return u.String(), nil
}

func (u UserAnswerStatus) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteToken(jsontext.String(u.slug))
}

func (u *UserAnswerStatus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	if tok.Kind() != '"' {
		return errors.New("user answer status must be a JSON string")
	}
	e, err := NewUserAnswerStatus(tok.String())
	if err != nil {
		return err
	}
	*u = e
	return nil
}
