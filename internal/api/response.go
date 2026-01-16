package api

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/store"
)

type Response interface {
	Response(w http.ResponseWriter, log *slog.Logger, user *store.User)
}

type ResponseError struct {
	code int
	err  error
}

type ResponseData struct {
	code int
	data any
}

var _ Response = (*ResponseError)(nil)
var _ Response = (*ResponseData)(nil)

func rErr(code int, err error) *ResponseError {
	return &ResponseError{code: code, err: err}
}

func rData(code int, d any) *ResponseData {
	return &ResponseData{code: code, data: d}
}

func (r *ResponseError) Response(w http.ResponseWriter, log *slog.Logger, user *store.User) {
	log.Debug("Internal error",
		slog.Any("err", r.err),
		slog.Int("code", r.code),
		slog.Int("user_id", user.ID),
	)
	http.Error(w, r.err.Error(), r.code)
}

func (r *ResponseData) Response(w http.ResponseWriter, log *slog.Logger, user *store.User) {
	w.WriteHeader(r.code)
	w.Header().Set("Content-Type", "application/json")
	err := json.MarshalWrite(w, r.data)
	if err != nil {
		log.Error("Failed to marshal response",
			slog.Any("err", err),
			slog.Int("code", r.code),
			slog.Int("user_id", user.ID),
		)
	}
}
