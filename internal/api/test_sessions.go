package api

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

type createTestSessionRequest struct {
	CourseSlug string `json:"course_slug"`
	ModuleIDs  []int  `json:"module_ids"`
	Shuffle    bool   `json:"shuffle"`
}

func (s *Service) createTestSession(r *http.Request, user *store.User) Response {
	var payload createTestSessionRequest
	if err := json.UnmarshalRead(r.Body, &payload); err != nil {
		return rErr(http.StatusBadRequest, fmt.Errorf("invalid json body: %w", err))
	}

	payload.CourseSlug = strings.TrimSpace(payload.CourseSlug)
	if payload.CourseSlug == "" {
		return rErr(http.StatusBadRequest, fmt.Errorf("missing course_slug"))
	}

	moduleIDs := slices.Clone(payload.ModuleIDs)
	slices.Sort(moduleIDs)

	cards, err := s.store.GetCards(r.Context(), payload.CourseSlug, moduleIDs)
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to load cards: %w", err))
	}

	if payload.Shuffle && len(cards) > 1 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		rng.Shuffle(len(cards), func(i, j int) {
			cards[i], cards[j] = cards[j], cards[i]
		})
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to create uuid v7: %w", err))
	}
	now := time.Now()
	session := &store.TestSession{
		UUID:       uid.String(),
		UserID:     user.ID,
		ModuleIDs:  moduleIDs,
		IsShuffled: payload.Shuffle,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	answers := make([]store.UserAnswer, 0, len(cards))
	for _, card := range cards {
		uid, err = uuid.NewV7()
		if err != nil {
			return rErr(http.StatusInternalServerError, fmt.Errorf("failed to create uuid v7: %w", err))
		}
		answers = append(answers, store.UserAnswer{
			UUID:      uid.String(),
			CardID:    card.ID,
			Status:    store.UserAnswerStatusNull,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	err = s.store.CreateTestSession(r.Context(), session, answers)
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to create test session: %w", err))
	}
	s.metrics.AppCreatedTestSessionsCountInc()

	return rData(http.StatusOK, session)
}

type getTestSessionResponse struct {
	TestSession *store.TestSession     `json:"test_session"`
	UserAnswers []store.FullUserAnswer `json:"user_answers"`
}

func (s *Service) getTestSession(r *http.Request, user *store.User) Response {
	groupUUID := r.PathValue("uuid")
	if groupUUID == "" {
		return rErr(http.StatusBadRequest, fmt.Errorf("missing uuid"))
	}
	if err := uuid.Validate(groupUUID); err != nil {
		return rErr(http.StatusBadRequest, fmt.Errorf("invalid uuid: %w", err))
	}

	ts, err := s.store.GetTestSessionByUUID(r.Context(), groupUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rErr(http.StatusNotFound, fmt.Errorf("test session not found: %w", err))
		}
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to get test session: %w", err))
	}
	if ts.UserID != user.ID {
		return rErr(http.StatusForbidden, fmt.Errorf("you can not get test session: %w", err))
	}

	answers, err := s.store.GetUserAnswersByTestSessionID(r.Context(), ts.ID)
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to load user answers: %w", err))
	}

	return rData(http.StatusOK, getTestSessionResponse{
		TestSession: ts,
		UserAnswers: answers,
	})
}

type getTestSessionsResponse struct {
	Data []store.TestSessionSummary `json:"data"`
}

func (s *Service) getTestSessions(r *http.Request, user *store.User) Response {
	sessions, err := s.store.GetTestSessions(r.Context(), user.ID)
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to load test sessions: %w", err))
	}

	return rData(http.StatusOK, getTestSessionsResponse{Data: sessions})
}
