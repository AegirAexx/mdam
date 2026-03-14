package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/AegirAexx/mdam/internal/config"
	"github.com/AegirAexx/mdam/internal/document"
	"github.com/AegirAexx/mdam/internal/git"
	"github.com/AegirAexx/mdam/internal/search"
	tmpl "github.com/AegirAexx/mdam/internal/template"
	"github.com/AegirAexx/mdam/internal/todo"
)

// --- Test helpers ---

// newTestModel returns a Model with a zero config suitable for unit tests.
func newTestModel() Model {
	return New(config.Config{})
}

// sendKey sends a key message through Update and returns the updated Model.
func sendKey(m Model, k string) Model {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	switch k {
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "ctrl+g":
		msg = tea.KeyMsg{Type: tea.KeyCtrlG}
	}
	updated, _ := m.Update(msg)
	return updated.(Model)
}

// sendMsg sends an arbitrary message through Update and returns the updated Model.
func sendMsg(m Model, msg tea.Msg) Model {
	updated, _ := m.Update(msg)
	return updated.(Model)
}

// fakeDocs is a small set of search.Result for use in tests.
var fakeDocs = []search.Result{
	{
		Path: "/notes/2026-03-14.md",
		Frontmatter: document.Frontmatter{
			Title:    "Journal 2026-03-14",
			Type:     "journal",
			Tags:     []string{"daily"},
			Modified: time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC),
		},
	},
	{
		Path: "/notes/2026-03-13.md",
		Frontmatter: document.Frontmatter{
			Title:    "Journal 2026-03-13",
			Type:     "journal",
			Tags:     []string{"daily"},
			Modified: time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC),
		},
	},
	{
		Path: "/notes/setup-nginx.md",
		Frontmatter: document.Frontmatter{
			Title:    "Setup Nginx",
			Type:     "kb",
			Tags:     []string{"devops"},
			Modified: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		},
	},
}

// fakeTasks returns a few open tasks for testing.
var fakeTasks = []todo.Task{
	{Raw: "- [ ] Review PR #42 @work", Status: "open", Text: "Review PR #42"},
	{Raw: "- [ ] Buy groceries @personal", Status: "open", Text: "Buy groceries"},
}

// modelWithDocs returns a Model pre-loaded with fakeDocs.
func modelWithDocs() Model {
	m := newTestModel()
	m = sendMsg(m, docsLoadedMsg{docs: fakeDocs})
	return m
}

// --- Phase 2 baseline tests (updated for new API) ---

func TestNewModel(t *testing.T) {
	m := newTestModel()
	if m.mode != ModeNormal {
		t.Errorf("mode = %v, want ModeNormal", m.mode)
	}
	if m.activePanel != PanelFiles {
		t.Errorf("activePanel = %v, want PanelFiles", m.activePanel)
	}
	if !m.loading {
		t.Error("loading should be true before Init completes")
	}
}

