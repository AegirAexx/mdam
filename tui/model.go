package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/AegirAexx/mdam/internal/config"
	"github.com/AegirAexx/mdam/internal/git"
	"github.com/AegirAexx/mdam/internal/search"
	tmpl "github.com/AegirAexx/mdam/internal/template"
)

// spinnerFrames is the loading animation sequence.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

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

	// theme holds all lipgloss styles for the active color palette.
	theme Theme

	// icons holds the glyph set (Nerd Font or ASCII fallback).
	icons Icons

	// --- Document state ---
	docs      []search.Result // all documents from last scan
	fileCursor int

	// --- Dashboard TODO preview ---
	dashTodoRendered string

	// --- Git state ---
	gitStatus  git.RepoStatus
	gitFileMap map[string]git.FileStatus // path → status for O(1) lookup

	// --- Search state ---
	searchResults []search.Result
	searchActive  bool // true = showing search results in file panel

	// --- Template picker state ---
	templates      []tmpl.Template
	pickerTemplates []tmpl.Template // filtered subset shown in the picker overlay
	pickerCursor   int
	pendingTmpl    tmpl.Template
	varNames     []string // unresolved {{var}} names from selected template
	varValues    []string // user-entered values (parallel to varNames)
	varCursor    int      // index of var currently being filled
	varInput     textinput.Model

	// --- Preview viewport ---
	preview viewport.Model

	// --- Pinned documents ---
	pinnedOrder []string        // ordered list of pinned paths (oldest first)
	pinnedPaths map[string]bool // set of pinned absolute paths (derived from pinnedOrder)

	// --- Delete confirmation ---
	deleteConfirmPath  string
	deleteConfirmTitle string // document title shown in delete confirm prompt (§7.1)

	// --- Journal tree ---
	journalExpanded map[string]bool // month key ("2026-04") → expanded
	journalCursor   int

	// --- KB tree ---
	kbExpanded map[string]bool // subtype key → expanded
	kbCursor   int

	// --- Read mode ---
	readViewport    viewport.Model
	readReturnView  View
	readReturnPanel PanelID
	readDocTitle    string // document title shown in read mode header (§3.5)

	// --- Dashboard ---
	dashCursor int  // index into flat dashItem list
	dashRight  bool // true = focus on right (TODO) column

	// --- Tag browser ---
	tagEntries   []tagEntry
	tagCursor    int
	tagDocCursor int // cursor within the documents panel on the right

	// --- Spinner ---
	spinnerFrame int

	// --- Loading state ---
	loading  bool
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

	var icons Icons
	if cfg.NerdFonts {
		icons = DefaultIcons()
	} else {
		icons = PlainIcons()
	}

	return Model{
		cfg:             cfg,
		mode:            ModeNormal,
		activePanel:     PanelFiles,
		activeView:      ViewDashboard,
		keys:            DefaultKeyMap(),
		theme:           NewTheme(cfg.Theme),
		icons:           icons,
		gitFileMap:      make(map[string]git.FileStatus),
		pinnedPaths:     make(map[string]bool),
		journalExpanded: make(map[string]bool),
		kbExpanded:      make(map[string]bool),
		loading:         true,
		cmdInput:        cmd,
		searchInput:     srch,
		varInput:        varIn,
		width:           80,
		height:          24,
		preview:         viewport.New(53, 20), // ~67% of 80 width, 20 rows
	}
}

// Init satisfies tea.Model. Kicks off the initial directory scan.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		cmdLoadDocs(m.cfg.BaseDir),
		cmdLoadDashTodo(m.cfg.TodoPath(), m.theme.GlamourStyle, m.width),
		cmdLoadGitStatus(m.cfg.BaseDir),
		cmdLoadPins(m.cfg.PinsPath()),
		cmdAutoCreateJournal(m.cfg),
		cmdTick(),
	)
}

