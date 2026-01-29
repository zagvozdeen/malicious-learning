package api

import (
	"fmt"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) getCards(r *http.Request, _ *store.User) Response {
	cards, err := s.store.GetCards(r.Context(), "", []int{1, 2})
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to get all cars: %w", err))
	}
	return rData(http.StatusOK, cards)
}