func TestModeString(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeNormal, "NORMAL"},
		{ModeCommand, "COMMAND"},
		{ModeSearch, "SEARCH"},
		{ModeTemplatePicker, "NEW DOC"},
		{ModeTemplateVars, "NEW DOC"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("Mode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestPanelCycle(t *testing.T) {
	p := PanelFiles
	if p.next() != PanelPreview {
		t.Errorf("PanelFiles.next() = %v, want PanelPreview", p.next())
	}
	if p.prev() != PanelTodo {
		t.Errorf("PanelFiles.prev() = %v, want PanelTodo", p.prev())
	}
	cur := PanelFiles
	for i := 0; i < int(panelCount); i++ {
		cur = cur.next()
	}
	if cur != PanelFiles {
		t.Errorf("full next cycle did not return to PanelFiles, got %v", cur)
	}
}

func TestCursorMovement(t *testing.T) {
	m := modelWithDocs()
	initial := m.fileCursor // 0

	m = sendKey(m, "j")
	if m.fileCursor != initial+1 {
		t.Errorf("after j: fileCursor = %d, want %d", m.fileCursor, initial+1)
	}

	m = sendKey(m, "k")
	if m.fileCursor != initial {
		t.Errorf("after k: fileCursor = %d, want %d", m.fileCursor, initial)
	}

	m = sendKey(m, "k")
	if m.fileCursor != 0 {
		t.Errorf("k at top: fileCursor = %d, want 0", m.fileCursor)
	}
}

func TestCursorMovementArrowKeys(t *testing.T) {
	m := modelWithDocs()
	m = sendKey(m, "down")
	if m.fileCursor != 1 {
		t.Errorf("after down: fileCursor = %d, want 1", m.fileCursor)
	}
	m = sendKey(m, "up")
	if m.fileCursor != 0 {
		t.Errorf("after up: fileCursor = %d, want 0", m.fileCursor)
	}
}

func TestJumpBottom(t *testing.T) {
	m := modelWithDocs()
	m = sendKey(m, "G")
	want := len(fakeDocs) - 1
	if m.fileCursor != want {
		t.Errorf("G: fileCursor = %d, want %d", m.fileCursor, want)
	}
}

func TestGGChord(t *testing.T) {
	m := modelWithDocs()
	m = sendKey(m, "G")
	if m.fileCursor == 0 {
		t.Skip("docs has only one item")
	}
	m = sendKey(m, "g")
	if m.fileCursor != len(fakeDocs)-1 {
		t.Errorf("after single g, cursor should not move yet")
	}
	m = sendKey(m, "g")
	if m.fileCursor != 0 {
		t.Errorf("after gg: fileCursor = %d, want 0", m.fileCursor)
	}
}

func TestTabCyclesPanels(t *testing.T) {
	m := newTestModel()
	if m.activePanel != PanelFiles {
		t.Fatalf("initial panel = %v, want PanelFiles", m.activePanel)
	}
	m = sendKey(m, "tab")
	if m.activePanel != PanelPreview {
		t.Errorf("after tab: panel = %v, want PanelPreview", m.activePanel)
	}
	m = sendKey(m, "tab")
	if m.activePanel != PanelTodo {
		t.Errorf("after second tab: panel = %v, want PanelTodo", m.activePanel)
	}
	m = sendKey(m, "tab")
	if m.activePanel != PanelFiles {
		t.Errorf("after third tab: panel = %v, want PanelFiles", m.activePanel)
	}
}

func TestShiftTabCyclesPanelsReverse(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, "shift+tab")
	if m.activePanel != PanelTodo {
		t.Errorf("shift+tab from PanelFiles = %v, want PanelTodo", m.activePanel)
	}
}

func TestEnterSearchMode(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, "/")
	if m.mode != ModeSearch {
		t.Errorf("mode = %v, want ModeSearch", m.mode)
	}
}

func TestEnterCommandMode(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, ":")
	if m.mode != ModeCommand {
		t.Errorf("mode = %v, want ModeCommand", m.mode)
	}
}

func TestEscapeFromSearch(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, "/")
	m = sendKey(m, "esc")
	if m.mode != ModeNormal {
		t.Errorf("after esc from search: mode = %v, want ModeNormal", m.mode)
	}
}

func TestEscapeFromCommand(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, ":")
	m = sendKey(m, "esc")
	if m.mode != ModeNormal {
		t.Errorf("after esc from command: mode = %v, want ModeNormal", m.mode)
	}
}

func TestHelpToggle(t *testing.T) {
	m := newTestModel()
	if m.showHelp {
		t.Error("showHelp should be false initially")
	}
	m = sendKey(m, "?")
	if !m.showHelp {
		t.Error("showHelp should be true after ?")
	}
	m = sendKey(m, "?")
	if m.showHelp {
		t.Error("showHelp should be false after second ?")
	}
}

func TestWindowResize(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m2 := updated.(Model)
	if m2.width != 120 || m2.height != 40 {
		t.Errorf("WindowSizeMsg: got %dx%d, want 120x40", m2.width, m2.height)
	}
}

func TestTodoCursorMovement(t *testing.T) {
	m := newTestModel()
	m = sendMsg(m, todosLoadedMsg{tasks: fakeTasks})
	// Switch to todo panel.
	m = sendKey(m, "4")
	m = sendKey(m, "j")
	if m.todoCursor != 1 {
		t.Errorf("todo cursor after j = %d, want 1", m.todoCursor)
	}
	m = sendKey(m, "k")
	if m.todoCursor != 0 {
		t.Errorf("todo cursor after k = %d, want 0", m.todoCursor)
	}
}

