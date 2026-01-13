package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/zagvozdeen/malicious-learning/internal/db/null"
	"github.com/zagvozdeen/malicious-learning/internal/store/models"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) createRootUser() error {
	uid, err := uuid.NewV7()
	if err != nil {
		return err
	}
	var password []byte
	password, err = bcrypt.GenerateFromPassword([]byte(s.cfg.RootUserPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := &models.User{
		UUID:      uid.String(),
		FirstName: s.cfg.RootUserName,
		Username:  null.WrapString(s.cfg.RootUserName),
		Password:  null.WrapString(string(password)),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = s.store.CreateUser(s.ctx, u)
	if err != nil {
		return err
	}
	return nil
}
