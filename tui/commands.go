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
)

// cmdLoadDocs scans baseDir for all managed markdown documents.
func cmdLoadDocs(baseDir string) tea.Cmd {
	return func() tea.Msg {
		docs, skipped, err := search.ListAll(baseDir)
		return docsLoadedMsg{docs: docs, skipCount: skipped, err: err}
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

// cmdCreateDoc renders a template with vars and writes a new document to disk.
func cmdCreateDoc(t tmpl.Template, vars map[string]string, cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		content, err := tmpl.Render(t, vars)
		if err != nil {
			return fileCreatedMsg{err: fmt.Errorf("rendering template: %w", err)}
		}

		// Determine destination directory by document type.
		// Type is a literal in the template frontmatter, not a {{variable}},
		// so we extract it from the template content rather than vars.
		docType := tmpl.TemplateType(t.Content)
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

// cmdAutoCreateJournal creates today's journal entry on startup if auto_create is enabled.
// Returns nil cmd if disabled. Triggers a doc re-scan if a new entry was created.
func cmdAutoCreateJournal(cfg config.Config) tea.Cmd {
	if !cfg.Journal.AutoCreate {
		return nil
	}
	return func() tea.Msg {
		existed := journal.Exists(cfg.JournalDir(), time.Now())
		_, err := journal.Create(cfg.JournalDir(), cfg.TemplatesDir(), time.Now())
		if err != nil {
			return journalAutoCreateMsg{err: err}
		}
		return journalAutoCreateMsg{created: !existed}
	}
}

// cmdEnsureAndOpenTodo creates the todo file if it does not exist,
// then sends todoReadyMsg with the path so the caller can open it in the editor.
func cmdEnsureAndOpenTodo(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		path := cfg.TodoPath()
		if _, err := os.Stat(path); os.IsNotExist(err) {
			now := time.Now().Format("2006-01-02")
			content := fmt.Sprintf(
				"---\ntype: todo\ntitle: TODO\ntags: []\ncreated: %s\nmodified: %s\n---\n",
				now, now,
			)
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return editorReturnMsg{err: fmt.Errorf("creating todo: %w", err)}
			}
		}
		return todoReadyMsg{path: path}
	}
}

// cmdLoadDashTodo reads the todo file and renders it with glamour for the dashboard.
func cmdLoadDashTodo(path, glamourStyle string, width int) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return dashTodoReadyMsg{err: err}
		}
		prepared := prepareDocPreview(string(content))
		rendered, err := renderGlamour(prepared, glamourStyle, width)
		if err != nil {
			return dashTodoReadyMsg{content: prepared}
		}
		return dashTodoReadyMsg{content: strings.TrimLeft(rendered, "\n")}
	}
}

// cmdLoadPreview reads a file and renders it with glamour, returning previewReadyMsg.
func cmdLoadPreview(path, glamourStyle string, width int) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return previewReadyMsg{content: fmt.Sprintf("  (error reading file: %v)", err)}
		}
		rendered, err := renderGlamour(string(content), glamourStyle, width)
		if err != nil {
			// Fallback: show raw content.
			return previewReadyMsg{content: string(content)}
		}
		return previewReadyMsg{content: rendered}
	}
}

// cmdLoadPins reads the pinned document paths from pinsPath.
func cmdLoadPins(pinsPath, baseDir string) tea.Cmd {
	return func() tea.Msg {
		pins, err := loadPins(pinsPath, baseDir)
		return pinsLoadedMsg{pins: pins, err: err}
	}
}

// cmdSavePins writes the pinned paths to pinsPath asynchronously.
func cmdSavePins(pinsPath, baseDir string, pins []string) tea.Cmd {
	return func() tea.Msg {
		_ = savePins(pinsPath, baseDir, pins) // errors silently dropped — pins are best-effort
		return nil
	}
}

// renderGlamour renders markdown content using glamour with the given style and
// word wrap width. Falls back to glamour.Render if the renderer fails to create.
func renderGlamour(content, style string, width int) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return glamour.Render(content, style)
	}
	return r.Render(content)
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

// prepareDocPreview strips YAML frontmatter from content and prepends the raw
// "tags:" line from the frontmatter block (if tags are present). The result is
// passed to glamour for rendering.
func prepareDocPreview(content string) string {
	tagsLine := ""
	if strings.HasPrefix(content, "---\n") {
		rest := content[4:]
		end := strings.Index(rest, "\n---\n")
		if end != -1 {
			for _, line := range strings.Split(rest[:end], "\n") {
				if strings.HasPrefix(line, "tags:") {
					tagsLine = line
					break
				}
			}
		}
	}
	body := stripFrontmatter(content)
	if tagsLine != "" {
		return tagsLine + "\n\n" + body
	}
	return body
}

// cmdLoadPreviewDoc reads a file, strips its frontmatter, prepends the tags
// line, renders with glamour, and trims any leading blank lines from the output.
func cmdLoadPreviewDoc(path, glamourStyle string, width int) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return previewReadyMsg{content: fmt.Sprintf("  (error reading file: %v)", err)}
		}
		prepared := prepareDocPreview(string(content))
		rendered, err := renderGlamour(prepared, glamourStyle, width)
		if err != nil {
			return previewReadyMsg{content: prepared}
		}
		return previewReadyMsg{content: strings.TrimLeft(rendered, "\n")}
	}
}

// cmdLoadRead reads path, strips frontmatter, renders with glamour, and sends readReadyMsg.
func cmdLoadRead(path, glamourStyle string, width int) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return readReadyMsg{content: fmt.Sprintf("  (error reading file: %v)", err)}
		}
		stripped := stripFrontmatter(string(content))
		rendered, err := renderGlamour(stripped, glamourStyle, width)
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
