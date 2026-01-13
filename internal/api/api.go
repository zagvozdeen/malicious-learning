package api

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/zagvozdeen/malicious-learning/internal/config"
	"github.com/zagvozdeen/malicious-learning/internal/store"
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
	//mux.HandleFunc("GET /v1/reports/top-routes", s.getTopRoutes)
	//mux.HandleFunc("GET /v1/reports/top-routes-dim", s.getTopRoutesDim)
	//mux.HandleFunc("GET /v1/reports/error-rate", s.getErrorRate)
	//mux.HandleFunc("GET /v1/reports/latency", s.getLatency)
	//mux.HandleFunc("GET /v1/reports/timeseries", s.getTimeSeries)

	return mux
}

func (s *Service) index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}
