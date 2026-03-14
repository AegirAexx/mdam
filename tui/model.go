package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/AegirAexx/mdam/internal/config"
	"github.com/AegirAexx/mdam/internal/git"
	"github.com/AegirAexx/mdam/internal/search"
	tmpl "github.com/AegirAexx/mdam/internal/template"
	"github.com/AegirAexx/mdam/internal/todo"
)

// Model is the root BubbleTea model for the mdam TUI.
// It owns all application state for the current render cycle.
type Model struct {
	cfg config.Config

	// mode is the current input mode.
	mode Mode

	// activePanel is the panel that currently has keyboard focus.
	activePanel PanelID

	// activeView controls which document set is displayed.
	activeView View

	// keys holds the active keybindings.
	keys KeyMap

	// --- Document state ---
	docs      []search.Result // all documents from last scan
	fileCursor int

	// --- TODO state ---
	todos      []todo.Task
	todoCursor int

	// --- Git state ---
	gitStatus  git.RepoStatus
	gitFileMap map[string]git.FileStatus // path → status for O(1) lookup

	// --- Search state ---
	searchResults []search.Result
	searchActive  bool // true = showing search results in file panel

	// --- Template picker state ---
	templates    []tmpl.Template
	pickerCursor int
	pendingTmpl  tmpl.Template
	varNames     []string // unresolved {{var}} names from selected template
	varValues    []string // user-entered values (parallel to varNames)
	varCursor    int      // index of var currently being filled
	varInput     textinput.Model

	// --- Loading state ---
	loading bool
	errorMsg string

	// --- Chord state ---
	lastKey string

	// --- Mode inputs ---
	cmdInput    textinput.Model
	searchInput textinput.Model

	// --- Help overlay ---
	showHelp bool

	// --- Status message ---
	statusMsg string

	// --- Terminal dimensions ---
	width  int
	height int
}

// New creates a new Model using the provided config.
func New(cfg config.Config) Model {
	cmd := textinput.New()
	cmd.Placeholder = "command"
	cmd.CharLimit = 256

	srch := textinput.New()
	srch.Placeholder = "search"
	srch.CharLimit = 256

	varIn := textinput.New()
	varIn.CharLimit = 256

	return Model{
		cfg:         cfg,
		mode:        ModeNormal,
		activePanel: PanelFiles,
		activeView:  ViewAll,
		keys:        DefaultKeyMap(),
		gitFileMap:  make(map[string]git.FileStatus),
		loading:     true,
		cmdInput:    cmd,
		searchInput: srch,
		varInput:    varIn,
		width:       80,
		height:      24,
	}
}

// Init satisfies tea.Model. Kicks off the initial directory scan.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		cmdLoadDocs(m.cfg.BaseDir),
		cmdLoadTodos(m.cfg.TodoPath()),
		cmdLoadGitStatus(m.cfg.BaseDir),
	)
}