func TestViewRendersWithoutPanic(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestViewContainsModeIndicator(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	view := m.View()
	if !strings.Contains(view, "NORMAL") {
		t.Errorf("View() does not contain NORMAL mode indicator")
	}
}

func TestViewHelpOverlay(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 40
	m.showHelp = true
	view := m.View()
	if !strings.Contains(view, "Keybindings") {
		t.Errorf("help view does not contain 'Keybindings': %q", view)
	}
}

func TestViewInSearchMode(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m = sendKey(m, "/")
	view := m.View()
	if !strings.Contains(view, "SEARCH") {
		t.Errorf("view in search mode missing SEARCH indicator")
	}
}

func TestViewInCommandMode(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m = sendKey(m, ":")
	view := m.View()
	if !strings.Contains(view, "COMMAND") {
		t.Errorf("view in command mode missing COMMAND indicator")
	}
}

func TestSearchEnterReturnsToNormal(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, "/")
	m = sendKey(m, "enter")
	if m.mode != ModeNormal {
		t.Errorf("after search Enter: mode = %v, want ModeNormal", m.mode)
	}
}

func TestCommandEnterReturnsToNormal(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, ":")
	m = sendKey(m, "enter")
	if m.mode != ModeNormal {
		t.Errorf("after command Enter: mode = %v, want ModeNormal", m.mode)
	}
}

func TestInit(t *testing.T) {
	m := newTestModel()
	cmd := m.Init()
	// Phase 3: Init returns a batch command (not nil).
	if cmd == nil {
		t.Error("Init() should return a command in Phase 3")
	}
}

// --- Layout helper tests ---

func TestPadRight(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"hi", 5, "hi   "},
		{"hello", 5, "hello"},
		{"toolong", 4, "tool"},
		{"", 3, "   "},
	}
	for _, tt := range tests {
		got := padRight(tt.s, tt.n)
		if got != tt.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello w…"},
		{"hi", 2, "hi"},
		{"hi", 1, "…"},
		{"hi", 0, ""},
	}
	for _, tt := range tests {
		got := truncate(tt.s, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
		}
	}
}

func TestPanelHeader(t *testing.T) {
	h := panelHeader("Files", true, 20)
	if !strings.HasPrefix(h, "▶") {
		t.Errorf("focused header should start with ▶, got %q", h)
	}
	h2 := panelHeader("Files", false, 20)
	if !strings.HasPrefix(h2, "─") {
		t.Errorf("unfocused header should start with ─, got %q", h2)
	}
}

// --- Phase 3 tests ---

func TestDocsLoaded(t *testing.T) {
	m := newTestModel()
	m = sendMsg(m, docsLoadedMsg{docs: fakeDocs})
	if m.loading {
		t.Error("loading should be false after docsLoadedMsg")
	}
	if len(m.docs) != len(fakeDocs) {
		t.Errorf("docs count = %d, want %d", len(m.docs), len(fakeDocs))
	}
}

func TestDocsLoadedError(t *testing.T) {
	m := newTestModel()
	m = sendMsg(m, docsLoadedMsg{err: fmt.Errorf("disk read error")})
	if m.errorMsg == "" {
		t.Error("errorMsg should be set on load error")
	}
}

func TestTodosLoaded(t *testing.T) {
	m := newTestModel()
	m = sendMsg(m, todosLoadedMsg{tasks: fakeTasks})
	if len(m.todos) != len(fakeTasks) {
		t.Errorf("todos count = %d, want %d", len(m.todos), len(fakeTasks))
	}
}

func TestGitStatusLoaded(t *testing.T) {
	m := newTestModel()
	m.cfg.BaseDir = "/notes"
	status := git.RepoStatus{
		Branch: "main",
		Ahead:  2,
		Files: []git.FileStatus{
			{Path: "2026-03-14.md", X: ' ', Y: 'M'},
		},
	}
	m = sendMsg(m, gitStatusMsg{status: status})
	if m.gitStatus.Branch != "main" {
		t.Errorf("branch = %q, want %q", m.gitStatus.Branch, "main")
	}
	if m.gitStatus.Ahead != 2 {
		t.Errorf("ahead = %d, want 2", m.gitStatus.Ahead)
	}
}

