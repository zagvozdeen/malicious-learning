package api

import (
	"fmt"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/api/core"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getModules(r *http.Request, _ *store.User) core.Response {
	slug := r.URL.Query().Get("course_slug")
	if slug == "" {
		return core.Err(http.StatusBadRequest, fmt.Errorf("missing course_slug"))
	}

	modules, err := s.store.GetModulesByCourseSlug(r.Context(), slug)
	if err != nil {
		return core.Err(http.StatusInternalServerError, fmt.Errorf("failed to get modules: %w", err))
	}
	return core.Data(http.StatusOK, modules)
}
