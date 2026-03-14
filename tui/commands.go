package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/AegirAexx/mdam/internal/config"
	"github.com/AegirAexx/mdam/internal/export"
	"github.com/AegirAexx/mdam/internal/git"
	"github.com/AegirAexx/mdam/internal/search"
	tmpl "github.com/AegirAexx/mdam/internal/template"
	"github.com/AegirAexx/mdam/internal/todo"
)

// cmdLoadDocs scans baseDir for all managed markdown documents.
func cmdLoadDocs(baseDir string) tea.Cmd {
	return func() tea.Msg {
		docs, err := search.ListAll(baseDir)
		return docsLoadedMsg{docs: docs, err: err}
	}
}

// cmdLoadTodos reads tasks from the global TODO file.
func cmdLoadTodos(path string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := todo.ReadTasks(path)
		return todosLoadedMsg{tasks: tasks, err: err}
	}
}

// cmdLoadGitStatus runs git status for the managed tree root.
func cmdLoadGitStatus(dir string) tea.Cmd {
	return func() tea.Msg {
		status, err := git.Status(dir)
		return gitStatusMsg{status: status, err: err}
	}
}

// cmdSearch runs a fuzzy search across the document tree.
func cmdSearch(baseDir, query string, filters search.Filters) tea.Cmd {
	return func() tea.Msg {
		results, err := search.Search(baseDir, query, filters)
		return searchDoneMsg{results: results, query: query, err: err}
	}
}

// cmdExport strips frontmatter from srcPath and writes to destDir.
func cmdExport(srcPath, destDir string) tea.Cmd {
	return func() tea.Msg {
		dest, err := export.ToFile(srcPath, destDir)
		return exportDoneMsg{dest: dest, err: err}
	}
}

// cmdSweep runs the TODO sweep against yesterday's journal entry.
func cmdSweep(journalDir, todoPath string) tea.Cmd {
	return func() tea.Msg {
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		journalPath := filepath.Join(journalDir, yesterday+".md")
		err := todo.Sweep(journalPath, todoPath)
		return sweepDoneMsg{err: err}
	}
}

// cmdArchive moves old completed tasks to the archive file.
func cmdArchive(todoPath, archivePath string, days int) tea.Cmd {
	return func() tea.Msg {
		dur := time.Duration(days) * 24 * time.Hour
		err := todo.Archive(todoPath, archivePath, dur)
		return sweepDoneMsg{err: err}
	}
}

// cmdCreateDoc renders a template with vars and writes a new document to disk.
func cmdCreateDoc(t tmpl.Template, vars map[string]string, cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		content, err := tmpl.Render(t, vars)
		if err != nil {
			return fileCreatedMsg{err: fmt.Errorf("rendering template: %w", err)}
		}

		// Determine destination directory by document type.
		docType := vars["type"]
		var destDir string
		switch docType {
		case "journal":
			destDir = cfg.JournalDir()
		case "kb", "howto", "meeting":
			destDir = cfg.KBDir()
		default:
			destDir = cfg.BaseDir
		}

		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return fileCreatedMsg{err: fmt.Errorf("creating directory: %w", err)}
		}

		title := vars["title"]
		if title == "" {
			title = t.Name
		}
		filename := toKebabCase(title) + ".md"
		path := filepath.Join(destDir, filename)

		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fileCreatedMsg{err: fmt.Errorf("writing file: %w", err)}
		}
		return fileCreatedMsg{path: path}
	}
}

// toKebabCase converts a string to kebab-case (lowercase, spaces → hyphens).
func toKebabCase(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		case r == ' ', r == '_':
			b.WriteRune('-')
		}
	}
	result := b.String()
	if result == "" {
		return "untitled"
	}
	return result
}