func TestSearchResults(t *testing.T) {
	m := newTestModel()
	results := fakeDocs[:1]
	m = sendMsg(m, searchDoneMsg{results: results, query: "nginx"})
	if !m.searchActive {
		t.Error("searchActive should be true after searchDoneMsg")
	}
	if len(m.searchResults) != 1 {
		t.Errorf("searchResults count = %d, want 1", len(m.searchResults))
	}
}

func TestVisibleDocsAll(t *testing.T) {
	m := modelWithDocs()
	m.activeView = ViewAll
	docs := m.visibleDocs()
	if len(docs) != len(fakeDocs) {
		t.Errorf("ViewAll: visible docs = %d, want %d", len(docs), len(fakeDocs))
	}
}

func TestVisibleDocsJournal(t *testing.T) {
	m := modelWithDocs()
	m.activeView = ViewJournal
	docs := m.visibleDocs()
	for _, d := range docs {
		if d.Frontmatter.Type != "journal" {
			t.Errorf("ViewJournal: got doc with type %q", d.Frontmatter.Type)
		}
	}
	if len(docs) != 2 {
		t.Errorf("ViewJournal: expected 2 journal docs, got %d", len(docs))
	}
}

func TestVisibleDocsKB(t *testing.T) {
	m := modelWithDocs()
	m.activeView = ViewKB
	docs := m.visibleDocs()
	for _, d := range docs {
		if d.Frontmatter.Type != "kb" {
			t.Errorf("ViewKB: got doc with type %q", d.Frontmatter.Type)
		}
	}
	if len(docs) != 1 {
		t.Errorf("ViewKB: expected 1 kb doc, got %d", len(docs))
	}
}

func TestVisibleDocsRecent(t *testing.T) {
	m := modelWithDocs()
	m.activeView = ViewRecent
	docs := m.visibleDocs()
	// Should be sorted by Modified descending.
	for i := 1; i < len(docs); i++ {
		if docs[i].Frontmatter.Modified.After(docs[i-1].Frontmatter.Modified) {
			t.Errorf("ViewRecent not sorted by Modified desc at index %d", i)
		}
	}
}

func TestVisibleDocsSearch(t *testing.T) {
	m := modelWithDocs()
	m.searchResults = fakeDocs[:1]
	m.searchActive = true
	docs := m.visibleDocs()
	if len(docs) != 1 {
		t.Errorf("searchActive: visible docs = %d, want 1", len(docs))
	}
}

func TestGitMarkerModified(t *testing.T) {
	fs := git.FileStatus{X: ' ', Y: 'M', Path: "test.md"}
	marker := gitMarkerForStatus(fs)
	if marker != "[M]" {
		t.Errorf("modified marker = %q, want [M]", marker)
	}
}

func TestGitMarkerUntracked(t *testing.T) {
	fs := git.FileStatus{X: '?', Y: '?', Path: "test.md"}
	marker := gitMarkerForStatus(fs)
	if marker != "[?]" {
		t.Errorf("untracked marker = %q, want [?]", marker)
	}
}

func TestGitMarkerStaged(t *testing.T) {
	fs := git.FileStatus{X: 'A', Y: ' ', Path: "test.md"}
	marker := gitMarkerForStatus(fs)
	if marker != "[A]" {
		t.Errorf("staged marker = %q, want [A]", marker)
	}
}

func TestGitMarkerNone(t *testing.T) {
	fs := git.FileStatus{X: ' ', Y: ' ', Path: "test.md"}
	marker := gitMarkerForStatus(fs)
	if marker != "" {
		t.Errorf("clean file marker = %q, want empty", marker)
	}
}

func TestCommandQuit(t *testing.T) {
	m := newTestModel()
	_, cmd := m.executeCommand("q")
	if cmd == nil {
		t.Error("executeCommand(q) should return tea.Quit cmd")
	}
}

func TestCommandTodoSweep(t *testing.T) {
	m := newTestModel()
	_, cmd := m.executeCommand("todo sweep")
	if cmd == nil {
		t.Error("executeCommand(todo sweep) should return a command")
	}
}

