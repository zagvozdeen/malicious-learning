package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/api/core"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getChanges(r *http.Request, user *store.User) core.Response {
	value, ok := s.processingTS.Load(user.ID)
	if !ok {
		return core.Data(http.StatusNoContent, nil)
	}
	ch, ok := value.(chan []byte)
	if !ok {
		return core.Err(http.StatusInternalServerError, errors.New("invalid changes channel"))
	}

	ctx, cancel := context.WithCancel(r.Context())
	go func() {
		select {
		case <-s.ctx.Done():
		case <-r.Context().Done():
		}
		cancel()
	}()

	return core.Flush(ctx, ch)
}
