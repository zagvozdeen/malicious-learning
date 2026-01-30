package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
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
	Tags      []string  `json:"tags"`
	ModuleID  int       `json:"module_id"`
	CourseID  int       `json:"course_id"`
	IsActive  bool      `json:"is_active"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *Card) GetHash() string {
	hasher := sha256.New()
	hasher.Write([]byte(strconv.Itoa(c.ModuleID)))
	hasher.Write([]byte{0})
	hasher.Write([]byte(strconv.Itoa(c.CourseID)))
	hasher.Write([]byte{0})
	hasher.Write([]byte(c.Question))
	hasher.Write([]byte{0})
	hasher.Write([]byte(c.Answer))
	hasher.Write([]byte{0})
	hasher.Write([]byte(strings.Join(c.Tags, ",")))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *Store) GetAllCards(ctx context.Context) (cards []Card, err error) {
	var rows pgx.Rows
	rows, err = s.querier(ctx).Query(
		ctx,
		"SELECT id, uid, uuid, question, answer, module_id, is_active, hash, created_at, updated_at FROM cards WHERE is_active = TRUE ORDER BY uid",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func (s *Store) GetCards(ctx context.Context, courseSlug string, moduleIDs []int) ([]Card, error) {
	if len(moduleIDs) == 0 || courseSlug == "" {
		return nil, nil
	}
	pattern := make([]string, 0, len(moduleIDs))
	args := make([]any, 0, len(moduleIDs)+1)
	for i := range moduleIDs {
		pattern = append(pattern, fmt.Sprintf("$%d", i+1))
		args = append(args, moduleIDs[i])
	}
	args = append(args, courseSlug)
	sql := fmt.Sprintf(
		"SELECT c.id, c.uid, c.uuid, c.question, c.answer, c.module_id, c.is_active, c.hash, c.created_at, c.updated_at FROM cards c JOIN courses co ON co.id = c.course_id WHERE c.is_active = true AND c.module_id in (%s) AND co.slug = $%d ORDER BY c.uid",
		strings.Join(pattern, ","),
		len(args),
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

func (s *Store) IsExistsCardByUIDAndHash(ctx context.Context, uid int, hash string) (exists bool, err error) {
	err = s.querier(ctx).QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM cards WHERE uid = $1 AND hash = $2 AND is_active = TRUE)",
		uid, hash,
	).Scan(&exists)
	return
}

func (s *Store) CreateCard(ctx context.Context, card *Card) error {
	return s.querier(ctx).QueryRow(ctx, `
		INSERT INTO cards (uid, uuid, question, answer, tags, module_id, course_id, is_active, hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`,
		card.UID,
		card.UUID,
		card.Question,
		card.Answer,
		card.Tags,
		card.ModuleID,
		card.CourseID,
		card.IsActive,
		card.Hash,
		card.CreatedAt,
		card.UpdatedAt,
	).Scan(&card.ID)
}

func (s *Store) DeactivateCard(ctx context.Context, card *Card) (err error) {
	_, err = s.querier(ctx).Exec(
		ctx,
		"UPDATE cards SET is_active = FALSE, updated_at = $1 WHERE uid = $2 AND is_active = TRUE",
		card.UpdatedAt, card.UID,
	)
	return
}
