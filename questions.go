package malicious_learning

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zagvozdeen/malicious-learning/internal/store"
	"github.com/zagvozdeen/malicious-learning/internal/store/models"
)

//go:embed questions.json
var s embed.FS

type question struct {
	UID      uint16 `json:"id"`
	Module   string `json:"module"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// ParseQuestions читает JSON из ./questions.json
// Unmarshall в []question структуру
// по полям (module, question, answer) считает хэш (выбрать оптимальный хэш для этого)
// в БД ищет такое же значение по `uid=? and hash=?`
// если такое значение нашлось, то ничего не делаем, даныные в актуальном состоянии
// если такого значения нет, но находится по `uid=? and is_active=true`, то старое нужно пометить как is_active=false и создать новое с is_active=true
// если такого значения нет даже по uid и is_active, то просто создать новую строчку
func ParseQuestions(ctx context.Context, store store.Storage) error {
	data, err := s.ReadFile("questions.json")
	if err != nil {
		return err
	}

	var questions []question
	if err = json.Unmarshal(data, &questions); err != nil {
		return err
	}

	for _, q := range questions {
		moduleName := strings.TrimSpace(q.Module)
		hash := questionHash(moduleName, q.Question, q.Answer)
		_, err = store.GetCardByUIDAndHash(ctx, int(q.UID), hash)
		if err == nil {
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		module, err := store.GetModuleByName(ctx, moduleName)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			moduleUUID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			now := time.Now()
			module = &models.Module{
				UUID:      moduleUUID.String(),
				Name:      moduleName,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err = store.CreateModule(ctx, module); err != nil {
				return err
			}
		}

		now := time.Now()
		activeCard, err := store.GetActiveCardByUID(ctx, int(q.UID))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err == nil && activeCard != nil {
			if err := store.DeactivateCardByID(ctx, activeCard.ID, now); err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}

		cardUUID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		card := &models.Card{
			UID:       int(q.UID),
			UUID:      cardUUID.String(),
			Question:  q.Question,
			Answer:    q.Answer,
			ModuleID:  module.ID,
			IsActive:  true,
			Hash:      hash,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := store.CreateCard(ctx, card); err != nil {
			return err
		}
	}
	return nil
}

func questionHash(module, question, answer string) string {
	hasher := sha256.New()
	hasher.Write([]byte(module))
	hasher.Write([]byte{0})
	hasher.Write([]byte(question))
	hasher.Write([]byte{0})
	hasher.Write([]byte(answer))
	return hex.EncodeToString(hasher.Sum(nil))
}
