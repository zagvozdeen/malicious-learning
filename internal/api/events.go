package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/zagvozdeen/malicious-learning/internal/store"
)

type Event struct {
	ID    int
	Event string
	Data  string
}

type Flusher func(Event) error

func (s *Service) getEvents(w http.ResponseWriter, r *http.Request, flusher http.Flusher, user *store.User) error {
	var f Flusher = func(e Event) error {
		_, err := io.WriteString(w, fmt.Sprintf("event: %s\ndata: %s\n\n", e.Event, e.Data))
		if err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}
	s.flushers.Store(user.ID, f)
	defer s.flushers.Delete(user.ID)

	_ = f(Event{
		Event: "pong",
		Data:  "pong",
	})

	select {
	case <-r.Context().Done():
	case <-s.ctx.Done():
	}
	return nil
}

func (s *Service) loopEvent() error {
	for {
		select {
		case event, ok := <-s.events:
			if !ok {
				return nil
			}
			v, ok := s.flushers.Load(event.ID)
			if !ok {
				continue
			}
			flush, ok := v.(Flusher)
			if !ok {
				continue
			}
			err := flush(event)
			if err != nil {
				continue
			}
		case <-s.ctx.Done():
			return nil
		}
	}
}
