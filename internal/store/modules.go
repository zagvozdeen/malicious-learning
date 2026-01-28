package store

import (
	"context"
	"time"
)

type Module struct {
	ID        int
	UUID      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Store) GetModulesByCourseID(ctx context.Context, id int) ([]Module, error) {
	// for instance: select id, uuid, name, created_at, updated_at from modules where id in (select distinct module_id from cards where course_id = ?)
	return nil, nil
}

func (s *Store) GetModuleByName(ctx context.Context, name string) (*Module, error) {
	module := &Module{}
	err := s.querier(ctx).QueryRow(
		ctx,
		"SELECT id, uuid, name, created_at, updated_at FROM modules WHERE name = $1",
		name,
	).Scan(
		&module.ID,
		&module.UUID,
		&module.Name,
		&module.CreatedAt,
		&module.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return module, nil
}

func (s *Store) CreateModule(ctx context.Context, module *Module) error {
	return s.querier(ctx).QueryRow(
		ctx,
		"INSERT INTO modules (uuid, name, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id",
		module.UUID, module.Name, module.CreatedAt, module.UpdatedAt,
	).Scan(&module.ID)
}
