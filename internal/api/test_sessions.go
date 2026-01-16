package api

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) createTestSession(r *http.Request, user *store.User) Response {
	query := r.URL.Query()

	shuffle, err := parseBool(query.Get("shuffle"))
	if err != nil {
		return rErr(http.StatusBadRequest, fmt.Errorf("invalid shuffle param: %w", err))
	}

	moduleIDs, err := parseModuleIDs(query.Get("modules"))
	if err != nil {
		return rErr(http.StatusBadRequest, fmt.Errorf("invalid modules param: %w", err))
	}

	cards, err := s.store.GetAllCards(r.Context())
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to load cards: %w", err))
	}

	filtered := cards
	if len(moduleIDs) > 0 {
		filtered = make([]store.Card, 0, len(cards))
		for _, card := range cards {
			if _, ok := moduleIDs[card.ModuleID]; ok {
				filtered = append(filtered, card)
			}
		}
	}

	if shuffle && len(filtered) > 1 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		rng.Shuffle(len(filtered), func(i, j int) {
			filtered[i], filtered[j] = filtered[j], filtered[i]
		})
	}

	sessionModuleIDs := make([]int, 0)
	if len(moduleIDs) > 0 {
		sessionModuleIDs = make([]int, 0, len(moduleIDs))
		for id := range moduleIDs {
			sessionModuleIDs = append(sessionModuleIDs, id)
		}
	} else {
		moduleSet := make(map[int]struct{})
		for _, card := range filtered {
			moduleSet[card.ModuleID] = struct{}{}
		}
		sessionModuleIDs = make([]int, 0, len(moduleSet))
		for id := range moduleSet {
			sessionModuleIDs = append(sessionModuleIDs, id)
		}
	}
	sort.Ints(sessionModuleIDs)

	uid, err := uuid.NewV7()
	if err != nil {
		return rErr(http.StatusInternalServerError, fmt.Errorf("failed to create uuid v7: %w", err))
	}
	now := time.Now()
	session := &store.TestSession{
		UUID:       uid.String(),
		UserID:     user.ID,
		ModuleIDs:  sessionModuleIDs,
		IsShuffled: shuffle,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	answers := make([]store.UserAnswer, 0, len(filtered))
	for _, card := range filtered {
		uaID, err := uuid.NewV7()
		if err != nil {
			return rErr(http.StatusInternalServerError, fmt.Errorf("failed to create uuid v7: %w", err))
		}
		answers = append(answers, store.UserAnswer{
			UUID:      uaID.String(),
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

func parseBool(value string) (bool, error) {
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

func parseModuleIDs(value string) (map[int]struct{}, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	ids := make(map[int]struct{})
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		ids[id] = struct{}{}
	}
	return ids, nil
}
