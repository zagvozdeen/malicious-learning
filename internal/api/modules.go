package api

import (
	"fmt"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getModules(r *http.Request, _ *store.User) Response {
	slug := r.URL.Query().Get("course_slug")
	if slug == "" {
		return rErr(http.StatusBadRequest, fmt.Errorf("missing course_slug"))
	}

	modules, err := s.store.GetModulesByCourseSlug(r.Context(), slug)
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to get modules: %w", err))
	}
	return rData(http.StatusOK, modules)
}
