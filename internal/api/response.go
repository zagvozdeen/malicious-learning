package api

import (
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/analytics"
)

type Response interface {
	Response(w http.ResponseWriter, req *http.Request, log *slog.Logger, metrics analytics.Metrics)
}

type ResponseError struct {
	code int
	err  error
}

type ResponseData struct {
	code int
	data any
}

type FlushData struct {
	data <-chan []byte
}

var _ Response = (*ResponseError)(nil)
var _ Response = (*ResponseData)(nil)
var _ Response = (*FlushData)(nil)

func rErr(code int, err error) *ResponseError {
	return &ResponseError{code: code, err: err}
}

func rData(code int, d any) *ResponseData {
	return &ResponseData{code: code, data: d}
}

func rFlash(data <-chan []byte) *FlushData {
	return &FlushData{data: data}
}

func (r *ResponseError) Response(w http.ResponseWriter, req *http.Request, log *slog.Logger, metrics analytics.Metrics) {
	log.Debug("Internal error", slog.Any("err", r.err), slog.Int("code", r.code))
	http.Error(w, r.err.Error(), r.code)
	metrics.AppResponsesTotalInc(req.Pattern, r.code)
}

func (r *ResponseData) Response(w http.ResponseWriter, req *http.Request, log *slog.Logger, metrics analytics.Metrics) {
	w.WriteHeader(r.code)
	if r.code == http.StatusNoContent && r.data == nil {
		metrics.AppResponsesTotalInc(req.Pattern, r.code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.MarshalWrite(w, r.data)
	if err != nil {
		log.Error("Failed to marshal response", slog.Any("err", err), slog.Int("code", r.code))
	}
	metrics.AppResponsesTotalInc(req.Pattern, r.code)
}

func (r *FlushData) Response(w http.ResponseWriter, req *http.Request, log *slog.Logger, metrics analytics.Metrics) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		rErr(http.StatusHTTPVersionNotSupported, errors.New("streaming not supported")).Response(w, req, log, metrics)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var b []byte
	for {
		select {
		case <-req.Context().Done():
			metrics.AppResponsesTotalInc(req.Pattern, http.StatusGone)
			return
		case b, ok = <-r.data:
			if !ok {
				metrics.AppResponsesTotalInc(req.Pattern, http.StatusOK)
				return
			}
			if _, err := w.Write(b); err != nil {
				metrics.AppResponsesTotalInc(req.Pattern, http.StatusInternalServerError)
				log.Error("Failed to write a piece of data", slog.Any("err", err))
				return
			}
			flusher.Flush()
		}
	}
}
