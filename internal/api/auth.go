package api

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	req := &authRequest{}
	err := json.UnmarshalRead(r.Body, req)
	if err != nil {
		return
		//return c.Err(http.StatusBadRequest, fmt.Errorf("failed to parse auth request: %w", err))
	}
	user, err := s.store.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		return
		//return c.SecureErr(http.StatusNotFound, "invalid email or password", fmt.Errorf("failed to get user by email %q: %w", req.Username, err))
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password.V), []byte(req.Password))
	if err != nil {
		return
		//return c.SecureErr(http.StatusBadRequest, "invalid email or password", fmt.Errorf("invalid password for user %q: %w", req.Username, err))
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ID:        strconv.Itoa(user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 365)),
	})
	token, err := t.SignedString([]byte(s.cfg.AppSecret))
	if err != nil {
		return
		//return c.Err(http.StatusInternalServerError, fmt.Errorf("failed to sign token: %w", err))
	}
	err = json.MarshalWrite(w, authResponse{Token: token})
	if err != nil {
		s.log.Warn("Failed to write response", slog.Any("err", err))
	}
}