// Update is the BubbleTea update function. It dispatches to per-mode handlers.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Resize preview viewport to match new terminal dimensions.
		leftWidth := m.width / 3
		rightWidth := m.width - leftWidth - 1
		previewHeight := m.height - 3 // tab bar + status bar + separator
		if previewHeight < 1 {
			previewHeight = 1
		}
		m.preview.Width = rightWidth
		m.preview.Height = previewHeight
		return m, nil

	// --- Spinner ---

	case tickMsg:
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		if m.loading {
			return m, cmdTick()
		}
		return m, nil

	// --- Async message handlers ---

	case docsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("scan error: %v", msg.err)
			return m, nil
		}
		m.docs = msg.docs
		m.errorMsg = ""
		if msg.skipCount > 0 {
			m.statusMsg = fmt.Sprintf("%d file(s) skipped — invalid frontmatter", msg.skipCount)
		}
		// Load templates while we're at it.
		m.templates, _ = tmpl.Discover(m.cfg.TemplatesDir())
		// Re-initialise the journal view so cursor repositions to the newest entry.
		if m.activeView == ViewJournal {
			m = initJournalView(m)
		}
		// Rebuild tag index so ViewTags is always current regardless of how it is reached.
		return m, cmdBuildTagIndex(msg.docs)

	case dashTodoReadyMsg:
		if msg.err == nil {
			m.dashTodoRendered = msg.content
		}
		return m, nil

	case todoReadyMsg:
		editor := resolveEditor(m.cfg.Editor)
		if editor == "" {
			m.statusMsg = "no editor configured ($EDITOR not set)"
			return m, nil
		}
		return m, cmdOpenEditor(msg.path, editor)

	case journalAutoCreateMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("journal auto-create: %v", msg.err)
		} else if msg.created {
			m.statusMsg = "created today's journal"
			return m, cmdLoadDocs(m.cfg.BaseDir)
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
		return m, cmdLoadDashTodo(m.cfg.TodoPath(), m.theme.GlamourStyle, m.width)

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
			cmdLoadDashTodo(m.cfg.TodoPath(), m.theme.GlamourStyle, m.width),
			cmdTick(),
		)

	case scratchReadyMsg:
		editor := resolveEditor(m.cfg.Editor)
		if editor == "" {
			m.statusMsg = "no editor configured ($EDITOR not set)"
			return m, nil
		}
		return m, cmdOpenEditor(msg.path, editor)

	case previewReadyMsg:
		m.preview.SetContent(msg.content)
		return m, nil

	case pinsLoadedMsg:
		if msg.err == nil {
			m.pinnedOrder = msg.pins
			m.pinnedPaths = pinsToMap(msg.pins)
		}
		return m, nil

	case tagIndexMsg:
		m.tagEntries = msg.entries
		return m, nil

	case readReadyMsg:
		m.readViewport.SetContent(msg.content)
		return m, nil

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
		case ModeRead:
			return m.updateRead(msg)
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

	// Pane switching (Tab cycles between the 4 named panes)
	case k == "tab":
		m.activeView = cycleView(m.activeView, 1)
		m.activePanel = PanelFiles
		m.statusMsg = ""
	case k == "shift+tab":
		m.activeView = cycleView(m.activeView, -1)
		m.activePanel = PanelFiles
		m.statusMsg = ""

	// Journal tree: l/h expand-collapse folders only (focus never leaves left panel).
	case (k == "l" || k == "right") && m.activeView == ViewJournal && m.activePanel == PanelFiles:
		rows := buildJournalRows(m.docs, m.journalExpanded)
		if m.journalCursor < len(rows) && rows[m.journalCursor].isFolder {
			// Expand this folder; collapse all others (one open at a time).
			key := rows[m.journalCursor].month
			for k2 := range m.journalExpanded {
				delete(m.journalExpanded, k2)
			}
			m.journalExpanded[key] = true
		}
		// File row: l is a no-op (already inside an open folder).
	case (k == "h" || k == "left") && m.activeView == ViewJournal && m.activePanel == PanelFiles:
		rows := buildJournalRows(m.docs, m.journalExpanded)
		if m.journalCursor < len(rows) && rows[m.journalCursor].isFolder {
			// Collapse this folder.
			delete(m.journalExpanded, rows[m.journalCursor].month)
		} else {
			// File row: collapse parent month folder and jump cursor to it.
			for i := m.journalCursor - 1; i >= 0; i-- {
				if rows[i].isFolder {
					delete(m.journalExpanded, rows[i].month)
					m.journalCursor = i
					break
				}
			}
		}

	// Dashboard: l/h switch between left and right columns
	case (k == "l" || k == "right") && m.activeView == ViewDashboard:
		m.dashRight = true
	case (k == "h" || k == "left") && m.activeView == ViewDashboard:
		m.dashRight = false

	// KB tree: l/h expand-collapse folders only (focus never leaves left panel).
	case (k == "l" || k == "right") && m.activeView == ViewKB && m.activePanel == PanelFiles:
		rows := buildKBRows(m.docs, m.kbExpanded)
		if m.kbCursor < len(rows) && rows[m.kbCursor].isFolder {
			key := rows[m.kbCursor].subtype
			for k2 := range m.kbExpanded {
				delete(m.kbExpanded, k2)
			}
			m.kbExpanded[key] = true
		}
		// File row: l is a no-op (already inside an open folder).
	case (k == "h" || k == "left") && m.activeView == ViewKB && m.activePanel == PanelFiles:
		rows := buildKBRows(m.docs, m.kbExpanded)
		if m.kbCursor < len(rows) && rows[m.kbCursor].isFolder {
			delete(m.kbExpanded, rows[m.kbCursor].subtype)
		} else {
			// File row: collapse parent subtype folder and jump cursor to it.
			for i := m.kbCursor - 1; i >= 0; i-- {
				if rows[i].isFolder {
					delete(m.kbExpanded, rows[i].subtype)
					m.kbCursor = i
					break
				}
			}
		}

	// Tag Browser: l/h move focus between panels without wrapping.
	case (k == "l" || k == "right") && m.activeView == ViewTags:
		m.activePanel = PanelPreview
	case (k == "h" || k == "left") && m.activeView == ViewTags:
		m.activePanel = PanelFiles

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

	// Number keys — pane switching
	case k == "1":
		m.activeView = ViewDashboard
		m.searchActive = false
		m.dashCursor = 0
		m.dashRight = false
		m.activePanel = PanelFiles
		m.statusMsg = ""
	case k == "2":
		m.activeView = ViewJournal
		m.searchActive = false
		m.activePanel = PanelFiles
		m.statusMsg = ""
		m.preview.SetContent("") // clear stale content from previous view
		m = initJournalView(m)
	case k == "3":
		m.activeView = ViewKB
		m.searchActive = false
		m.kbCursor = 0
		m.activePanel = PanelFiles
		m.statusMsg = ""
		m.preview.SetContent("") // clear stale content from previous view
	case k == "4":
		m.activeView = ViewTags
		m.searchActive = false
		m.tagCursor = 0
		m.activePanel = PanelFiles // always reset to left panel on entry (§9 Bug 1)
		m.statusMsg = ""
		return m, cmdBuildTagIndex(m.docs)

	// Open in $EDITOR
	case k == "enter":
		// Dashboard right panel: open todo.md in editor.
		if m.activeView == ViewDashboard && m.dashRight {
			return m, cmdEnsureAndOpenTodo(m.cfg)
		}
		// Dashboard: enter on left column file item opens in editor.
		if m.activeView == ViewDashboard && !m.dashRight {
			items := buildDashItems(m)
			if m.dashCursor < len(items) && !items[m.dashCursor].isHeader {
				path := items[m.dashCursor].doc.Path
				editor := resolveEditor(m.cfg.Editor)
				if editor == "" {
					m.statusMsg = "no editor configured ($EDITOR not set)"
					return m, nil
				}
				return m, cmdOpenEditor(path, editor)
			}
			return m, nil
		}

		// KB tree: enter on folder row toggles expand/collapse.
		if m.activeView == ViewKB && m.activePanel == PanelFiles {
			rows := buildKBRows(m.docs, m.kbExpanded)
			if m.kbCursor < len(rows) {
				row := rows[m.kbCursor]
				if row.isFolder {
					if m.kbExpanded[row.subtype] {
						delete(m.kbExpanded, row.subtype)
					} else {
						for k2 := range m.kbExpanded {
							delete(m.kbExpanded, k2)
						}
						m.kbExpanded[row.subtype] = true
					}
					return m, nil
				}
				editor := resolveEditor(m.cfg.Editor)
				if editor == "" {
					m.statusMsg = "no editor configured ($EDITOR not set)"
					return m, nil
				}
				return m, cmdOpenEditor(row.path, editor)
			}
			return m, nil
		}

		// Journal tree: enter on folder row toggles expand/collapse.
		if m.activeView == ViewJournal && m.activePanel == PanelFiles {
			rows := buildJournalRows(m.docs, m.journalExpanded)
			if m.journalCursor < len(rows) {
				row := rows[m.journalCursor]
				if row.isFolder {
					if m.journalExpanded[row.month] {
						delete(m.journalExpanded, row.month)
					} else {
						for k2 := range m.journalExpanded {
							delete(m.journalExpanded, k2)
						}
						m.journalExpanded[row.month] = true
					}
					return m, nil
				}
				// File row: open in editor.
				editor := resolveEditor(m.cfg.Editor)
				if editor == "" {
					m.statusMsg = "no editor configured ($EDITOR not set)"
					return m, nil
				}
				return m, cmdOpenEditor(row.path, editor)
			}
			m.statusMsg = "no document selected"
			return m, nil
		}

		// Tag browser: Enter on doc panel opens the highlighted tagged doc.
		if m.activeView == ViewTags && m.activePanel == PanelPreview {
			tagged := m.taggedDocs()
			if m.tagDocCursor < len(tagged) {
				editor := resolveEditor(m.cfg.Editor)
				if editor == "" {
					m.statusMsg = "no editor configured ($EDITOR not set)"
					return m, nil
				}
				return m, cmdOpenEditor(tagged[m.tagDocCursor].Path, editor)
			}
			m.statusMsg = "no document selected"
			return m, nil
		}
		if selected := m.selectedDoc(); selected != "" {
			editor := resolveEditor(m.cfg.Editor)
			if editor == "" {
				m.statusMsg = "no editor configured ($EDITOR not set)"
				return m, nil
			}
			return m, cmdOpenEditor(selected, editor)
		}
		m.statusMsg = "no document selected"

	// Open in full-screen read mode
	case k == "o":
		// Tag Browser: o on doc panel opens the highlighted tagged doc in read mode.
		if m.activeView == ViewTags && m.activePanel == PanelPreview {
			tagged := m.taggedDocs()
			if m.tagDocCursor < len(tagged) {
				d := tagged[m.tagDocCursor]
				title := d.Frontmatter.Title
				if title == "" {
					title = filepath.Base(d.Path)
				}
				m.readReturnView = m.activeView
				m.readReturnPanel = m.activePanel
				m.readDocTitle = title
				m.readViewport = viewport.New(m.width, m.height-3)
				m.mode = ModeRead
				return m, cmdLoadRead(d.Path, m.theme.GlamourStyle, m.width)
			}
			m.statusMsg = "no document selected"
			return m, nil
		}
		// Dashboard right panel: read todo.md.
		if m.activeView == ViewDashboard && m.dashRight {
			m.readReturnView = m.activeView
			m.readReturnPanel = m.activePanel
			m.readDocTitle = "TODO"
			m.readViewport = viewport.New(m.width, m.height-3)
			m.mode = ModeRead
			return m, cmdLoadRead(m.cfg.TodoPath(), m.theme.GlamourStyle, m.width)
		}
		// Dashboard: o on left column file item enters read mode.
		if m.activeView == ViewDashboard && !m.dashRight {
			items := buildDashItems(m)
			if m.dashCursor < len(items) && !items[m.dashCursor].isHeader && !items[m.dashCursor].isBlank && !items[m.dashCursor].isPlaceholder {
				it := items[m.dashCursor]
				path := it.doc.Path
				title := it.doc.Frontmatter.Title
				if title == "" {
					title = filepath.Base(path)
				}
				m.readReturnView = m.activeView
				m.readReturnPanel = m.activePanel
				m.readDocTitle = title
				m.readViewport = viewport.New(m.width, m.height-3)
				m.mode = ModeRead
				return m, cmdLoadRead(path, m.theme.GlamourStyle, m.width)
			}
			m.statusMsg = "no document selected"
			return m, nil
		}
		if path := m.selectedDoc(); path != "" {
			m.readReturnView = m.activeView
			m.readReturnPanel = m.activePanel
			m.readDocTitle = m.selectedDocTitle()
			m.readViewport = viewport.New(m.width, m.height-3)
			m.mode = ModeRead
			return m, cmdLoadRead(path, m.theme.GlamourStyle, m.width)
		}
		m.statusMsg = "no document selected"

	// Rescan
	case k == "R":
		m.loading = true
		m.statusMsg = ""
		return m, tea.Batch(cmdLoadDocs(m.cfg.BaseDir), cmdLoadGitStatus(m.cfg.BaseDir), cmdTick())

	// New document
	case k == "n":
		if len(m.templates) == 0 {
			builtins := tmpl.BuiltinTemplates()
			for name, content := range builtins {
				m.templates = append(m.templates, tmpl.Template{Name: name, Content: content})
			}
		}
		// Expose only top-level document types (journal and kb) in the picker.
		m.pickerTemplates = nil
		for _, t := range m.templates {
			if t.Name == "journal" || t.Name == "kb" {
				m.pickerTemplates = append(m.pickerTemplates, t)
			}
		}
		m.mode = ModeTemplatePicker
		m.pickerCursor = 0
		m.statusMsg = ""

	// Scratch pad
	case k == "s":
		return m, cmdEnsureAndOpenScratch(m.cfg)

	// Todo
	case k == "t":
		return m, cmdEnsureAndOpenTodo(m.cfg)

	// Export
	case k == "e":
		if selected := m.selectedDoc(); selected != "" {
			return m, cmdExport(selected, m.cfg.ExportDir)
		}
		m.statusMsg = "no document selected"

	// Pin / unpin
	case k == "p":
		if selected := m.selectedDoc(); selected != "" {
			m.pinnedOrder = togglePin(m.pinnedOrder, selected)
			m.pinnedPaths = pinsToMap(m.pinnedOrder)
			return m, cmdSavePins(m.cfg.PinsPath(), m.pinnedOrder)
		}
		m.statusMsg = "no document selected"

	}

	// Fire preview load when cursor moves in file panel.
	if m.activePanel == PanelFiles {
		if selected := m.selectedDoc(); selected != "" {
			switch m.activeView {
			case ViewJournal, ViewKB:
				return m, cmdLoadPreviewDoc(selected, m.theme.GlamourStyle, m.preview.Width)
			default:
				return m, cmdLoadPreview(selected, m.theme.GlamourStyle, m.preview.Width)
			}
		}
	}

	return m, nil
}

