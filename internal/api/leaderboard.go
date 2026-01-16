package api

import (
	"fmt"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/store"
)

type getLeaderboardResponse struct {
	Data []store.LeaderboardEntry `json:"data"`
}

func (s *Service) getLeaderboard(r *http.Request, _ *store.User) Response {
	entries, err := s.store.GetLeaderboard(r.Context())
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to load leaderboard: %w", err))
	}

	return rData(http.StatusOK, getLeaderboardResponse{Data: entries})
}