// Update is the BubbleTea update function. It dispatches to per-mode handlers.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	// --- Async message handlers ---

	case docsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("scan error: %v", msg.err)
		} else {
			m.docs = msg.docs
			m.errorMsg = ""
			// Load templates while we're at it.
			m.templates, _ = tmpl.Discover(m.cfg.TemplatesDir())
		}
		return m, nil

	case todosLoadedMsg:
		if msg.err == nil {
			m.todos = todo.FilterTasks(msg.tasks, "open", "")
		}
		return m, nil

	case gitStatusMsg:
		if msg.err == nil {
			m.gitStatus = msg.status
			m.buildGitFileMap()
		}
		return m, nil

	case searchDoneMsg:
		if msg.err == nil {
			m.searchResults = msg.results
			m.searchActive = true
			if len(msg.results) == 0 {
				m.statusMsg = fmt.Sprintf("no results for %q", msg.query)
			} else {
				m.statusMsg = fmt.Sprintf("%d results for %q", len(msg.results), msg.query)
			}
		} else {
			m.errorMsg = fmt.Sprintf("search error: %v", msg.err)
		}
		m.fileCursor = 0
		return m, nil

	case exportDoneMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("export failed: %v", msg.err)
		} else {
			m.statusMsg = "exported → " + msg.dest
		}
		return m, nil

	case sweepDoneMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("sweep error: %v", msg.err)
		} else {
			m.statusMsg = "sweep done"
		}
		return m, cmdLoadTodos(m.cfg.TodoPath())

	case fileCreatedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("create failed: %v", msg.err)
		} else {
			m.statusMsg = "created " + filepath.Base(msg.path)
		}
		m.mode = ModeNormal
		return m, cmdLoadDocs(m.cfg.BaseDir)

	case editorReturnMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("editor error: %v", msg.err)
		} else {
			m.statusMsg = ""
		}
		m.loading = true
		return m, tea.Batch(
			cmdLoadDocs(m.cfg.BaseDir),
			cmdLoadGitStatus(m.cfg.BaseDir),
			cmdLoadTodos(m.cfg.TodoPath()),
		)

	case scratchReadyMsg:
		editor := resolveEditor(m.cfg.Editor)
		if editor == "" {
			m.statusMsg = "no editor configured ($EDITOR not set)"
			return m, nil
		}
		return m, cmdOpenEditor(msg.path, editor)

	// --- Key events ---

	case tea.KeyMsg:
		switch m.mode {
		case ModeNormal:
			return m.updateNormal(msg)
		case ModeCommand:
			return m.updateCommand(msg)
		case ModeSearch:
			return m.updateSearch(msg)
		case ModeTemplatePicker:
			return m.updateTemplatePicker(msg)
		case ModeTemplateVars:
			return m.updateTemplateVars(msg)
		}
	}
	return m, nil
}

// updateNormal handles key events in Normal mode.
func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	// Handle gg chord: two consecutive 'g' presses → jump to top.
	if k == "g" {
		if m.lastKey == "g" {
			m.lastKey = ""
			return m.jumpTop(), nil
		}
		m.lastKey = "g"
		return m, nil
	}
	m.lastKey = k

	switch {
	// Quit
	case k == "q", k == "ctrl+c":
		return m, tea.Quit

	// Navigation
	case k == "j", k == "down":
		m = m.moveCursorDown()
	case k == "k", k == "up":
		m = m.moveCursorUp()
	case k == "G", k == "end":
		m = m.jumpBottom()
	case k == "home":
		m = m.jumpTop()

	// Panel switching
	case k == "tab":
		m.activePanel = m.activePanel.next()
		m.statusMsg = ""
	case k == "shift+tab":
		m.activePanel = m.activePanel.prev()
		m.statusMsg = ""

	// Panel navigation (h/l move focus between panels)
	case k == "l", k == "right":
		m.activePanel = m.activePanel.next()
	case k == "h", k == "left":
		m.activePanel = m.activePanel.prev()

	// Mode transitions
	case k == "/":
		m.mode = ModeSearch
		m.searchInput.SetValue("")
		m.statusMsg = ""
		return m, m.searchInput.Focus()

	case k == ":":
		m.mode = ModeCommand
		m.cmdInput.SetValue("")
		m.statusMsg = ""
		return m, m.cmdInput.Focus()

	// Help overlay
	case k == "?":
		m.showHelp = !m.showHelp

	// Number keys — view switching
	case k == "1":
		m.activeView = ViewAll
		m.searchActive = false
		m.fileCursor = 0
		m.statusMsg = ""
	case k == "2":
		m.activeView = ViewJournal
		m.searchActive = false
		m.fileCursor = 0
		m.statusMsg = ""
	case k == "3":
		m.activeView = ViewKB
		m.searchActive = false
		m.fileCursor = 0
		m.statusMsg = ""
	case k == "4":
		m.activeView = ViewTodo
		m.activePanel = PanelTodo
		m.statusMsg = ""
	case k == "5":
		m.activeView = ViewRecent
		m.searchActive = false
		m.fileCursor = 0
		m.statusMsg = ""

	// Open in $EDITOR
	case k == "enter":
		if selected := m.selectedDoc(); selected != "" {
			editor := resolveEditor(m.cfg.Editor)
			if editor == "" {
				m.statusMsg = "no editor configured ($EDITOR not set)"
				return m, nil
			}
			return m, cmdOpenEditor(selected, editor)
		}
		m.statusMsg = "no document selected"

	// Rescan
	case k == "R":
		m.loading = true
		m.statusMsg = ""
		return m, tea.Batch(cmdLoadDocs(m.cfg.BaseDir), cmdLoadGitStatus(m.cfg.BaseDir))

	// New document
	case k == "n":
		if len(m.templates) == 0 {
			// Try to load built-ins if templates dir has none.
			builtins := tmpl.BuiltinTemplates()
			for name, content := range builtins {
				m.templates = append(m.templates, tmpl.Template{Name: name, Content: content})
			}
		}
		m.mode = ModeTemplatePicker
		m.pickerCursor = 0
		m.statusMsg = ""

	// Scratch pad
	case k == "s":
		return m, cmdEnsureAndOpenScratch(m.cfg)

	// Export
	case k == "e":
		if selected := m.selectedDoc(); selected != "" {
			return m, cmdExport(selected, m.cfg.ExportDir)
		}
		m.statusMsg = "no document selected"

	// Lazygit handoff
	case k == "ctrl+g":
		if m.cfg.Git.Lazygit {
			return m, cmdOpenLazygit(m.cfg.BaseDir)
		}
		m.statusMsg = "lazygit disabled (git.lazygit = false)"

	// Delete — Phase 3 stub (destructive; full impl deferred)
	case k == "d":
		m.statusMsg = "delete — not yet implemented"
	}

	return m, nil
}