// updateRead handles key events in full-screen read mode (ModeRead).
func (m Model) updateRead(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.mode = ModeNormal
		m.activeView = m.readReturnView
		m.activePanel = m.readReturnPanel
	case "j", "down":
		m.readViewport.LineDown(1)
	case "k", "up":
		m.readViewport.LineUp(1)
	case "d":
		m.readViewport.HalfViewDown()
	case "u":
		m.readViewport.HalfViewUp()
	case "f":
		m.readViewport.ViewDown()
	case "b":
		m.readViewport.ViewUp()
	case "g":
		m.readViewport.GotoTop()
	case "G":
		m.readViewport.GotoBottom()
	case "ctrl+c":
		return m, tea.Quit
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
		if m.pickerCursor < len(m.pickerTemplates)-1 {
			m.pickerCursor++
		}

	case "k", "up":
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}

	case "enter":
		if len(m.pickerTemplates) == 0 {
			m.mode = ModeNormal
			return m, nil
		}
		m.pendingTmpl = m.pickerTemplates[m.pickerCursor]

		// Journal entries bypass the variable flow — delegate to journal.Create.
		if tmpl.TemplateType(m.pendingTmpl.Content) == "journal" {
			m.mode = ModeNormal
			return m, cmdJournalCreate(m.cfg)
		}

		// Render builtins first so only genuinely user-supplied vars remain.
		rendered, _ := tmpl.Render(m.pendingTmpl, map[string]string{})
		rawVars := tmpl.UnresolvedVars(rendered)
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
			vars := make(map[string]string, len(m.varNames))
			for i, name := range m.varNames {
				vars[name] = m.varValues[i]
			}
			m.varInput.Blur()
			return m, cmdCreateDoc(m.pendingTmpl, vars, m.cfg)
		}
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
		m.statusMsg = "todo sweep is not yet available"
		return m, nil
	case cmd == "todo archive":
		m.statusMsg = "todo archive is not yet available"
		return m, nil
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
	if m.activeView == ViewDashboard {
		items := buildDashItems(m)
		if m.dashCursor >= 0 && m.dashCursor < len(items) {
			it := items[m.dashCursor]
			if !it.isHeader && !it.isBlank && !it.isPlaceholder {
				return it.doc.Path
			}
		}
		return ""
	}
	if m.activeView == ViewJournal {
		return m.journalSelectedPath()
	}
	if m.activeView == ViewKB {
		return m.kbSelectedPath()
	}
	docs := m.visibleDocs()
	if m.fileCursor >= 0 && m.fileCursor < len(docs) {
		return docs[m.fileCursor].Path
	}
	return ""
}

