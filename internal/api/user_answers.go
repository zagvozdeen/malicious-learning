package api

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

type updateUserAnswerRequest struct {
	Status string `json:"status"`
}

func (s *Service) updateUserAnswer(r *http.Request, user *store.User) Response {
	var payload updateUserAnswerRequest
	if err := json.UnmarshalRead(r.Body, &payload); err != nil {
		return rErr(http.StatusBadRequest, fmt.Errorf("invalid json body: %w", err))
	}

	uuidValue := r.PathValue("uuid")
	if uuidValue == "" {
		return rErr(http.StatusBadRequest, fmt.Errorf("missing uuid"))
	}
	if err := uuid.Validate(uuidValue); err != nil {
		return rErr(http.StatusBadRequest, fmt.Errorf("invalid uuid: %w", err))
	}

	status, err := store.ParseUserAnswerStatus(strings.TrimSpace(payload.Status))
	if err != nil {
		return rErr(http.StatusBadRequest, fmt.Errorf("invalid status: %w", err))
	}

	ua, err := s.store.GetUserAnswerByUUID(r.Context(), uuidValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rErr(http.StatusNotFound, fmt.Errorf("user answer not found: %w", err))
		}
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to get user answer: %w", err))
	}
	if ua.UserID != user.ID {
		return rErr(http.StatusForbidden, fmt.Errorf("you can not edit this user answer"))
	}
	ua.Status = status
	ua.UpdatedAt = time.Now()
	err = s.store.UpdateUserAnswer(r.Context(), ua)
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to update user answer: %w", err))
	}

	return rData(http.StatusOK, ua)
}
