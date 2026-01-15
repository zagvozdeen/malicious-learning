package malicious_learning

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/zagvozdeen/malicious-learning/internal/store"
	"github.com/zagvozdeen/malicious-learning/internal/store/models"
	"gopkg.in/yaml.v3"
)

//go:embed questions.json
var s embed.FS

type question struct {
	UID      uint16 `json:"id"`
	Module   string `json:"module"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// ParseQuestions читает Markdown из ./questions.
// В каждом файле должен быть YAML frontmatter (question, module), а тело — Markdown-ответ.
// По полям (module, question, answer) считает хэш (выбрать оптимальный хэш для этого)
// в БД ищет такое же значение по `uid=? and hash=?`
// если такое значение нашлось, то ничего не делаем, даныные в актуальном состоянии
// если такого значения нет, но находится по `uid=? and is_active=true`, то старое нужно пометить как is_active=false и создать новое с is_active=true
// если такого значения нет даже по uid и is_active, то просто создать новую строчку
func ParseQuestions(ctx context.Context, store store.Storage) error {
	questions, err := loadQuestionsFromMarkdown("questions")
	if err != nil {
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

var markdownParser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

func loadQuestionsFromMarkdown(dir string) ([]question, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}

	sort.Strings(fileNames)

	questions := make([]question, 0, len(fileNames))
	for _, name := range fileNames {
		path := filepath.Join(dir, name)
		q, err := parseQuestionMarkdown(path)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		questions = append(questions, q)
	}

	return questions, nil
}

func parseQuestionMarkdown(path string) (question, error) {
	fileName := filepath.Base(path)
	ext := filepath.Ext(fileName)
	if ext != ".md" {
		return question{}, fmt.Errorf("unsupported extension: %s", ext)
	}

	idText := strings.TrimSuffix(fileName, ext)
	uid, err := strconv.ParseUint(idText, 10, 16)
	if err != nil {
		return question{}, fmt.Errorf("invalid question id %q: %w", idText, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return question{}, err
	}

	frontmatterBytes, bodyBytes, err := splitFrontmatter(data)
	if err != nil {
		return question{}, err
	}

	var frontmatter questionFrontmatter
	if err := yaml.Unmarshal(frontmatterBytes, &frontmatter); err != nil {
		return question{}, err
	}
	if strings.TrimSpace(frontmatter.Question) == "" {
		return question{}, errors.New("frontmatter question is empty")
	}
	if strings.TrimSpace(frontmatter.Module) == "" {
		return question{}, errors.New("frontmatter module is empty")
	}

	var rendered bytes.Buffer
	if err := markdownParser.Convert(bodyBytes, &rendered); err != nil {
		return question{}, err
	}

	return question{
		UID:      uint16(uid),
		Module:   frontmatter.Module,
		Question: frontmatter.Question,
		Answer:   rendered.String(),
	}, nil
}

func splitFrontmatter(data []byte) ([]byte, []byte, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, nil, err
	}
	if strings.TrimSpace(line) != "---" {
		return nil, nil, errors.New("frontmatter must start with ---")
	}

	var frontmatter bytes.Buffer
	for {
		line, err = reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, nil, err
		}
		if strings.TrimSpace(line) == "---" {
			break
		}
		frontmatter.WriteString(line)
		if errors.Is(err, io.EOF) {
			return nil, nil, errors.New("frontmatter must end with ---")
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, err
	}

	return frontmatter.Bytes(), body, nil
}