// selectedDocTitle returns the frontmatter title of the document under the cursor,
// falling back to the base filename if the title is empty.
func (m Model) selectedDocTitle() string {
	var path string
	var title string
	if m.activeView == ViewDashboard {
		items := buildDashItems(m)
		if m.dashCursor >= 0 && m.dashCursor < len(items) {
			it := items[m.dashCursor]
			if !it.isHeader && !it.isBlank && !it.isPlaceholder {
				path = it.doc.Path
				title = it.label
			}
		}
	} else if m.activeView == ViewJournal {
		path = m.journalSelectedPath()
		rows := buildJournalRows(filterByType(m.docs, "journal"), m.journalExpanded)
		if m.journalCursor >= 0 && m.journalCursor < len(rows) {
			r := rows[m.journalCursor]
			if !r.isFolder {
				for _, d := range m.docs {
					if d.Path == r.path {
						title = d.Frontmatter.Title
						break
					}
				}
			}
		}
	} else if m.activeView == ViewKB {
		path = m.kbSelectedPath()
		rows := buildKBRows(m.docs, m.kbExpanded)
		if m.kbCursor >= 0 && m.kbCursor < len(rows) {
			r := rows[m.kbCursor]
			if !r.isFolder {
				title = r.title
			}
		}
	} else {
		docs := m.visibleDocs()
		if m.fileCursor >= 0 && m.fileCursor < len(docs) {
			path = docs[m.fileCursor].Path
			title = docs[m.fileCursor].Frontmatter.Title
		}
	}
	if title != "" {
		return title
	}
	if path != "" {
		return filepath.Base(path)
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
		return filterKBDocs(m.docs)
	case ViewTags:
		return nil // tag browser uses tagEntries, not docs
	default:
		return m.docs
	}
}

