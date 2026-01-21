package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Card struct {
	ID        int       `json:"id"`
	UID       int       `json:"uid"`
	UUID      string    `json:"uuid"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	ModuleID  int       `json:"module_id"`
	IsActive  bool      `json:"is_active"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Store) GetCards(ctx context.Context, moduleIDs []int) ([]Card, error) {
	if len(moduleIDs) == 0 {
		return nil, nil
	}
	pattern := make([]string, 0, len(moduleIDs))
	args := make([]any, 0, len(moduleIDs))
	for i := range moduleIDs {
		pattern = append(pattern, fmt.Sprintf("$%d", i+1))
		args = append(args, moduleIDs[i])
	}
	sql := fmt.Sprintf(
		"SELECT id, uid, uuid, question, answer, module_id, is_active, hash, created_at, updated_at FROM cards WHERE is_active = true AND module_id in (%s) ORDER BY uid",
		strings.Join(pattern, ","),
	)
	rows, err := s.querier(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := make([]Card, 0)
	for rows.Next() {
		var card Card
		err = rows.Scan(
			&card.ID,
			&card.UID,
			&card.UUID,
			&card.Question,
			&card.Answer,
			&card.ModuleID,
			&card.IsActive,
			&card.Hash,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return cards, nil
}

func (s *Store) GetCardByUIDAndHash(ctx context.Context, uid int, hash string) (*Card, error) {
	var card Card
	err := s.querier(ctx).QueryRow(ctx, `
		SELECT id, uid, uuid, question, answer, module_id, is_active, hash, created_at, updated_at
		FROM cards
		WHERE uid = $1 AND hash = $2
		LIMIT 1
	`, uid, hash).Scan(
		&card.ID,
		&card.UID,
		&card.UUID,
		&card.Question,
		&card.Answer,
		&card.ModuleID,
		&card.IsActive,
		&card.Hash,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (s *Store) GetActiveCardByUID(ctx context.Context, uid int) (*Card, error) {
	var card Card
	err := s.querier(ctx).QueryRow(ctx, `
		SELECT id, uid, uuid, question, answer, module_id, is_active, hash, created_at, updated_at
		FROM cards
		WHERE uid = $1 AND is_active = true
		ORDER BY id DESC
		LIMIT 1
	`, uid).Scan(
		&card.ID,
		&card.UID,
		&card.UUID,
		&card.Question,
		&card.Answer,
		&card.ModuleID,
		&card.IsActive,
		&card.Hash,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (s *Store) CreateCard(ctx context.Context, card *Card) error {
	return s.querier(ctx).QueryRow(ctx, `
		INSERT INTO cards (uid, uuid, question, answer, module_id, is_active, hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`,
		card.UID,
		card.UUID,
		card.Question,
		card.Answer,
		card.ModuleID,
		card.IsActive,
		card.Hash,
		card.CreatedAt,
		card.UpdatedAt,
	).Scan(&card.ID)
}

func (s *Store) DeactivateCardByID(ctx context.Context, id int, updatedAt time.Time) error {
	tag, err := s.querier(ctx).Exec(ctx, `
		UPDATE cards
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`, updatedAt, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
