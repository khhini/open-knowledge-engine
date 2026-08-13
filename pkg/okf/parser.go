package okf

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"gopkg.in/yaml.v3"
)

var markdownParser = goldmark.New(
	goldmark.WithExtensions(extension.Table, extension.GFM),
)

func ParseConcept(fullPath string) (*Concept, error) {
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read concept file in %s: %w", fullPath, err)
	}

	parts := bytes.SplitN(content, []byte("\n---"), 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid frontmatter in %s", fullPath)
	}

	frontmatterBytes := bytes.TrimPrefix(parts[0], []byte("---\n"))
	bodyMarkdown := string(parts[1])

	var fm Frontmatter
	if err := yaml.Unmarshal(frontmatterBytes, &fm); err != nil {
		return nil, fmt.Errorf("failed formating frontmatter to yaml: %w", err)
	}

	if fm.Type == "" {
		return nil, fmt.Errorf("missing type key in %s", fullPath)
	}

	// Render Markdown body to HTML
	var htmlBuf bytes.Buffer
	if err := markdownParser.Convert([]byte(bodyMarkdown), &htmlBuf); err != nil {
		htmlBuf.WriteString(bodyMarkdown) // Fallback to raw text if error
	}

	conceptID := strings.TrimSuffix(fullPath, ".md")
	trustTier := EvaluateTrustTier(&fm, time.Now())

	return &Concept{
		ID:           conceptID,
		FilePath:     fullPath,
		Frontmatter:  fm,
		BodyMarkdown: bodyMarkdown,
		BodyHTML:     template.HTML(htmlBuf.String()),
		TrustTier:    string(trustTier),
		IsStale:      trustTier == TierStale,
	}, nil
}

func FormatConcept(fm *Frontmatter, body string) ([]byte, error) {
	if fm == nil {
		return nil, fmt.Errorf("frontmatter cannot be nil")
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(fm); err != nil {
		encoder.Close()
		return nil, fmt.Errorf("failed to marshal frontmatter: %w", err)
	}
	encoder.Close()

	buf.WriteString("---\n\n")
	buf.WriteString(body)

	if len(body) > 0 && body[len(body)-1] != '\n' {
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}
