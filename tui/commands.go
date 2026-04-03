package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/AegirAexx/mdam/internal/config"
	"github.com/AegirAexx/mdam/internal/export"
	"github.com/AegirAexx/mdam/internal/git"
	"github.com/AegirAexx/mdam/internal/journal"
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

// cmdJournalCreate creates today's journal entry and opens it in the editor.
// If the entry already exists, it is opened without modification.
func cmdJournalCreate(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		path, err := journal.Create(cfg.JournalDir(), cfg.TemplatesDir(), time.Now())
		if err != nil {
			return fileCreatedMsg{err: fmt.Errorf("creating journal: %w", err)}
		}
		return scratchReadyMsg{path: path}
	}
}

// resolveEditor returns the editor binary to use, preferring cfgEditor then $EDITOR.
// Returns an empty string if neither is configured.
func resolveEditor(cfgEditor string) string {
	if cfgEditor != "" {
		return cfgEditor
	}
	return os.Getenv("EDITOR")
}

// cmdOpenEditor suspends the TUI and opens path in the given editor.
// Sends editorReturnMsg when the editor exits.
func cmdOpenEditor(path, editor string) tea.Cmd {
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorReturnMsg{err: err}
	})
}

// cmdEnsureAndOpenScratch creates the scratch pad document if it does not exist,
// then sends scratchReadyMsg with the path so the caller can open it in the editor.
func cmdEnsureAndOpenScratch(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		path := cfg.ScratchPath()
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return editorReturnMsg{err: fmt.Errorf("creating scratch dir: %w", err)}
			}
			now := time.Now().Format("2006-01-02")
			content := fmt.Sprintf(
				"---\ntype: scratch\ntitle: Scratch Pad\ntags: []\ncreated: %s\nmodified: %s\n---\n",
				now, now,
			)
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return editorReturnMsg{err: fmt.Errorf("creating scratch: %w", err)}
			}
		}
		return scratchReadyMsg{path: path}
	}
}

// cmdLoadPreview reads a file and renders it with glamour, returning previewReadyMsg.
func cmdLoadPreview(path, glamourStyle string, width int) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return previewReadyMsg{content: fmt.Sprintf("  (error reading file: %v)", err)}
		}
		rendered, err := glamour.Render(string(content), glamourStyle)
		if err != nil {
			// Fallback: show raw content.
			return previewReadyMsg{content: string(content)}
		}
		return previewReadyMsg{content: rendered}
	}
}

// cmdLoadPins reads the pinned document paths from pinsPath.
func cmdLoadPins(pinsPath string) tea.Cmd {
	return func() tea.Msg {
		pins, err := loadPins(pinsPath)
		return pinsLoadedMsg{pins: pins, err: err}
	}
}

// cmdSavePins writes the pinned paths to pinsPath asynchronously.
func cmdSavePins(pinsPath string, pins map[string]bool) tea.Cmd {
	return func() tea.Msg {
		_ = savePins(pinsPath, pins) // errors silently dropped — pins are best-effort
		return nil
	}
}

// stripFrontmatter removes the leading YAML frontmatter block (---…---) from
// content. If no frontmatter is present, the original content is returned.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return content
	}
	rest := content[4:] // skip the opening "---\n"
	idx := strings.Index(rest, "\n---\n")
	if idx == -1 {
		return content
	}
	return strings.TrimSpace(rest[idx+5:]) // skip "\n---\n"
}

// cmdLoadRead reads path, strips frontmatter, renders with glamour, and sends readReadyMsg.
func cmdLoadRead(path, glamourStyle string, width int) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return readReadyMsg{content: fmt.Sprintf("  (error reading file: %v)", err)}
		}
		stripped := stripFrontmatter(string(content))
		rendered, err := glamour.Render(stripped, glamourStyle)
		if err != nil {
			return readReadyMsg{content: stripped}
		}
		return readReadyMsg{content: rendered}
	}
}

// cmdBuildTagIndex builds the tag index from docs and sends tagIndexMsg.
func cmdBuildTagIndex(docs []search.Result) tea.Cmd {
	return func() tea.Msg {
		return tagIndexMsg{entries: buildTagIndex(docs)}
	}
}

// cmdTick returns a one-shot 100 ms timer for the loading spinner animation.
func cmdTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
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
