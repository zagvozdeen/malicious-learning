package api

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
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

type updateUserAnswerResponse struct {
	UserAnswer  *store.UserAnswer  `json:"data"`
	TestSession *store.TestSession `json:"test_session"`
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

	ctx, err := s.store.Begin(r.Context())
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to begin transaction: %w", err))
	}
	defer s.store.Rollback(ctx)

	var ua *store.UserAnswer
	ua, err = s.store.GetUserAnswerByUUID(ctx, uuidValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rErr(http.StatusNotFound, fmt.Errorf("user answer not found: %w", err))
		}
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to get user answer: %w", err))
	}
	if ua.Status != store.UserAnswerStatusNull {
		return rErr(http.StatusForbidden, fmt.Errorf("user answer status must be null"))
	}
	var ts *store.TestSession
	var still int
	ts, still, err = s.store.GetTestSessionByID(ctx, ua.TestSessionID)
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to get test session: %w", err))
	}
	if ts.UserID != user.ID {
		return rErr(http.StatusForbidden, fmt.Errorf("you can not edit this user answer"))
	}
	if !ts.IsActive {
		return rErr(http.StatusForbidden, fmt.Errorf("test session is not active"))
	}
	ua.Status = status
	ua.UpdatedAt = time.Now()
	err = s.store.UpdateUserAnswer(ctx, ua)
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to update user answer: %w", err))
	}
	if still-1 == 0 {
		ts.IsActive = false
		ts.UpdatedAt = time.Now()
		err = s.store.UpdateTestSession(ctx, ts)
		if err != nil {
			return rErr(http.StatusInternalServerError, fmt.Errorf("failed to update test session: %w", err))
		}
	}
	s.store.Commit(ctx)
	s.metrics.AppUpdatedUserAnswersCountInc()

	if still-1 == 0 {
		go func() {
			if err = s.getUserRecommendationsByTestSessionID(user, ts.ID); err != nil {
				s.log.Error("Failed to create recommendations", slog.Any("err", err))
			}
		}()
	}

	return rData(http.StatusOK, updateUserAnswerResponse{
		UserAnswer:  ua,
		TestSession: ts,
	})
}
