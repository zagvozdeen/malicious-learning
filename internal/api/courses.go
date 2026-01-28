package api

import (
	"fmt"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getCourses(r *http.Request, _ *store.User) Response {
	courses, err := s.store.GetCourses(r.Context())
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to get courses: %w", err))
	}
	return rData(http.StatusOK, courses)
}
