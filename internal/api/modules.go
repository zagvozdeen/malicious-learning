package api

import (
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getModules(r *http.Request, _ *store.User) Response {
	return nil // TODO
}
