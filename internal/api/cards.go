package api

import (
	"fmt"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/api/core"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getCards(r *http.Request, _ *store.User) core.Response {
	cards, err := s.store.GetAllCards(r.Context())
	if err != nil {
		return core.Err(http.StatusInternalServerError, fmt.Errorf("failed to get all cars: %w", err))
	}
	return core.Data(http.StatusOK, cards)
}