func TestCommandUnknown(t *testing.T) {
	m := newTestModel()
	m2, cmd := m.executeCommand("foobar")
	m3 := m2.(Model)
	if cmd != nil {
		t.Error("executeCommand(unknown) should return nil command")
	}
	if !strings.Contains(m3.statusMsg, "foobar") {
		t.Errorf("unknown command status = %q, want 'foobar' in it", m3.statusMsg)
	}
}

func TestTemplatePickerNavigation(t *testing.T) {
	m := newTestModel()
	m.mode = ModeTemplatePicker
	m.templates = []tmpl.Template{
		{Name: "journal"},
		{Name: "kb"},
	}
	m.pickerCursor = 0

	m = sendKey(m, "j")
	if m.pickerCursor != 1 {
		t.Errorf("picker cursor after j = %d, want 1", m.pickerCursor)
	}
	m = sendKey(m, "k")
	if m.pickerCursor != 0 {
		t.Errorf("picker cursor after k = %d, want 0", m.pickerCursor)
	}
}

func TestTemplatePickerEscape(t *testing.T) {
	m := newTestModel()
	m.mode = ModeTemplatePicker
	m = sendKey(m, "esc")
	if m.mode != ModeNormal {
		t.Errorf("esc from picker: mode = %v, want ModeNormal", m.mode)
	}
}

func TestStatusBarShowsBranch(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m.gitStatus = git.RepoStatus{Branch: "main"}
	m.loading = false
	view := m.View()
	if !strings.Contains(view, "main") {
		t.Errorf("status bar missing branch name 'main' in view:\n%s", view)
	}
}

func TestStatusBarShowsDocCount(t *testing.T) {
	m := modelWithDocs()
	m.width = 80
	m.height = 24
	view := m.View()
	if !strings.Contains(view, "3 docs") {
		t.Errorf("status bar missing doc count '3 docs' in view:\n%s", view)
	}
}

func TestStatusBarLoadingState(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	// loading is true by default before docsLoadedMsg.
	view := m.View()
	if !strings.Contains(view, "scanning") {
		t.Errorf("loading state should show 'scanning' in view:\n%s", view)
	}
}

func TestViewSwitching(t *testing.T) {
	m := newTestModel()
	tests := []struct {
		key      string
		wantView View
	}{
		{"1", ViewAll},
		{"2", ViewJournal},
		{"3", ViewKB},
		{"5", ViewRecent},
	}
	for _, tt := range tests {
		m2 := sendKey(m, tt.key)
		if m2.activeView != tt.wantView {
			t.Errorf("key %q: activeView = %v, want %v", tt.key, m2.activeView, tt.wantView)
		}
	}
}

func TestView4SwitchesToTodoPanel(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, "4")
	if m.activePanel != PanelTodo {
		t.Errorf("key 4: activePanel = %v, want PanelTodo", m.activePanel)
	}
	if m.activeView != ViewTodo {
		t.Errorf("key 4: activeView = %v, want ViewTodo", m.activeView)
	}
}

