package converter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zagvozdeen/malicious-learning/data"
	"github.com/zagvozdeen/malicious-learning/internal/store"
	"gopkg.in/yaml.v3"
)

type CourseDescription struct {
	Name string `yaml:"name"`
}

type CardDescription struct {
	Name   string   `yaml:"name"`
	Module string   `yaml:"module"`
	Tags   []string `yaml:"tags"`
}

func Run(ctx context.Context, storage store.Storage) error {
	entries, err := data.Courses.ReadDir("courses")
	if err != nil {
		return fmt.Errorf("failed to read courses dir: %w", err)
	}
	head := &strings.Builder{}
	var hlighter *highlighter
	hlighter, err = newHighlighter(head)
	if err != nil {
		return fmt.Errorf("failed to create highlighter: %w", err)
	}
	ctx, err = storage.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer storage.Rollback(ctx)
	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("found file in courses dir: %s", entry.Name())
		}
		var course *store.Course
		course, err = storage.GetCourseBySlug(ctx, entry.Name())
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("failed to get course by slug: %w", err)
			}
			var b []byte
			b, err = data.Courses.ReadFile(fmt.Sprintf("courses/%s/0_index.yaml", entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to read 0_index.yaml file in %s: %w", entry.Name(), err)
			}
			cd := &CourseDescription{}
			err = yaml.Unmarshal(b, cd)
			if err != nil {
				return fmt.Errorf("failed to unmarshal yaml to struct: %w", err)
			}
			var uid uuid.UUID
			uid, err = uuid.NewV7()
			if err != nil {
				return fmt.Errorf("failed to generate course uuid: %w", err)
			}
			course = &store.Course{
				UUID:      uid.String(),
				Slug:      entry.Name(),
				Name:      cd.Name,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err = storage.CreateCourse(ctx, course)
			if err != nil {
				return fmt.Errorf("failed to create course: %w", err)
			}
		}
		var cardEntries []fs.DirEntry
		dirName := fmt.Sprintf("courses/%s", entry.Name())
		cardEntries, err = data.Courses.ReadDir(dirName)
		if err != nil {
			return fmt.Errorf("failed to read dir %q: %w", dirName, err)
		}
		for _, cardEntry := range cardEntries {
			if cardEntry.IsDir() {
				return fmt.Errorf("dir %s haves a dir %q", dirName, cardEntry.Name())
			}
			if cardEntry.Name() == "0_index.yaml" {
				continue
			}
			if path.Ext(cardEntry.Name()) != ".md" {
				return fmt.Errorf("dir %s haves not markdown file %s", dirName, cardEntry.Name())
			}
			parts := strings.SplitN(cardEntry.Name(), "_", 2)
			if len(parts) != 2 {
				return fmt.Errorf("failed to split card %q", cardEntry.Name())
			}
			var id int
			id, err = strconv.Atoi(parts[0])
			if err != nil {
				return fmt.Errorf("failed to parse card id: %w", err)
			}
			var b []byte
			fileName := fmt.Sprintf("courses/%s/%s", entry.Name(), cardEntry.Name())
			b, err = data.Courses.ReadFile(fileName)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", fileName, err)
			}
			cd := &CardDescription{}
			b, err = frontmatter.Parse(bytes.NewReader(b), cd, frontmatter.NewFormat("---", "---", yaml.Unmarshal))
			if err != nil {
				return fmt.Errorf("failed to parse front-matter: %w", err)
			}
			cd.Name = strings.TrimSpace(cd.Name)
			cd.Module = strings.TrimSpace(cd.Module)
			if cd.Module == "" || cd.Name == "" {
				return fmt.Errorf("failed to parse card description %q: module or name is empty", cardEntry.Name())
			}
			extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
			p := parser.NewWithExtensions(extensions)
			doc := p.Parse(b)
			renderer := html.NewRenderer(html.RendererOptions{
				Flags: html.CommonFlags | html.HrefTargetBlank,
				RenderNodeHook: func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
					if _, ok := node.(*ast.Table); ok {
						if entering {
							_, _ = io.WriteString(w, `<div class="table-wrapper"><table>`)
						} else {
							_, _ = io.WriteString(w, `</table></div>`)
						}
						return ast.GoToNext, true
					}
					if code, ok := node.(*ast.CodeBlock); ok {
						err = hlighter.highlight(w, string(code.Literal), string(code.Info))
						if err != nil {
							return ast.GoToNext, false
						}
						return ast.GoToNext, true
					}
					return ast.GoToNext, false
				},
			})
			xml := markdown.Render(doc, renderer)
			var module *store.Module
			module, err = storage.GetModuleByName(ctx, cd.Module)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					return fmt.Errorf("failed to get module by name: %w", err)
				}
				var uid uuid.UUID
				uid, err = uuid.NewV7()
				if err != nil {
					return fmt.Errorf("failed to generate module uuid: %w", err)
				}
				module = &store.Module{
					UUID:      uid.String(),
					Name:      cd.Module,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				err = storage.CreateModule(ctx, module)
				if err != nil {
					return fmt.Errorf("failed to create module: %w", err)
				}
			}
			var uid uuid.UUID
			uid, err = uuid.NewV7()
			if err != nil {
				return fmt.Errorf("failed to generate card uuid: %w", err)
			}
			card := &store.Card{
				UID:       id,
				UUID:      uid.String(),
				Question:  cd.Name,
				Answer:    string(xml) + head.String(),
				Tags:      cd.Tags,
				ModuleID:  module.ID,
				CourseID:  course.ID,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			card.Hash = card.GetHash()
			var exists bool
			exists, err = storage.IsExistsCardByUIDAndHash(ctx, card.UID, card.Hash)
			if err != nil {
				return fmt.Errorf("failed to check existing card: %w", err)
			}
			if exists {
				continue
			}
			err = storage.DeactivateCard(ctx, card)
			if err != nil {
				return fmt.Errorf("failed to deactivate card: %w", err)
			}
			err = storage.CreateCard(ctx, card)
			if err != nil {
				return fmt.Errorf("failed to create card: %w", err)
			}
		}
	}
	storage.Commit(ctx)
	return nil
}
