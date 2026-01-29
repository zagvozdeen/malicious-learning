package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const template = `---
name: 
module: 
tags: 
---

`

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	dir := flag.String("dir", "", "subfolder under data/courses to fill")
	count := flag.Int("count", 0, "number of templates to create")
	flag.Parse()

	logger.Info("start", "dir", *dir, "count", *count)

	if *dir == "" {
		exitf("missing required -dir")
	}
	if *count < 0 {
		exitf("-count must be >= 0")
	}

	basePath := filepath.Join("data", "courses", *dir)
	logger.Info("resolved path", "path", basePath)
	info, err := os.Stat(basePath)
	if err != nil {
		exitf("cannot access %s: %v", basePath, err)
	}
	if !info.IsDir() {
		exitf("%s is not a directory", basePath)
	}

	created := 0
	skipped := 0
	for i := 0; i < *count; i++ {
		index := i + 1
		pattern := filepath.Join(basePath, fmt.Sprintf("%d_*.md", index))
		matches, err := filepath.Glob(pattern)
		if err != nil {
			exitf("bad glob pattern %s: %v", pattern, err)
		}
		if len(matches) > 0 {
			logger.Info("skip existing", "index", index)
			skipped++
			continue
		}

		filename := filepath.Join(basePath, fmt.Sprintf("%d_.md", index))
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			if os.IsExist(err) {
				logger.Info("skip existing", "index", index)
				skipped++
				continue
			}
			exitf("create %s: %v", filename, err)
		}
		if _, err := file.WriteString(template); err != nil {
			file.Close()
			exitf("write %s: %v", filename, err)
		}
		if err := file.Close(); err != nil {
			exitf("close %s: %v", filename, err)
		}
		logger.Info("created", "file", filename)
		created++
	}

	logger.Info("done", "created", created, "skipped", skipped)
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
