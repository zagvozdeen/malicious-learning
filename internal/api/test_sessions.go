package api

import (
	"encoding/json/v2"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

func (s *Service) createTestSession(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	shuffle, err := parseBool(query.Get("shuffle"))
	if err != nil {
		http.Error(w, "invalid shuffle param", http.StatusBadRequest)
		return
	}

	moduleIDs, err := parseModuleIDs(query.Get("modules"))
	if err != nil {
		http.Error(w, "invalid modules param", http.StatusBadRequest)
		return
	}

	cards, err := s.store.GetAllCards(r.Context())
	if err != nil {
		s.log.Error("Failed to load cards", slog.Any("err", err))
		http.Error(w, "failed to load cards", http.StatusInternalServerError)
		return
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

	ctx := r.Context().Value("user")
	user, ok := ctx.(*store.User)
	if !ok || user == nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	groupUUID := uuid.NewString()
	now := time.Now()
	answers := make([]store.UserAnswer, 0, len(filtered))
	for _, card := range filtered {
		answers = append(answers, store.UserAnswer{
			UUID:      uuid.NewString(),
			GroupUUID: groupUUID,
			CardID:    card.ID,
			UserID:    user.ID,
			Status:    store.UserAnswerStatusNull,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	err = s.store.CreateUserAnswers(r.Context(), answers)
	if err != nil {
		s.log.Error("Failed to create user answers", slog.Any("err", err))
		http.Error(w, "failed to create test session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err = json.MarshalWrite(w, map[string]any{
		"group_uuid": groupUUID,
	})
	if err != nil {
		s.log.Warn("Failed to write response", slog.Any("err", err))
		return
	}
	s.log.Info("Created test session")
}

func (s *Service) getTestSession(w http.ResponseWriter, r *http.Request) {
	groupUUID := r.PathValue("uuid")
	if groupUUID == "" {
		http.Error(w, "missing uuid", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(groupUUID); err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	ctx := r.Context().Value("user")
	user, ok := ctx.(*store.User)
	if !ok || user == nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	answers, err := s.store.GetUserAnswersByGroupUUID(r.Context(), groupUUID)
	if err != nil {
		s.log.Error("Failed to load test session", slog.Any("err", err), slog.String("group_uuid", groupUUID))
		http.Error(w, "failed to load test session", http.StatusInternalServerError)
		return
	}
	if len(answers) == 0 {
		http.Error(w, "test session not found", http.StatusNotFound)
		return
	}

	type testSessionAnswer struct {
		UUID       string                 `json:"uuid"`
		GroupUUID  string                 `json:"group_uuid"`
		CardID     int                    `json:"card_id"`
		Status     store.UserAnswerStatus `json:"status"`
		Answer     string                 `json:"answer"`
		Question   string                 `json:"question"`
		ModuleID   int                    `json:"module_id"`
		ModuleName string                 `json:"module_name"`
	}

	items := make([]testSessionAnswer, 0, len(answers))
	for _, answer := range answers {
		if answer.UserID != user.ID {
			s.log.Warn("Forbidden test session access", slog.String("group_uuid", groupUUID), slog.Int("user_id", user.ID))
			http.Error(w, "test session not found", http.StatusNotFound)
			return
		}
		items = append(items, testSessionAnswer{
			UUID:       answer.UUID,
			GroupUUID:  answer.GroupUUID,
			CardID:     answer.CardID,
			Status:     answer.Status,
			Answer:     answer.Answer,
			Question:   answer.Question,
			ModuleID:   answer.ModuleID,
			ModuleName: answer.ModuleName,
		})
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err = json.MarshalWrite(w, map[string]any{
		"data": items,
	})
	if err != nil {
		s.log.Warn("Failed to write response", slog.Any("err", err))
	}
}

type getTestSessionsResponse struct {
	Data []store.TestSessionSummary `json:"data"`
}

func (s *Service) getTestSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context().Value("user")
	user, ok := ctx.(*store.User)
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
