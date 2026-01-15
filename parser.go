package malicious_learning

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type questionFrontmatter struct {
	Question string `yaml:"question"`
	Module   string `yaml:"module"`
}

// ParseQuestionsToMarkdown converts the embedded questions.json into Markdown files under ./questions.
func ParseQuestionsToMarkdown() error {
	data, err := s.ReadFile("questions.json")
	if err != nil {
		return err
	}

	var questions []question
	if err := json.Unmarshal(data, &questions); err != nil {
		return err
	}

	//outputDir := "questions"
	//if err := os.MkdirAll(outputDir, 0o755); err != nil {
	//	return err
	//}

	for _, q := range questions {
		if err := writeQuestionMarkdown("questions", q); err != nil {
			return err
		}
	}

	return nil
}

func writeQuestionMarkdown(outputDir string, q question) error {
	frontmatter := questionFrontmatter{
		Question: q.Question,
		Module:   strings.TrimSpace(q.Module),
	}

	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return err
	}

	var content bytes.Buffer
	content.WriteString("---\n")
	content.Write(yamlData)
	content.WriteString("---\n\n")
	content.WriteString(q.Answer)
	if !strings.HasSuffix(q.Answer, "\n") {
		content.WriteString("\n")
	}

	fileName := fmt.Sprintf("%d.md", q.UID)
	filePath := filepath.Join(outputDir, fileName)
	return os.WriteFile(filePath, content.Bytes(), 0o644)
}