func TestRescanSetsLoading(t *testing.T) {
	m := modelWithDocs() // loading is false
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")})
	m3 := m2.(Model)
	if !m3.loading {
		t.Error("R key should set loading = true")
	}
	if cmd == nil {
		t.Error("R key should return a command")
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Setup Nginx", "setup-nginx"},
		{"My KB Doc", "my-kb-doc"},
		{"already-kebab", "already-kebab"},
		{"", "untitled"},
		{"UPPERCASE", "uppercase"},
		{"underscore_sep", "underscore-sep"},
	}
	for _, tt := range tests {
		got := toKebabCase(tt.input)
		if got != tt.want {
			t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExportNoDocSelected(t *testing.T) {
	m := newTestModel()
	m = sendMsg(m, docsLoadedMsg{docs: []search.Result{}})
	m = sendKey(m, "e")
	if !strings.Contains(m.statusMsg, "no document") {
		t.Errorf("export with no doc: statusMsg = %q, want 'no document'", m.statusMsg)
	}
}

// --- Phase 4 tests ---

func TestResolveEditor(t *testing.T) {
	tests := []struct {
		cfgEditor string
		envEditor string
		want      string
	}{
		{"nvim", "vim", "nvim"},   // config takes precedence over $EDITOR
		{"", "vim", "vim"},        // falls back to $EDITOR
		{"", "", ""},              // empty when neither is set
	}
	for _, tt := range tests {
		t.Setenv("EDITOR", tt.envEditor)
		got := resolveEditor(tt.cfgEditor)
		if got != tt.want {
			t.Errorf("resolveEditor(%q) with $EDITOR=%q = %q, want %q",
				tt.cfgEditor, tt.envEditor, got, tt.want)
		}
	}
}

func TestEditorReturnTriggersRescan(t *testing.T) {
	m := modelWithDocs()
	m.loading = false
	m2, cmd := m.Update(editorReturnMsg{})
	m3 := m2.(Model)
	if !m3.loading {
		t.Error("editorReturnMsg should set loading = true")
	}
	if cmd == nil {
		t.Error("editorReturnMsg should return rescan commands")
	}
	if m3.statusMsg != "" {
		t.Errorf("editorReturnMsg success: statusMsg = %q, want empty", m3.statusMsg)
	}
}

func TestEditorReturnErrorSetsStatus(t *testing.T) {
	m := newTestModel()
	m = sendMsg(m, editorReturnMsg{err: fmt.Errorf("nvim crashed")})
	if !strings.Contains(m.statusMsg, "editor error") {
		t.Errorf("editorReturnMsg with err: statusMsg = %q, want 'editor error'", m.statusMsg)
	}
}

func TestEnterNoDocSelected(t *testing.T) {
	m := newTestModel()
	m = sendMsg(m, docsLoadedMsg{docs: []search.Result{}})
	m.cfg.Editor = "vi"
	m = sendKey(m, "enter")
	if !strings.Contains(m.statusMsg, "no document") {
		t.Errorf("Enter with no doc: statusMsg = %q, want 'no document'", m.statusMsg)
	}
}

func TestEnterNoEditorConfigured(t *testing.T) {
	m := modelWithDocs()
	m.cfg.Editor = ""
	t.Setenv("EDITOR", "")
	m = sendKey(m, "enter")
	if !strings.Contains(m.statusMsg, "no editor") {
		t.Errorf("Enter with no editor: statusMsg = %q, want 'no editor'", m.statusMsg)
	}
}

func TestEnterWithDocReturnsCmd(t *testing.T) {
	m := modelWithDocs()
	m.cfg.Editor = "vi"
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("Enter with doc selected should return a tea.Cmd (editor handoff)")
	}
}

func TestScratchReadyMsgOpensEditor(t *testing.T) {
	m := newTestModel()
	m.cfg.Editor = "vi"
	_, cmd := m.Update(scratchReadyMsg{path: "/notes/scratch.md"})
	if cmd == nil {
		t.Error("scratchReadyMsg should return cmdOpenEditor")
	}
}

func TestScratchReadyMsgNoEditor(t *testing.T) {
	m := newTestModel()
	m.cfg.Editor = ""
	t.Setenv("EDITOR", "")
	m2, cmd := m.Update(scratchReadyMsg{path: "/notes/scratch.md"})
	m3 := m2.(Model)
	if cmd != nil {
		t.Error("scratchReadyMsg with no editor should return nil cmd")
	}
	if !strings.Contains(m3.statusMsg, "no editor") {
		t.Errorf("scratchReadyMsg no editor: statusMsg = %q, want 'no editor'", m3.statusMsg)
	}
}

func TestLazygitDisabled(t *testing.T) {
	m := newTestModel()
	m.cfg.Git.Lazygit = false
	m = sendKey(m, "ctrl+g")
	if !strings.Contains(m.statusMsg, "disabled") {
		t.Errorf("ctrl+g lazygit disabled: statusMsg = %q, want 'disabled'", m.statusMsg)
	}
}

func TestLazygitEnabledReturnsCmd(t *testing.T) {
	m := newTestModel()
	m.cfg.Git.Lazygit = true
	m.cfg.BaseDir = "/notes"
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlG})
	if cmd == nil {
		t.Error("ctrl+g with lazygit enabled should return a tea.Cmd")
	}
}
