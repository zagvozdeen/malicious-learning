package api

import (
	"context"
	"encoding/json/v2"
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

	mux.HandleFunc("GET /", s.index)
	mux.HandleFunc("POST /api/test-session", s.createTestSession)
	//mux.HandleFunc("GET /v1/reports/top-routes-dim", s.getTopRoutesDim)
	//mux.HandleFunc("GET /v1/reports/error-rate", s.getErrorRate)
	//mux.HandleFunc("GET /v1/reports/latency", s.getLatency)
	//mux.HandleFunc("GET /v1/reports/timeseries", s.getTimeSeries)

	return mux
}

func (s *Service) index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
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
	now := time.Now().UTC()
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

	if err := s.store.CreateUserAnswers(r.Context(), answers); err != nil {
		s.log.Error("Failed to create user answers", slog.Any("err", err))
		http.Error(w, "failed to create test session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.MarshalWrite(w, map[string]any{
		"group_uuid": groupUUID,
		"count":      len(answers),
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
