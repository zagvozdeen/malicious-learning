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
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/golang-jwt/jwt/v5"
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
		s.log.Info("Root user created", slog.String("username", s.cfg.RootUserName))
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

	mux.HandleFunc("GET /", s.index)
	mux.HandleFunc("POST /api/auth", s.login)
	mux.HandleFunc("POST /api/test-session", s.auth(s.createTestSession))
	mux.HandleFunc("PATCH /api/user-answers/{uuid}", s.auth(s.updateUserAnswer))

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

	w.Header().Set("Content-Type", "application/json")
	err = json.MarshalWrite(w, map[string]any{
		"group_uuid": groupUUID,
		"count":      len(answers),
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

func (s *Service) auth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		switch {
		case strings.HasPrefix(token, "tma "):
			token = strings.TrimPrefix(token, "tma ")
			values, err := url.ParseQuery(token)
			if err != nil {
				return
				//return c.SecureErr(http.StatusUnauthorized, "invalid token", fmt.Errorf("parse token %q: %w", token, err))
			}
			u, ok := bot.ValidateWebappRequest(values, s.cfg.TelegramBotToken)
			if !ok {
				return
				//return c.Err(http.StatusUnauthorized, errors.New("failed to validate token"))
			}
			var user *models.User
			user, err = s.store.GetUserByTID(r.Context(), u.ID)
			if err != nil {
				return
				//return c.SecureErr(http.StatusUnauthorized, "failed to get user", fmt.Errorf("failed to get user: %w", err))
			}
			fn(w, r.WithContext(context.WithValue(r.Context(), "user", user)))
			return
		case strings.HasPrefix(token, "Bearer "):
			token = strings.TrimPrefix(token, "Bearer ")
			var claims jwt.RegisteredClaims
			t, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (any, error) {
				return []byte(s.cfg.AppSecret), nil
			})
			if err != nil {
				return
				//return c.SecureErr(http.StatusUnauthorized, "invalid token", fmt.Errorf("parse token %q: %w", token, err))
			}
			if !t.Valid {
				return
				//return c.Err(http.StatusUnauthorized, errors.New("invalid token"))
			}
			id, err := strconv.Atoi(claims.ID)
			if err != nil {
				return
				//return c.Err(http.StatusUnauthorized, fmt.Errorf("invalid token ID %q: %w", claims.ID, err))
			}
			var user *models.User
			user, err = s.store.GetUserByID(r.Context(), id)
			if err != nil {
				return
				//return c.SecureErr(http.StatusUnauthorized, "failed to get user", fmt.Errorf("get user by ID %d: %w", id, err))
			}
			fn(w, r.WithContext(context.WithValue(r.Context(), "user", user)))
			return
			//return next(c)
		default:
			return
			//return c.SecureErr(http.StatusUnauthorized, "no token provided", fmt.Errorf("no token provided, token: %s", token))
		}
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
