package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const dir = "/path/to/files"
const date = "2020-01-01"

type datedFile struct {
	name string
	date time.Time
}

func main() {
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Fatalf("invalid date %q: %v", date, err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("read dir %q: %v", dir, err)
	}

	files := make([]datedFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		base := strings.TrimSuffix(name, ".md")
		fileDate, err := time.Parse("2006-01-02", base)
		if err != nil {
			continue
		}

		if fileDate.Before(targetDate) {
			continue
		}

		files = append(files, datedFile{name: name, date: fileDate})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].date.Equal(files[j].date) {
			return files[i].name < files[j].name
		}
		return files[i].date.Before(files[j].date)
	})

	var out strings.Builder
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(dir, file.name))
		if err != nil {
			log.Fatalf("read file %q: %v", file.name, err)
		}

		out.WriteString("# ")
		out.WriteString(file.date.Format("2006-01-02"))
		out.WriteString("\n\n")
		out.Write(content)
		out.WriteString("\n\n\n\n")
	}

	if err := os.WriteFile("history.md", []byte(out.String()), 0o644); err != nil {
		log.Fatalf("write output: %v", err)
	}
}
