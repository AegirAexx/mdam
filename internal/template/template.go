// Package template provides template discovery and variable interpolation
// for mdam document creation.
package template

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Template represents a discovered document template.
type Template struct {
	// Name is the template identifier (filename without extension).
	Name string
	// Path is the absolute path to the template file.
	Path string
	// Content is the raw template content including frontmatter.
	Content string
}

// BuiltinVars are the variable names resolved automatically at creation time.
var BuiltinVars = []string{
	"{{date}}",
	"{{date_short}}",
	"{{title}}",
	"{{author}}",
	"{{tags}}",
	"{{type}}",
}

// Discover scans dir for .md template files and returns them.
// Returns an empty slice (not an error) if the directory does not exist.
func Discover(dir string) ([]Template, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Template{}, nil
		}
		return nil, fmt.Errorf("reading templates directory: %w", err)
	}

	var templates []Template
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		tmplName := strings.TrimSuffix(name, ".md")
		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading template %s: %w", name, err)
		}
		templates = append(templates, Template{
			Name:    tmplName,
			Path:    path,
			Content: string(content),
		})
	}
	return templates, nil
}

// Find returns the named template from a directory, or an error if not found.
func Find(dir, name string) (Template, error) {
	templates, err := Discover(dir)
	if err != nil {
		return Template{}, err
	}
	for _, t := range templates {
		if t.Name == name {
			return t, nil
		}
	}
	return Template{}, fmt.Errorf("template %q not found in %s", name, dir)
}

// dateFormatRe matches {{date:FORMAT}} placeholders.
var dateFormatRe = regexp.MustCompile(`\{\{date:([^}]+)\}\}`)

// renderContent applies variable substitution to raw template content.
// Caller-supplied vars are applied first (highest precedence), then
// {{date:FORMAT}} patterns using Go's time layout syntax, then the
// fixed built-ins {{date}} and {{date_short}}.
func renderContent(content string, vars map[string]string, now time.Time) string {
	// 1. Caller-supplied variables take precedence over all built-ins.
	for k, v := range vars {
		content = strings.ReplaceAll(content, "{{"+k+"}}", v)
	}
	// 2. Parameterised date: {{date:FORMAT}} where FORMAT is a Go time layout.
	content = dateFormatRe.ReplaceAllStringFunc(content, func(m string) string {
		sub := dateFormatRe.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		return now.Format(sub[1])
	})
	// 3. Fixed built-ins for backwards compatibility.
	content = strings.ReplaceAll(content, "{{date}}", now.UTC().Format(time.RFC3339))
	content = strings.ReplaceAll(content, "{{date_short}}", now.Format("2006-01-02"))
	return content
}

// Render interpolates variables in a template's content and returns the result.
// vars is a map of variable name (without braces) to value.
// Built-in variables (date, date_short, date:FORMAT) are resolved automatically.
// Any unresolved variables in the rendered output are returned as-is —
// the caller is responsible for prompting the user for missing values.
func Render(t Template, vars map[string]string) (string, error) {
	return RenderAt(t, vars, time.Now())
}

// RenderAt is like Render but uses now as the reference time for all date
// variables. Use this when the document date differs from the current time
// (e.g. backdated journal entries).
func RenderAt(t Template, vars map[string]string, now time.Time) (string, error) {
	return renderContent(t.Content, vars, now), nil
}

// TemplateType extracts the value of the "type:" field from the frontmatter
// block of a template's content. Returns an empty string if the field is absent.
func TemplateType(content string) string {
	for _, line := range strings.SplitN(content, "\n", 30) {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "type:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "type:"))
		}
	}
	return ""
}

// UnresolvedVars returns all {{variable}} placeholders remaining in content
// that have not been substituted.
func UnresolvedVars(content string) []string {
	var found []string
	remaining := content
	for {
		start := strings.Index(remaining, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(remaining[start:], "}}")
		if end == -1 {
			break
		}
		varName := remaining[start : start+end+2]
		found = append(found, varName)
		remaining = remaining[start+end+2:]
	}
	return found
}

// BuiltinTemplates returns the set of built-in template contents keyed by name.
// These are used when no templates directory exists or to seed a new managed tree.
func BuiltinTemplates() map[string]string {
	return map[string]string{
		"journal": journalTemplate,
		"kb":      kbTemplate,
	}
}

// WriteBuiltins writes built-in templates to the given directory. An existing
// file is overwritten only when its content differs from the current built-in,
// so user-customised templates that match the built-in are left alone and
// stale copies from older versions are updated automatically.
func WriteBuiltins(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating templates directory: %w", err)
	}
	for name, content := range BuiltinTemplates() {
		path := filepath.Join(dir, name+".md")
		existing, err := os.ReadFile(path)
		if err == nil && string(existing) == content {
			continue // already up-to-date
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing template %s: %w", name, err)
		}
	}
	return nil
}

var journalTemplate = `---
type: journal
title: {{date:Monday - January 02 2006}}
tags: []
created: {{date_short}}
modified: {{date_short}}
---

# {{date:Monday - January 02 2006}}

## Notes

## TODOs

- [ ]
`

var kbTemplate = `---
type: kb
title: {{title}}
tags: []
created: {{date_short}}
modified: {{date_short}}
---

# {{title}}

`

