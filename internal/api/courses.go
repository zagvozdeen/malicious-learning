package api

import (
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getCourses(r *http.Request, _ *store.User) Response {
	return nil // TODO
}
