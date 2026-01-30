package api

import (
	"fmt"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/api/core"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

type getLeaderboardResponse struct {
	Data []store.LeaderboardEntry `json:"data"`
}

func (s *Service) getLeaderboard(r *http.Request, _ *store.User) core.Response {
	entries, err := s.store.GetLeaderboard(r.Context())
	if err != nil {
		return core.Err(http.StatusInternalServerError, fmt.Errorf("failed to load leaderboard: %w", err))
	}

	return core.Data(http.StatusOK, getLeaderboardResponse{Data: entries})
}
