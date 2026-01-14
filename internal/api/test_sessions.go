package api

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/store/models"
)

type getTestSessionsResponse struct {
	Data []models.TestSessionSummary `json:"data"`
}

func (s *Service) getTestSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context().Value("user")
	user, ok := ctx.(*models.User)
	if !ok || user == nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	sessions, err := s.store.GetTestSessions(r.Context(), user.ID)
	if err != nil {
		s.log.Error("Failed to load test sessions", slog.Any("err", err), slog.Int("user_id", user.ID))
		http.Error(w, "failed to load test sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.MarshalWrite(w, getTestSessionsResponse{Data: sessions})
	if err != nil {
		s.log.Warn("Failed to write response", slog.Any("err", err))
	}
}
