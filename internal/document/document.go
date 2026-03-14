// Package document provides the core markdown document model and frontmatter
// parsing for mdam-managed files.
package document

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ValidTypes is the set of allowed document types.
var ValidTypes = map[string]bool{
	"journal":  true,
	"kb":       true,
	"todo":     true,
	"scratch":  true,
	"unsorted": true,
}

// Frontmatter holds the required and optional YAML fields for a managed document.
type Frontmatter struct {
	Title    string    `yaml:"title"`
	Tags     []string  `yaml:"tags"`
	Created  time.Time `yaml:"created"`
	Modified time.Time `yaml:"modified"`
	Type     string    `yaml:"type"`
	// Extra captures any additional frontmatter fields.
	Extra map[string]interface{} `yaml:"-"`
}

// Document represents a parsed managed markdown file.
type Document struct {
	Path        string
	Frontmatter Frontmatter
	Body        string
}

var (
	// kebabRe matches valid kebab-case filenames (letters, digits, hyphens).
	kebabRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	// journalRe matches YYYY-MM-DD journal filenames.
	journalRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// ValidateFilename checks that a filename (without extension) is kebab-case,
// POSIX/URL-safe. Journal filenames (YYYY-MM-DD) are also accepted.
func ValidateFilename(name string) error {
	if journalRe.MatchString(name) {
		return nil
	}
	if !kebabRe.MatchString(name) {
		return fmt.Errorf("filename %q is not kebab-case (only lowercase letters, digits, and hyphens allowed)", name)
	}
	return nil
}

// ToKebabCase converts a string to kebab-case by lowercasing and replacing
// non-alphanumeric runs with hyphens.
func ToKebabCase(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// ParseFile reads and parses a markdown document at the given path.
func ParseFile(path string) (Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Document{}, fmt.Errorf("reading file: %w", err)
	}
	fm, body, err := parseFrontmatter(data)
	if err != nil {
		return Document{}, fmt.Errorf("parsing frontmatter in %s: %w", path, err)
	}
	return Document{
		Path:        path,
		Frontmatter: fm,
		Body:        body,
	}, nil
}

// parseFrontmatter splits raw file content into Frontmatter and body.
// Returns an error if the frontmatter delimiters are missing or YAML is invalid.
func parseFrontmatter(data []byte) (Frontmatter, string, error) {
	content := string(data)

	if !strings.HasPrefix(content, "---") {
		return Frontmatter{}, "", fmt.Errorf("missing opening frontmatter delimiter")
	}

	// Find the closing delimiter (skip the opening "---\n").
	rest := content[3:]
	// Allow for \r\n or \n
	rest = strings.TrimLeft(rest, "\r\n")
	// Actually we need to find the next "---" on its own line.
	// Re-approach: split on lines.
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return Frontmatter{}, "", fmt.Errorf("missing closing frontmatter delimiter")
	}

	// lines[0] should be "---"
	closing := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			closing = i
			break
		}
	}
	if closing == -1 {
		return Frontmatter{}, "", fmt.Errorf("missing closing frontmatter delimiter")
	}

	yamlContent := strings.Join(lines[1:closing], "\n")
	bodyLines := lines[closing+1:]
	body := strings.TrimLeft(strings.Join(bodyLines, "\n"), "\r\n")

	// Parse into a raw map first to capture extra fields.
	raw := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		return Frontmatter{}, "", fmt.Errorf("invalid YAML: %w", err)
	}

	// Parse into typed struct.
	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return Frontmatter{}, "", fmt.Errorf("invalid YAML: %w", err)
	}

	// Collect extra fields.
	known := map[string]bool{
		"title": true, "tags": true, "created": true,
		"modified": true, "type": true,
	}
	extra := map[string]interface{}{}
	for k, v := range raw {
		if !known[k] {
			extra[k] = v
		}
	}
	if len(extra) > 0 {
		fm.Extra = extra
	}

	return fm, body, nil
}

// ValidateFrontmatter checks that all required fields are present and valid.
func ValidateFrontmatter(fm Frontmatter) error {
	if strings.TrimSpace(fm.Title) == "" {
		return fmt.Errorf("missing required field: title")
	}
	if fm.Tags == nil {
		return fmt.Errorf("missing required field: tags")
	}
	if fm.Created.IsZero() {
		return fmt.Errorf("missing required field: created")
	}
	if fm.Modified.IsZero() {
		return fmt.Errorf("missing required field: modified")
	}
	if strings.TrimSpace(fm.Type) == "" {
		return fmt.Errorf("missing required field: type")
	}
	if !ValidTypes[fm.Type] {
		return fmt.Errorf("invalid type %q: must be one of journal, kb, todo, scratch, unsorted", fm.Type)
	}
	return nil
}

// RenderFrontmatter serialises a Frontmatter back to YAML between --- delimiters.
func RenderFrontmatter(fm Frontmatter) (string, error) {
	// Build ordered map to get predictable field order.
	type fmYAML struct {
		Title    string                 `yaml:"title"`
		Tags     []string               `yaml:"tags"`
		Created  time.Time              `yaml:"created"`
		Modified time.Time              `yaml:"modified"`
		Type     string                 `yaml:"type"`
		Extra    map[string]interface{} `yaml:",inline"`
	}
	out := fmYAML{
		Title:    fm.Title,
		Tags:     fm.Tags,
		Created:  fm.Created,
		Modified: fm.Modified,
		Type:     fm.Type,
		Extra:    fm.Extra,
	}
	b, err := yaml.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("marshalling frontmatter: %w", err)
	}
	return "---\n" + string(b) + "---\n", nil
}

// Write serialises the document (frontmatter + body) and writes it to doc.Path.
func (d Document) Write() error {
	header, err := RenderFrontmatter(d.Frontmatter)
	if err != nil {
		return fmt.Errorf("rendering frontmatter: %w", err)
	}
	content := header + "\n" + d.Body
	if err := os.MkdirAll(filepath.Dir(d.Path), 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	return os.WriteFile(d.Path, []byte(content), 0o644)
}
