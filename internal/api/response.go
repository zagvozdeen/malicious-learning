package api

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/analytics"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

type Response interface {
	Response(w http.ResponseWriter, req *http.Request, log *slog.Logger, user *store.User, metrics analytics.Metrics)
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

func (r *ResponseError) Response(w http.ResponseWriter, req *http.Request, log *slog.Logger, user *store.User, metrics analytics.Metrics) {
	userID := 0
	if user != nil {
		userID = user.ID
	}
	log.Debug("Internal error",
		slog.Any("err", r.err),
		slog.Int("code", r.code),
		slog.Int("user_id", userID),
	)
	http.Error(w, r.err.Error(), r.code)
	metrics.AppResponsesTotalInc(req.Pattern, r.code)
}

func (r *ResponseData) Response(w http.ResponseWriter, req *http.Request, log *slog.Logger, user *store.User, metrics analytics.Metrics) {
	w.WriteHeader(r.code)
	w.Header().Set("Content-Type", "application/json")
	err := json.MarshalWrite(w, r.data)
	if err != nil {
		userID := 0
		if user != nil {
			userID = user.ID
		}
		log.Error("Failed to marshal response",
			slog.Any("err", err),
			slog.Int("code", r.code),
			slog.Int("user_id", userID),
		)
	}
	metrics.AppResponsesTotalInc(req.Pattern, r.code)
}
