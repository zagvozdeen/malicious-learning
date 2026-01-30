package api

import (
	"errors"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getChanges(_ *http.Request, user *store.User) Response {
	value, ok := s.processingTS.Load(user.ID)
	if !ok {
		return rData(http.StatusNoContent, nil)
	}
	ch, ok := value.(chan []byte)
	if !ok {
		return rErr(http.StatusInternalServerError, errors.New("invalid changes channel"))
	}

	return rFlash(ch)
}
