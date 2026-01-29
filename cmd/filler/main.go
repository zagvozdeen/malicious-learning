package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/zagvozdeen/malicious-learning/internal/config"
	"github.com/zagvozdeen/malicious-learning/internal/logger"
)

const template = `---
name: 
module: 
tags:
  - грейды
---
`

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.New()
	log, stop := logger.New(cfg)
	defer stop()

	dir := flag.String("dir", "", "subfolder under data/courses to fill")
	count := flag.Int("count", 0, "number of templates to create")
	flag.Parse()

	log.Info("start", "dir", *dir, "count", *count)

	if err := run(ctx, log, *count, *dir); err != nil {
		log.Error("Run error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger, count int, dir string) error {
	if dir == "" {
		return fmt.Errorf("missing required -dir")
	}
	if count < 0 {
		return fmt.Errorf("-count must be >= 0")
	}

	basePath := filepath.Join("data", "courses", dir)
	log.Info("resolved path", "path", basePath)
	info, err := os.Stat(basePath)
	if err != nil {
		return fmt.Errorf("cannot access %s: %v", basePath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", basePath)
	}

	created := 0
	skipped := 0
	for i := range count {
		if ctx.Err() != nil {
			return fmt.Errorf("context error: %v", ctx.Err())
		}
		pattern := filepath.Join(basePath, fmt.Sprintf("%d_*.md", i+1))
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("bad glob pattern %s: %v", pattern, err)
		}
		if len(matches) > 0 {
			log.Info("skip existing", "index", i+1)
			skipped++
			continue
		}

		filename := filepath.Join(basePath, fmt.Sprintf("%d_.md", i+1))
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			if os.IsExist(err) {
				log.Info("skip existing", "index", i+1)
				skipped++
				continue
			}
			return fmt.Errorf("create %s: %v", filename, err)
		}
		if _, err := file.WriteString(template); err != nil {
			file.Close()
			return fmt.Errorf("write %s: %v", filename, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close %s: %v", filename, err)
		}
		log.Info("created", "file", filename)
		created++
	}

	log.Info("done", "created", created, "skipped", skipped)
	return nil
}
