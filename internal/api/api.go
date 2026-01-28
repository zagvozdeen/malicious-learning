package api

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/zagvozdeen/malicious-learning"
	"github.com/zagvozdeen/malicious-learning/internal/analytics"
	"github.com/zagvozdeen/malicious-learning/internal/config"
	"github.com/zagvozdeen/malicious-learning/internal/store"
)

type Service struct {
	ctx          context.Context
	cfg          *config.Config
	log          *slog.Logger
	store        store.Storage
	processingTS sync.Map
	events       chan Event
	flushers     sync.Map
	metrics      analytics.Metrics
	bot          *bot.Bot
	botStarted   chan struct{}
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger, store store.Storage, metrics analytics.Metrics) *Service {
	return &Service{
		ctx:          ctx,
		cfg:          cfg,
		log:          log,
		store:        store,
		processingTS: sync.Map{},
		events:       make(chan Event, 10),
		flushers:     sync.Map{},
		metrics:      metrics,
		botStarted:   make(chan struct{}),
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
			s.log.Warn("Failed to parse data", slog.Any("err", err))
			return
		}
		s.log.Info("Questions parsed")
	})
	wg.Go(func() {
		if err := s.loopEvent(); err != nil {
			s.log.Warn("Failed to handle events", slog.Any("err", err))
			return
		}
		s.log.Info("Event loop stopped")
	})
	wg.Go(func() {
		if err := s.startSendingMetrics(); err != nil {
			s.log.Warn("Failed to start sending metrics", slog.Any("err", err))
			return
		}
		s.log.Info("Metrics sender has been stopped")
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
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "dev.html")
		})
		mux.Handle("GET /node_modules/", http.StripPrefix("/node_modules/", http.FileServer(http.Dir("node_modules"))))
	}

	mux.HandleFunc("POST /api/auth", s.login)
	mux.HandleFunc("GET /api/test-sessions", s.auth(s.getTestSessions))
	mux.HandleFunc("GET /api/test-sessions/{uuid}", s.auth(s.getTestSession))
	mux.HandleFunc("POST /api/test-sessions", s.auth(s.createTestSession))
	mux.HandleFunc("PATCH /api/user-answers/{uuid}", s.auth(s.updateUserAnswer))
	mux.HandleFunc("GET /api/leaderboard", s.auth(s.getLeaderboard))
	mux.HandleFunc("GET /api/events", s.sseAuth(s.getEvents))
	mux.HandleFunc("GET /api/cards", s.auth(s.getCards))
	mux.HandleFunc("GET /api/courses", s.auth(s.getCourses))
	mux.HandleFunc("GET /api/modules", s.auth(s.getModules))

	return mux
}