// docCounts returns the count of journal and kb (all kb_* types) documents.
func docCounts(docs []search.Result) (journal, kb int) {
	for _, d := range docs {
		t := strings.ToLower(d.Frontmatter.Type)
		switch {
		case t == "journal":
			journal++
		case strings.HasPrefix(t, "kb"):
			kb++
		}
	}
	return
}

// highlightedRelPath returns the relative path of the currently highlighted document
// (relative to baseDir), or "" if none is selected or the path cannot be relativised.
func highlightedRelPath(m Model) string {
	var absPath string
	switch {
	case m.activeView == ViewDashboard && !m.dashRight:
		items := buildDashItems(m)
		if m.dashCursor < len(items) {
			it := items[m.dashCursor]
			if !it.isHeader && !it.isBlank && !it.isPlaceholder {
				absPath = it.doc.Path
			}
		}
	case m.activeView == ViewJournal:
		absPath = m.journalSelectedPath()
	case m.activeView == ViewKB:
		absPath = m.kbSelectedPath()
	case m.activeView == ViewTags && m.activePanel == PanelPreview:
		tagged := m.taggedDocs()
		if m.tagDocCursor >= 0 && m.tagDocCursor < len(tagged) {
			absPath = tagged[m.tagDocCursor].Path
		}
	default:
		docs := m.visibleDocs()
		if m.fileCursor >= 0 && m.fileCursor < len(docs) {
			absPath = docs[m.fileCursor].Path
		}
	}
	if absPath == "" || m.cfg.BaseDir == "" {
		return ""
	}
	rel, err := filepath.Rel(m.cfg.BaseDir, absPath)
	if err != nil {
		return ""
	}
	return rel
}

