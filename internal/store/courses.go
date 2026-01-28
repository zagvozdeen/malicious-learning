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
	rows, err := s.querier(ctx).Query(
		ctx,
		"SELECT id, uuid, slug, name, created_at, updated_at FROM courses ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	courses := make([]Course, 0)
	for rows.Next() {
		var course Course
		err = rows.Scan(
			&course.ID,
			&course.UUID,
			&course.Slug,
			&course.Name,
			&course.CreatedAt,
			&course.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return courses, nil
}

func (s *Store) GetCourseBySlug(ctx context.Context, slug string) (*Course, error) {
	course := &Course{}
	err := s.querier(ctx).QueryRow(
		ctx,
		"SELECT id, uuid, slug, name, created_at, updated_at FROM courses WHERE slug = $1",
		slug,
	).Scan(&course.ID, &course.UUID, &course.Slug, &course.Name, &course.CreatedAt, &course.UpdatedAt)
	return course, err
}

func (s *Store) CreateCourse(ctx context.Context, course *Course) error {
	return s.querier(ctx).QueryRow(
		ctx,
		"INSERT INTO courses (uuid, slug, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		course.UUID, course.Slug, course.Name, course.CreatedAt, course.UpdatedAt,
	).Scan(&course.ID)
}
