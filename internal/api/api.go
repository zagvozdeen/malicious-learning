package api

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zagvozdeen/malicious-learning"
	"github.com/zagvozdeen/malicious-learning/internal/config"
	"github.com/zagvozdeen/malicious-learning/internal/store"
	"github.com/zagvozdeen/malicious-learning/internal/store/models"
)

type Service struct {
	ctx   context.Context
	cfg   *config.Config
	log   *slog.Logger
	store store.Storage
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger, store store.Storage) *Service {
	return &Service{
		ctx:   ctx,
		cfg:   cfg,
		log:   log,
		store: store,
	}
}

func (s *Service) Run() {
	addr := net.JoinHostPort(s.cfg.APIHost, s.cfg.APIPort)
	server := &http.Server{
		Addr:     addr,
		Handler:  s.getRoutes(),
		ErrorLog: slog.NewLogLogger(s.log.Handler(), slog.LevelDebug),
	}
	stop := make(chan error, 1)
	wg := &sync.WaitGroup{}
	wg.Go(func() {
		stop <- server.ListenAndServe()
		close(stop)
	})
	wg.Go(func() {
		if err := s.createRootUser(); err != nil {
			s.log.Warn("Failed to create root user", slog.Any("err", err))
			return
		}
		s.log.Info("Root user created or already exists", slog.String("username", s.cfg.RootUserName))
	})
	wg.Go(func() {
		if err := s.startBot(); err != nil {
			s.log.Warn("Failed to start bot", slog.Any("err", err))
			return
		}
		s.log.Info("Bot stopped")
	})
	wg.Go(func() {
		if err := malicious_learning.ParseQuestions(s.ctx, s.store); err != nil {
			s.log.Warn("Failed to parse questions", slog.Any("err", err))
			return
		}
		s.log.Info("Questions parsed")
	})
	select {
	case <-time.After(time.Millisecond * 500):
		s.log.Info(fmt.Sprintf("Server started on %s", addr))
	case err := <-stop:
		s.log.Error("Failed to listen and serve server", slog.Any("err", err))
		return
	}
	<-s.ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	err := server.Shutdown(ctx)
	if err != nil {
		s.log.Warn("Failed to shutdown server", slog.Any("err", err))
	}
	cancel()
	wg.Wait()
	s.log.Info("Server has been stopped")
}

func (s *Service) getRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	if !s.cfg.IsProduction {
		mux.HandleFunc("GET /", s.index)
		mux.Handle("GET /node_modules/", http.StripPrefix("/node_modules/", http.FileServer(http.Dir("node_modules"))))
	}

	mux.HandleFunc("POST /api/auth", s.login)
	mux.HandleFunc("GET /api/test-sessions", s.auth(s.getTestSessions))
	mux.HandleFunc("POST /api/test-sessions", s.auth(s.createTestSession))
	mux.HandleFunc("GET /api/test-sessions/{uuid}", s.auth(s.getTestSession))
	mux.HandleFunc("PATCH /api/user-answers/{uuid}", s.auth(s.updateUserAnswer))

	return mux
}

func (s *Service) index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "dev.html")
}

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
		filtered = make([]models.Card, 0, len(cards))
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
	answers := make([]models.UserAnswer, 0, len(filtered))
	for _, card := range filtered {
		answers = append(answers, models.UserAnswer{
			UUID:      uuid.NewString(),
			GroupUUID: groupUUID,
			CardID:    card.ID,
			UserID:    user.ID,
			Status:    models.UserAnswerStatusNull,
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
		UUID       string                  `json:"uuid"`
		GroupUUID  string                  `json:"group_uuid"`
		CardID     int                     `json:"card_id"`
		Status     models.UserAnswerStatus `json:"status"`
		Answer     string                  `json:"answer"`
		Question   string                  `json:"question"`
		ModuleID   int                     `json:"module_id"`
		ModuleName string                  `json:"module_name"`
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

func (s *Service) updateUserAnswer(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Status string `json:"status"`
	}
	if err := json.UnmarshalRead(r.Body, &payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	uuidValue := r.PathValue("uuid")
	if uuidValue == "" {
		http.Error(w, "missing uuid", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(uuidValue); err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	statusValue := strings.TrimSpace(payload.Status)
	var status models.UserAnswerStatus
	switch statusValue {
	case string(models.UserAnswerStatusRemember):
		status = models.UserAnswerStatusRemember
	case string(models.UserAnswerStatusForgot):
		status = models.UserAnswerStatusForgot
	default:
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	ua, err := s.store.GetUserAnswerByUUID(r.Context(), uuidValue)
	if err != nil {
		http.Error(w, "user answer not found", http.StatusNotFound)
		return
	}
	ua.Status = status
	ua.UpdatedAt = time.Now()
	err = s.store.UpdateUserAnswer(r.Context(), ua)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "user answer not found", http.StatusNotFound)
			return
		}
		s.log.Error("Failed to update user answer", slog.Any("err", err))
		http.Error(w, "failed to update user answer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.MarshalWrite(w, map[string]any{
		"uuid":   uuidValue,
		"status": status,
	})
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
