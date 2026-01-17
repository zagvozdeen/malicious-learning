package api

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/zagvozdeen/malicious-learning/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (s *Service) login(w http.ResponseWriter, r *http.Request) {
	req := &authRequest{}
	err := json.UnmarshalRead(r.Body, req)
	if err != nil {
		s.log.Warn("Failed to decode auth request", slog.Any("err", err))
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	user, err := s.store.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Warn("Invalid credentials", slog.String("username", req.Username))
			http.Error(w, "invalid username or password", http.StatusUnauthorized)
			return
		}
		s.log.Error("Failed to load user", slog.Any("err", err), slog.String("username", req.Username))
		http.Error(w, "failed to authenticate", http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password.V), []byte(req.Password))
	if err != nil {
		s.log.Warn("Invalid credentials", slog.String("username", req.Username))
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ID:        strconv.Itoa(user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 365)),
	})
	token, err := t.SignedString([]byte(s.cfg.AppSecret))
	if err != nil {
		s.log.Error("Failed to sign auth token", slog.Any("err", err))
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.MarshalWrite(w, authResponse{Token: token})
	if err != nil {
		s.log.Warn("Failed to write response", slog.Any("err", err))
	}
}

type HandlerFunc func(*http.Request, *store.User) Response
type SSEHandlerFunc func(http.ResponseWriter, *http.Request, http.Flusher, *store.User) error

func (s *Service) auth(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, res := s.checkAuth(r, r.Header.Get("Authorization"))
		if res != nil {
			res.Response(w, s.log, user)
			return
		}
		fn(r, user).Response(w, s.log, user)
	}
}

func (s *Service) sseAuth(fn SSEHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, res := s.checkAuth(r, r.URL.Query().Get("token"))
		if res != nil {
			res.Response(w, s.log, user)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			rErr(http.StatusHTTPVersionNotSupported, errors.New("streaming not supported")).Response(w, s.log, user)
			return
		}

		if err := fn(w, r, flusher, user); err != nil {
			s.log.Error("SSE auth failed", slog.Any("err", err))
		}
	}
}

func (s *Service) checkAuth(r *http.Request, token string) (*store.User, Response) {
	switch {
	case strings.HasPrefix(token, "tma "):
		return s.authTMA(r, token)
	case strings.HasPrefix(token, "Bearer "):
		return s.authBearer(r, token)
	default:
		return nil, rErr(http.StatusUnauthorized, fmt.Errorf("missing authorization header"))
	}
}

func (s *Service) authTMA(r *http.Request, token string) (*store.User, Response) {
	token = strings.TrimPrefix(token, "tma ")
	values, err := url.ParseQuery(token)
	if err != nil {
		return nil, rErr(http.StatusUnauthorized, fmt.Errorf("failed to parse tma token: %w", err))
	}
	u, ok := bot.ValidateWebappRequest(values, s.cfg.TelegramBotToken)
	if !ok {
		return nil, rErr(http.StatusUnauthorized, fmt.Errorf("invalid tma token"))
	}
	var user *store.User
	user, err = s.store.GetUserByTID(r.Context(), u.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, rErr(http.StatusUnauthorized, fmt.Errorf("tma user not found: %w", err))
		}
		return nil, rErr(http.StatusInternalServerError, fmt.Errorf("failed to load user: %w", err))
	}
	return user, nil
}

func (s *Service) authBearer(r *http.Request, token string) (*store.User, Response) {
	token = strings.TrimPrefix(token, "Bearer ")
	var claims jwt.RegisteredClaims
	t, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (any, error) {
		return []byte(s.cfg.AppSecret), nil
	})
	if err != nil {
		return nil, rErr(http.StatusUnauthorized, fmt.Errorf("failed to parse token: %w", err))
	}
	if !t.Valid {
		return nil, rErr(http.StatusUnauthorized, fmt.Errorf("invalid token"))
	}
	id, err := strconv.Atoi(claims.ID)
	if err != nil {
		return nil, rErr(http.StatusUnauthorized, fmt.Errorf("invalid token: %w, id=%s", err, claims.ID))
	}
	var user *store.User
	user, err = s.store.GetUserByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, rErr(http.StatusUnauthorized, fmt.Errorf("user not found: %w", err))
		}
		return nil, rErr(http.StatusInternalServerError, fmt.Errorf("failed to load user: %w", err))
	}
	return user, nil
}
