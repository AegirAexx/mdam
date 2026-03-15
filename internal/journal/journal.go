// Package journal manages daily journal entries for mdam.
package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AegirAexx/mdam/internal/document"
	"github.com/AegirAexx/mdam/internal/template"
)

// DateFormat is the filename format for journal entries.
const DateFormat = "2006-01-02"

var journalFilenameRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\.md$`)

// TodayPath returns the full path for today's journal entry.
func TodayPath(journalDir string) string {
	return EntryPath(journalDir, time.Now())
}

// EntryPath returns the full path for the journal entry on the given date.
func EntryPath(journalDir string, date time.Time) string {
	return filepath.Join(journalDir, date.Format(DateFormat)+".md")
}

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(s string) (time.Time, error) {
	t, err := time.Parse(DateFormat, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD format", s)
	}
	return t, nil
}

// Exists reports whether a journal entry exists for the given date.
func Exists(journalDir string, date time.Time) bool {
	_, err := os.Stat(EntryPath(journalDir, date))
	return err == nil
}

// Create creates a journal entry for the given date using the journal template.
// If the entry already exists, it is returned without modification.
// tmplDir is scanned for a "journal" template; built-in is used as fallback.
func Create(journalDir, tmplDir string, date time.Time) (string, error) {
	path := EntryPath(journalDir, date)

	// No-op if entry already exists.
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	content, err := renderJournalTemplate(tmplDir, date)
	if err != nil {
		return "", fmt.Errorf("rendering journal template: %w", err)
	}

	if err := os.MkdirAll(journalDir, 0o755); err != nil {
		return "", fmt.Errorf("creating journal directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing journal entry: %w", err)
	}

	return path, nil
}

// renderJournalTemplate finds and renders the journal template. Falls back to
// the built-in template if the templates directory has no journal template.
func renderJournalTemplate(tmplDir string, date time.Time) (string, error) {
	vars := map[string]string{
		"date_short": date.Format(DateFormat),
		"date":       date.UTC().Format(time.RFC3339),
	}

	tmpl, err := template.Find(tmplDir, "journal")
	if err != nil {
		// Fall back to built-in.
		builtins := template.BuiltinTemplates()
		content, ok := builtins["journal"]
		if !ok {
			return "", fmt.Errorf("no journal template found")
		}
		tmpl = template.Template{Name: "journal", Content: content}
	}

	return template.Render(tmpl, vars)
}

// List returns all journal entry paths in journalDir, sorted newest-first.
func List(journalDir string) ([]string, error) {
	entries, err := os.ReadDir(journalDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("reading journal directory: %w", err)
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if journalFilenameRe.MatchString(entry.Name()) {
			paths = append(paths, filepath.Join(journalDir, entry.Name()))
		}
	}

	// Sort newest-first (lexicographic descending works for YYYY-MM-DD).
	sort.Slice(paths, func(i, j int) bool {
		return paths[i] > paths[j]
	})
	return paths, nil
}

// ListByMonth returns journal entry paths for a given year-month (e.g. "2026-03").
func ListByMonth(journalDir, yearMonth string) ([]string, error) {
	all, err := List(journalDir)
	if err != nil {
		return nil, err
	}
	var filtered []string
	for _, p := range all {
		base := strings.TrimSuffix(filepath.Base(p), ".md")
		if strings.HasPrefix(base, yearMonth) {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

// PastEntries returns journal entry paths for dates strictly before today,
// sorted newest-first.
func PastEntries(journalDir string) ([]string, error) {
	all, err := List(journalDir)
	if err != nil {
		return nil, err
	}
	today := time.Now().Format(DateFormat)
	var past []string
	for _, p := range all {
		base := strings.TrimSuffix(filepath.Base(p), ".md")
		if base < today {
			past = append(past, p)
		}
	}
	return past, nil
}

// ScaffoldFrontmatter creates a valid Frontmatter for a journal entry on date.
func ScaffoldFrontmatter(date time.Time) document.Frontmatter {
	now := time.Now().UTC()
	return document.Frontmatter{
		Title:    date.Format(DateFormat),
		Tags:     []string{},
		Created:  now,
		Modified: now,
		Type:     "journal",
	}
}