// updateCommand handles key events in Command mode.
func (m Model) updateCommand(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.cmdInput.Blur()
		m.statusMsg = ""
		return m, nil

	case "enter":
		cmd := strings.TrimSpace(m.cmdInput.Value())
		m.mode = ModeNormal
		m.cmdInput.Blur()
		return m.executeCommand(cmd)

	case "ctrl+c":
		return m, tea.Quit
	}

	var teaCmd tea.Cmd
	m.cmdInput, teaCmd = m.cmdInput.Update(msg)
	return m, teaCmd
}

// updateSearch handles key events in Search mode.
func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.searchInput.Blur()
		m.searchActive = false
		m.statusMsg = ""
		return m, nil

	case "enter":
		query := strings.TrimSpace(m.searchInput.Value())
		m.mode = ModeNormal
		m.searchInput.Blur()
		if query == "" {
			m.searchActive = false
			m.statusMsg = ""
			return m, nil
		}
		return m, cmdSearch(m.cfg.BaseDir, query, search.Filters{})

	case "ctrl+c":
		return m, tea.Quit
	}

	var teaCmd tea.Cmd
	m.searchInput, teaCmd = m.searchInput.Update(msg)
	return m, teaCmd
}

// updateTemplatePicker handles key events in the template picker mode.
func (m Model) updateTemplatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		return m, nil

	case "j", "down":
		if m.pickerCursor < len(m.templates)-1 {
			m.pickerCursor++
		}

	case "k", "up":
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}

	case "enter":
		if len(m.templates) == 0 {
			m.mode = ModeNormal
			return m, nil
		}
		m.pendingTmpl = m.templates[m.pickerCursor]
		// Find unresolved variables (excluding builtins already handled by Render).
		rawVars := tmpl.UnresolvedVars(m.pendingTmpl.Content)
		m.varNames = nil
		seen := map[string]bool{}
		for _, v := range rawVars {
			inner := strings.TrimPrefix(strings.TrimSuffix(v, "}}"), "{{")
			if !seen[inner] {
				seen[inner] = true
				m.varNames = append(m.varNames, inner)
			}
		}
		if len(m.varNames) == 0 {
			// No variables to fill; create immediately.
			return m, cmdCreateDoc(m.pendingTmpl, map[string]string{}, m.cfg)
		}
		m.varValues = make([]string, len(m.varNames))
		m.varCursor = 0
		m.varInput.SetValue("")
		m.varInput.Placeholder = m.varNames[0]
		m.mode = ModeTemplateVars
		return m, m.varInput.Focus()

	case "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