// taggedDocs returns the documents that carry the currently selected tag.
func (m Model) taggedDocs() []search.Result {
	if m.tagCursor >= len(m.tagEntries) {
		return nil
	}
	tag := m.tagEntries[m.tagCursor].Name
	var out []search.Result
	for _, d := range m.docs {
		for _, t := range d.Frontmatter.Tags {
			if t == tag {
				out = append(out, d)
				break
			}
		}
	}
	return out
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

// recentDocs returns up to n docs sorted by filesystem modification time descending.
// If n is 0, all docs are returned (sorted).
func recentDocs(docs []search.Result, n int) []search.Result {
	sorted := make([]search.Result, len(docs))
	copy(sorted, docs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ModTime.After(sorted[j].ModTime)
	})
	if n > 0 && len(sorted) > n {
		return sorted[:n]
	}
	return sorted
}

// --- Cursor movement helpers ---

func (m Model) moveCursorDown() Model {
	switch {
	case m.activeView == ViewDashboard && m.dashRight:
		// Right column (todo preview) — no cursor movement.
	case m.activeView == ViewDashboard && !m.dashRight:
		items := buildDashItems(m)
		next := m.dashCursor + 1
		for next < len(items) && (items[next].isHeader || items[next].isBlank || items[next].isPlaceholder) {
			next++
		}
		if next < len(items) {
			m.dashCursor = next
		}
	case m.activeView == ViewJournal && m.activePanel == PanelFiles:
		rows := buildJournalRows(m.docs, m.journalExpanded)
		if m.journalCursor < len(rows)-1 {
			m.journalCursor++
		}
	case m.activeView == ViewKB && m.activePanel == PanelFiles:
		rows := buildKBRows(m.docs, m.kbExpanded)
		if m.kbCursor < len(rows)-1 {
			m.kbCursor++
		}
	case m.activeView == ViewTags && m.activePanel == PanelFiles:
		if m.tagCursor < len(m.tagEntries)-1 {
			m.tagCursor++
			m.tagDocCursor = 0 // reset doc cursor when tag changes
		}
	case m.activeView == ViewTags && m.activePanel == PanelPreview:
		if m.tagCursor < len(m.tagEntries) {
			tagged := m.taggedDocs()
			if m.tagDocCursor < len(tagged)-1 {
				m.tagDocCursor++
			}
		}
	case m.activePanel == PanelFiles:
		docs := m.visibleDocs()
		if m.fileCursor < len(docs)-1 {
			m.fileCursor++
		}
	}
	return m
}

