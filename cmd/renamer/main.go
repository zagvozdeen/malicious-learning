package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/adrg/frontmatter"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/zagvozdeen/malicious-learning/data"
	"github.com/zagvozdeen/malicious-learning/internal/config"
	"github.com/zagvozdeen/malicious-learning/internal/logger"
	"gopkg.in/yaml.v3"
)

type CardDescription struct {
	Name string `yaml:"name"`
}

var (
	filePattern  = regexp.MustCompile(`^(\d+)_\.md$`)
	slugTokenRe  = regexp.MustCompile(`[a-z0-9]+(?:_[a-z0-9]+)*`)
	slugCleanRe  = regexp.MustCompile(`[^a-z0-9_]+`)
	promptPrefix = `Есть вопрос с собеседования, твоя задача написать название файла на английском языке, которое в 1-5 слов характеризовало вопрос в формате snake_case. Пример: "Что такое Go?", твой ответ должен быть таким: "what_is_go"`
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.New()
	log, stop := logger.New(cfg)
	defer stop()

	if err := run(ctx, cfg, log); err != nil {
		log.Error("renamer failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config, log *slog.Logger) error {
	if cfg.NeuroAPI == "" || cfg.NeuroToken == "" {
		return errors.New("missing NEURO_API or NEURO_TOKEN env")
	}

	client := openai.NewClient(
		option.WithBaseURL(cfg.NeuroAPI),
		option.WithAPIKey(cfg.NeuroToken),
		//option.WithDebugLog(slog.NewLogLogger(log.Handler(), slog.LevelDebug)),
	)

	log.Info("start renamer")
	entries, err := data.Courses.ReadDir("courses")
	if err != nil {
		return fmt.Errorf("failed to read courses dir: %w", err)
	}

	renamed := 0
	skipped := 0
	for _, entry := range entries {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !entry.IsDir() {
			log.Warn("skip non-dir entry", "entry", entry.Name())
			continue
		}
		courseSlug := entry.Name()
		embeddedDir := path.Join("courses", courseSlug)
		diskDir := filepath.Join("data", "courses", courseSlug)
		log.Info("processing course", "course", courseSlug)

		cardEntries, err := data.Courses.ReadDir(embeddedDir)
		if err != nil {
			return fmt.Errorf("failed to read dir %q: %w", embeddedDir, err)
		}
		for _, cardEntry := range cardEntries {
			//if renamed > 20 {
			//	log.Info("STOPPING RENAME", "entry", entry.Name())
			//	return nil
			//}

			if ctx.Err() != nil {
				return ctx.Err()
			}
			if cardEntry.IsDir() {
				log.Debug("skip nested dir", "course", courseSlug, "entry", cardEntry.Name())
				continue
			}
			name := cardEntry.Name()
			if name == "0_index.yaml" {
				continue
			}
			matches := filePattern.FindStringSubmatch(name)
			if len(matches) == 0 {
				log.Debug("skip non-matching file", "course", courseSlug, "file", name)
				continue
			}
			id := matches[1]

			embeddedPath := path.Join(embeddedDir, name)
			content, err := data.Courses.ReadFile(embeddedPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", embeddedPath, err)
			}
			cd := &CardDescription{}
			if _, err := frontmatter.Parse(bytes.NewReader(content), cd, frontmatter.NewFormat("---", "---", yaml.Unmarshal)); err != nil {
				log.Error("failed to parse front-matter", "course", courseSlug, "file", name, "err", err)
				skipped++
				continue
			}
			question := strings.TrimSpace(cd.Name)
			if question == "" {
				log.Info("skip empty name", "course", courseSlug, "file", name)
				skipped++
				continue
			}

			log.Info("requesting slug", "course", courseSlug, "file", name, "question", question)
			slug, err := requestSlug(ctx, client, question)
			if err != nil {
				log.Error("failed to request slug", "course", courseSlug, "file", name, "err", err)
				skipped++
				continue
			}
			if slug == "" {
				log.Error("empty slug from model", "course", courseSlug, "file", name)
				skipped++
				continue
			}

			newName := fmt.Sprintf("%s_%s.md", id, slug)
			oldPath := filepath.Join(diskDir, name)
			newPath := filepath.Join(diskDir, newName)
			if oldPath == newPath {
				log.Info("skip already renamed", "course", courseSlug, "file", name)
				skipped++
				continue
			}
			if _, err := os.Stat(newPath); err == nil {
				log.Warn("target file exists, skip", "course", courseSlug, "file", name, "target", newName)
				skipped++
				continue
			} else if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("failed to stat target %s: %w", newPath, err)
			}

			if err := os.Rename(oldPath, newPath); err != nil {
				log.Error("failed to rename", "course", courseSlug, "from", oldPath, "to", newPath, "err", err)
				skipped++
				continue
			}
			renamed++
			log.Info("renamed", "course", courseSlug, "from", name, "to", newName)
		}
	}
	log.Info("done", "renamed", renamed, "skipped", skipped)
	return nil
}

func requestSlug(ctx context.Context, client openai.Client, question string) (string, error) {
	prompt := fmt.Sprintf("%s\nВопрос: %q", promptPrefix, question)
	res, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT5Mini,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}
	if len(res.Choices) == 0 {
		return "", errors.New("empty model choices")
	}
	raw := strings.TrimSpace(res.Choices[0].Message.Content)
	return sanitizeSlug(raw), nil
}

func sanitizeSlug(raw string) string {
	clean := strings.TrimSpace(raw)
	clean = strings.Trim(clean, "`\"'")
	clean = strings.ToLower(clean)
	if match := slugTokenRe.FindString(clean); match != "" {
		clean = match
	}
	clean = strings.ReplaceAll(clean, "-", "_")
	clean = strings.ReplaceAll(clean, " ", "_")
	clean = slugCleanRe.ReplaceAllString(clean, "_")
	for strings.Contains(clean, "__") {
		clean = strings.ReplaceAll(clean, "__", "_")
	}
	clean = strings.Trim(clean, "_")
	return clean
}
