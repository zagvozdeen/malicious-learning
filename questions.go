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
	"path"
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
	"gopkg.in/yaml.v3"
)

//go:embed questions/*.md
var s embed.FS

type question struct {
	UID      uint16 `json:"id"`
	Module   string `json:"module"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type questionFrontmatter struct {
	Question string `yaml:"question"`
	Module   string `yaml:"module"`
}

// ParseQuestions читает Markdown из ./questions.
// В каждом файле должен быть YAML frontmatter (question, module), а тело — Markdown-ответ.
// По полям (module, question, answer) считает хэш (выбрать оптимальный хэш для этого)
// в БД ищет такое же значение по `uid=? and hash=?`
// если такое значение нашлось, то ничего не делаем, даныные в актуальном состоянии
// если такого значения нет, но находится по `uid=? and is_active=true`, то старое нужно пометить как is_active=false и создать новое с is_active=true
// если такого значения нет даже по uid и is_active, то просто создать новую строчку
func ParseQuestions(ctx context.Context, storage store.Storage) error {
	questions, err := loadQuestionsFromMarkdown("questions")
	if err != nil {
		return err
	}

	for _, q := range questions {
		moduleName := strings.TrimSpace(q.Module)
		hash := questionHash(moduleName, q.Question, q.Answer)
		_, err = storage.GetCardByUIDAndHash(ctx, int(q.UID), hash)
		if err == nil {
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		module, err := storage.GetModuleByName(ctx, moduleName)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			moduleUUID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			now := time.Now()
			module = &store.Module{
				UUID:      moduleUUID.String(),
				Name:      moduleName,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err = storage.CreateModule(ctx, module); err != nil {
				return err
			}
		}

		now := time.Now()
		activeCard, err := storage.GetActiveCardByUID(ctx, int(q.UID))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err == nil && activeCard != nil {
			if err := storage.DeactivateCardByID(ctx, activeCard.ID, now); err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}

		cardUUID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		card := &store.Card{
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
		if err := storage.CreateCard(ctx, card); err != nil {
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
	entries, err := s.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fileNames := make([]string, len(entries))
	for i, entry := range entries {
		if !entry.IsDir() && path.Ext(entry.Name()) == ".md" {
			fileNames[i] = entry.Name()
		}
	}

	sort.Strings(fileNames)

	questions := make([]question, 0, len(fileNames))
	for _, name := range fileNames {
		filePath := path.Join(dir, name)
		data, err := s.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", filePath, err)
		}
		q, err := parseQuestionMarkdown(filePath, data)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filePath, err)
		}
		questions = append(questions, q)
	}

	return questions, nil
}

func parseQuestionMarkdown(filePath string, data []byte) (question, error) {
	fileName := path.Base(filePath)
	ext := path.Ext(fileName)
	if ext != ".md" {
		return question{}, fmt.Errorf("unsupported extension: %s", ext)
	}

	idText := strings.TrimSuffix(fileName, ext)
	uid, err := strconv.ParseUint(idText, 10, 16)
	if err != nil {
		return question{}, fmt.Errorf("invalid question id %q: %w", idText, err)
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

	questionHTML, err := renderMarkdown(frontmatter.Question)
	if err != nil {
		return question{}, err
	}

	var rendered bytes.Buffer
	if err := markdownParser.Convert(bodyBytes, &rendered); err != nil {
		return question{}, err
	}

	return question{
		UID:      uint16(uid),
		Module:   frontmatter.Module,
		Question: questionHTML,
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

func renderMarkdown(input string) (string, error) {
	var rendered bytes.Buffer
	if err := markdownParser.Convert([]byte(input), &rendered); err != nil {
		return "", err
	}
	return rendered.String(), nil
}