// updateTemplateVars handles key events while collecting template variable values.
func (m Model) updateTemplateVars(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeTemplatePicker
		m.varInput.Blur()
		return m, nil

	case "enter":
		m.varValues[m.varCursor] = m.varInput.Value()
		m.varCursor++
		if m.varCursor >= len(m.varNames) {
			// All vars collected; create document.
			vars := make(map[string]string, len(m.varNames))
			for i, name := range m.varNames {
				vars[name] = m.varValues[i]
			}
			m.varInput.Blur()
			return m, cmdCreateDoc(m.pendingTmpl, vars, m.cfg)
		}
		// Advance to next variable.
		m.varInput.SetValue("")
		m.varInput.Placeholder = m.varNames[m.varCursor]
		return m, nil

	case "ctrl+c":
		return m, tea.Quit
	}

	var teaCmd tea.Cmd
	m.varInput, teaCmd = m.varInput.Update(msg)
	return m, teaCmd
}

// executeCommand interprets a colon-command and returns the updated model and cmd.
func (m Model) executeCommand(cmd string) (tea.Model, tea.Cmd) {
	switch {
	case cmd == "q", cmd == "quit":
		return m, tea.Quit
	case cmd == "todo sweep":
		return m, cmdSweep(m.cfg.JournalDir(), m.cfg.TodoPath())
	case cmd == "todo archive":
		return m, cmdArchive(m.cfg.TodoPath(), m.cfg.ArchivePath(), m.cfg.Todo.ArchiveAfterDays)
	case cmd == "":
		return m, nil
	default:
		m.statusMsg = fmt.Sprintf(":%s — unknown command", cmd)
		return m, nil
	}
}

// buildGitFileMap rebuilds the gitFileMap from gitStatus.Files using absolute paths.
func (m *Model) buildGitFileMap() {
	m.gitFileMap = make(map[string]git.FileStatus, len(m.gitStatus.Files))
	for _, f := range m.gitStatus.Files {
		absPath := filepath.Join(m.cfg.BaseDir, f.Path)
		m.gitFileMap[absPath] = f
	}
}

// selectedDoc returns the absolute path of the document under the cursor,
// considering the active view and search state.
func (m Model) selectedDoc() string {
	docs := m.visibleDocs()
	if m.fileCursor >= 0 && m.fileCursor < len(docs) {
		return docs[m.fileCursor].Path
	}
	return ""
}

// visibleDocs returns the slice of documents appropriate for the current view.
func (m Model) visibleDocs() []search.Result {
	if m.searchActive {
		return m.searchResults
	}
	switch m.activeView {
	case ViewJournal:
		return filterByType(m.docs, "journal")
	case ViewKB:
		return filterByType(m.docs, "kb")
	case ViewRecent:
		return recentDocs(m.docs, 20)
	default:
		return m.docs
	}
}

// filterByType returns results matching the given document type.
func filterByType(docs []search.Result, docType string) []search.Result {
	var out []search.Result
	for _, d := range docs {
		if strings.EqualFold(d.Frontmatter.Type, docType) {
			out = append(out, d)
		}
	}
	return out
}

// recentDocs returns up to n docs sorted by Modified descending.
func recentDocs(docs []search.Result, n int) []search.Result {
	sorted := make([]search.Result, len(docs))
	copy(sorted, docs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Frontmatter.Modified.After(sorted[j].Frontmatter.Modified)
	})
	if len(sorted) > n {
		return sorted[:n]
	}
	return sorted
}

// --- Cursor movement helpers ---

func (m Model) moveCursorDown() Model {
	switch m.activePanel {
	case PanelFiles:
		docs := m.visibleDocs()
		if m.fileCursor < len(docs)-1 {
			m.fileCursor++
		}
	case PanelTodo:
		if m.todoCursor < len(m.todos)-1 {
			m.todoCursor++
		}
	}
	return m
}

func (m Model) moveCursorUp() Model {
	switch m.activePanel {
	case PanelFiles:
		if m.fileCursor > 0 {
			m.fileCursor--
		}
	case PanelTodo:
		if m.todoCursor > 0 {
			m.todoCursor--
		}
	}
	return m
}

func (m Model) jumpTop() Model {
	switch m.activePanel {
	case PanelFiles:
		m.fileCursor = 0
	case PanelTodo:
		m.todoCursor = 0
	}
	return m
}

func (m Model) jumpBottom() Model {
	switch m.activePanel {
	case PanelFiles:
		docs := m.visibleDocs()
		if len(docs) > 0 {
			m.fileCursor = len(docs) - 1
		}
	case PanelTodo:
		if len(m.todos) > 0 {
			m.todoCursor = len(m.todos) - 1
		}
	}
	return m
}