func (m Model) moveCursorUp() Model {
	switch {
	case m.activeView == ViewDashboard && m.dashRight:
		// Right column (todo preview) — no cursor movement.
	case m.activeView == ViewDashboard && !m.dashRight:
		items := buildDashItems(m)
		prev := m.dashCursor - 1
		for prev >= 0 && (items[prev].isHeader || items[prev].isBlank || items[prev].isPlaceholder) {
			prev--
		}
		if prev >= 0 {
			m.dashCursor = prev
		}
	case m.activeView == ViewJournal && m.activePanel == PanelFiles:
		if m.journalCursor > 0 {
			m.journalCursor--
		}
	case m.activeView == ViewKB && m.activePanel == PanelFiles:
		if m.kbCursor > 0 {
			m.kbCursor--
		}
	case m.activeView == ViewTags && m.activePanel == PanelFiles:
		if m.tagCursor > 0 {
			m.tagCursor--
			m.tagDocCursor = 0 // reset doc cursor when tag changes
		}
	case m.activeView == ViewTags && m.activePanel == PanelPreview:
		if m.tagDocCursor > 0 {
			m.tagDocCursor--
		}
	case m.activePanel == PanelFiles:
		if m.fileCursor > 0 {
			m.fileCursor--
		}
	}
	return m
}

func (m Model) jumpTop() Model {
	if m.activePanel == PanelFiles {
		switch m.activeView {
		case ViewDashboard:
			// Move to first navigable (non-header, non-blank, non-placeholder) item.
			items := buildDashItems(m)
			for i, it := range items {
				if !it.isHeader && !it.isBlank && !it.isPlaceholder {
					m.dashCursor = i
					break
				}
			}
		case ViewJournal:
			m.journalCursor = 0
		case ViewKB:
			m.kbCursor = 0
		case ViewTags:
			m.tagCursor = 0
		default:
			m.fileCursor = 0
		}
	}
	return m
}

func (m Model) jumpBottom() Model {
	if m.activePanel == PanelFiles {
		switch m.activeView {
		case ViewJournal:
			rows := buildJournalRows(m.docs, m.journalExpanded)
			if len(rows) > 0 {
				m.journalCursor = len(rows) - 1
			}
		case ViewKB:
			rows := buildKBRows(m.docs, m.kbExpanded)
			if len(rows) > 0 {
				m.kbCursor = len(rows) - 1
			}
		case ViewTags:
			if len(m.tagEntries) > 0 {
				m.tagCursor = len(m.tagEntries) - 1
			}
		default:
			docs := m.visibleDocs()
			if len(docs) > 0 {
				m.fileCursor = len(docs) - 1
			}
		}
	}
	return m
}
