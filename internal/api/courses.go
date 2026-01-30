package api

import (
	"fmt"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/api/core"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getCourses(r *http.Request, _ *store.User) core.Response {
	courses, err := s.store.GetCourses(r.Context())
	if err != nil {
		return core.Err(http.StatusInternalServerError, fmt.Errorf("failed to get courses: %w", err))
	}
	return core.Data(http.StatusOK, courses)
}
