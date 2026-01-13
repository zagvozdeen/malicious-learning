package api

import (
	"context"
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/zagvozdeen/malicious-learning/internal/store/models"
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

func (s *Service) auth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		switch {
		case strings.HasPrefix(token, "tma "):
			token = strings.TrimPrefix(token, "tma ")
			values, err := url.ParseQuery(token)
			if err != nil {
				s.log.Warn("Failed to parse TMA token", slog.Any("err", err))
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			u, ok := bot.ValidateWebappRequest(values, s.cfg.TelegramBotToken)
			if !ok {
				s.log.Warn("Invalid TMA token")
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			var user *models.User
			user, err = s.store.GetUserByTID(r.Context(), u.ID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					s.log.Warn("TMA user not found", slog.Int64("tid", u.ID))
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}
				s.log.Error("Failed to load TMA user", slog.Any("err", err), slog.Int64("tid", u.ID))
				http.Error(w, "failed to authenticate", http.StatusInternalServerError)
				return
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
				s.log.Warn("Failed to parse auth token", slog.Any("err", err))
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			if !t.Valid {
				s.log.Warn("Invalid auth token")
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			id, err := strconv.Atoi(claims.ID)
			if err != nil {
				s.log.Warn("Invalid auth token ID", slog.String("id", claims.ID), slog.Any("err", err))
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			var user *models.User
			user, err = s.store.GetUserByID(r.Context(), id)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					s.log.Warn("Auth user not found", slog.Int("id", id))
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}
				s.log.Error("Failed to load auth user", slog.Any("err", err), slog.Int("id", id))
				http.Error(w, "failed to authenticate", http.StatusInternalServerError)
				return
			}
			fn(w, r.WithContext(context.WithValue(r.Context(), "user", user)))
			return
		default:
			s.log.Warn("Missing authorization header")
			http.Error(w, "authorization required", http.StatusUnauthorized)
			return
		}
	}
}
