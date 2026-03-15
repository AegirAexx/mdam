// Package template provides template discovery and variable interpolation
// for mdam document creation.
package template

import (
	"fmt"
	"os"
	"path/filepath"
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

// Render interpolates variables in a template's content and returns the result.
// vars is a map of variable name (without braces) to value.
// Builtin variables (date, date_short) are resolved automatically.
// Any unresolved variables in the rendered output are returned as-is —
// the caller is responsible for prompting the user for missing values.
func Render(t Template, vars map[string]string) (string, error) {
	now := time.Now()
	content := t.Content

	// Resolve caller-supplied variables first so they take precedence over built-ins.
	for k, v := range vars {
		content = strings.ReplaceAll(content, "{{"+k+"}}", v)
	}

	// Resolve built-in variables for any remaining unresolved placeholders.
	builtins := map[string]string{
		"date":       now.UTC().Format(time.RFC3339),
		"date_short": now.Format("2006-01-02"),
	}
	for k, v := range builtins {
		content = strings.ReplaceAll(content, "{{"+k+"}}", v)
	}

	return content, nil
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
		"howto":   howtoTemplate,
		"meeting": meetingTemplate,
		"scratch": scratchTemplate,
	}
}

// WriteBuiltins writes built-in templates to the given directory, skipping any
// that already exist.
func WriteBuiltins(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating templates directory: %w", err)
	}
	for name, content := range BuiltinTemplates() {
		path := filepath.Join(dir, name+".md")
		if _, err := os.Stat(path); err == nil {
			continue // already exists
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing template %s: %w", name, err)
		}
	}
	return nil
}

var journalTemplate = `---
type: journal
title: {{date_short}}
tags: []
created: {{date_short}}
modified: {{date_short}}
---

# Journal — {{date_short}}

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

var howtoTemplate = `---
type: kb
title: {{title}}
tags: []
created: {{date_short}}
modified: {{date_short}}
kb_type: howto
---

# HowTo: {{title}}

## Prerequisites

## Steps

## Notes

`

var meetingTemplate = `---
type: kb
title: {{title}}
tags: []
created: {{date_short}}
modified: {{date_short}}
kb_type: meeting
---

# Meeting: {{title}}

**Date:** {{date_short}}

## Attendees

## Agenda

## Notes

## Actions

`

var scratchTemplate = `---
type: scratch
title: Scratch Pad
tags: []
created: {{date_short}}
modified: {{date_short}}
---

`
