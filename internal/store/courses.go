package store

import (
	"context"
	"time"
)

type Course struct {
	ID        int       `json:"id"`
	UUID      string    `json:"uuid"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Store) GetCourses(ctx context.Context) ([]Course, error) {
	return nil, nil // TODO
}

func (s *Store) CreateCourse(ctx context.Context, course *Course) error {
	return nil // TODO
}
