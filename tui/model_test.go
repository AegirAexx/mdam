package tui

import (
	"fmt"
	"path/filepath"
	"regexp"
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

// --- ANSI stripping helper ---

// ansiRE matches ANSI escape sequences (used to make View() assertions terminal-agnostic).
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes all ANSI escape codes from s.
func stripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

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

// --- stripANSI smoke test ---

func TestStripANSI(t *testing.T) {
	input := "\x1b[32mHello\x1b[0m World"
	got := stripANSI(input)
	want := "Hello World"
	if got != want {
		t.Errorf("stripANSI(%q) = %q, want %q", input, got, want)
	}
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
		{ModeDeleteConfirm, "DELETE?"},
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

func TestWindowResizeUpdatesViewport(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m2 := updated.(Model)
	// Viewport width should be set to the right panel width.
	leftWidth := 120 / 3
	rightWidth := 120 - leftWidth - 1
	if m2.preview.Width != rightWidth {
		t.Errorf("viewport width = %d, want %d", m2.preview.Width, rightWidth)
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
	view := stripANSI(m.View())
	if !strings.Contains(view, "NORMAL") {
		t.Errorf("View() does not contain NORMAL mode indicator")
	}
}

func TestViewHelpOverlay(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 40
	m.showHelp = true
	view := stripANSI(m.View())
	if !strings.Contains(view, "Keybindings") {
		t.Errorf("help view does not contain 'Keybindings': %q", view)
	}
}

func TestViewInSearchMode(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m = sendKey(m, "/")
	view := stripANSI(m.View())
	if !strings.Contains(view, "SEARCH") {
		t.Errorf("view in search mode missing SEARCH indicator")
	}
}

func TestViewInCommandMode(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m = sendKey(m, ":")
	view := stripANSI(m.View())
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

func TestStyledPanelHeaderFocused(t *testing.T) {
	th := NewTheme("tokyonight")
	icons := PlainIcons()
	h := stripANSI(styledPanelHeader("Files", true, 20, th, icons))
	if !strings.HasPrefix(h, "▶") {
		t.Errorf("styled focused header should start with ▶, got %q", h)
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

func TestVisibleDocsTagsReturnsNil(t *testing.T) {
	m := modelWithDocs()
	m.activeView = ViewTags
	docs := m.visibleDocs()
	if docs != nil {
		t.Errorf("ViewTags: visibleDocs should return nil, got %v", docs)
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
	m.pickerTemplates = []tmpl.Template{
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
	view := stripANSI(m.View())
	if !strings.Contains(view, "main") {
		t.Errorf("status bar missing branch name 'main' in view:\n%s", view)
	}
}

func TestStatusBarShowsDocCount(t *testing.T) {
	m := modelWithDocs()
	m.width = 80
	m.height = 24
	view := stripANSI(m.View())
	if !strings.Contains(view, "3 docs") {
		t.Errorf("status bar missing doc count '3 docs' in view:\n%s", view)
	}
}

func TestStatusBarLoadingState(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	// loading is true by default before docsLoadedMsg.
	view := stripANSI(m.View())
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
		{"1", ViewDashboard}, // key 1 now maps to dashboard
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

func TestView6SwitchesToTags(t *testing.T) {
	m := modelWithDocs()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("6")})
	m2 := sendKey(m, "6")
	if m2.activeView != ViewTags {
		t.Errorf("key 6: activeView = %v, want ViewTags", m2.activeView)
	}
	if cmd == nil {
		t.Error("key 6 should return cmdBuildTagIndex")
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
		{"nvim", "vim", "nvim"},
		{"", "vim", "vim"},
		{"", "", ""},
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

// --- Phase 5 tests ---

func TestDeleteKeyEntersConfirmMode(t *testing.T) {
	m := modelWithDocs()
	m = sendKey(m, "d")
	if m.mode != ModeDeleteConfirm {
		t.Errorf("d key: mode = %v, want ModeDeleteConfirm", m.mode)
	}
	if m.deleteConfirmPath == "" {
		t.Error("d key: deleteConfirmPath should be set")
	}
}

func TestDeleteKeyNoDocSelected(t *testing.T) {
	m := newTestModel()
	m = sendMsg(m, docsLoadedMsg{docs: []search.Result{}})
	m = sendKey(m, "d")
	if m.mode == ModeDeleteConfirm {
		t.Error("d with no doc: should not enter delete confirm mode")
	}
	if !strings.Contains(m.statusMsg, "no document") {
		t.Errorf("d with no doc: statusMsg = %q, want 'no document'", m.statusMsg)
	}
}

func TestDeleteConfirmYReturnsCmd(t *testing.T) {
	m := modelWithDocs()
	m = sendKey(m, "d") // enter confirm mode
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Error("confirming delete (y) should return cmdDeleteDoc")
	}
}

func TestDeleteConfirmNCancels(t *testing.T) {
	m := modelWithDocs()
	m = sendKey(m, "d")
	m2 := sendKey(m, "n")
	if m2.mode != ModeNormal {
		t.Errorf("n in delete confirm: mode = %v, want ModeNormal", m2.mode)
	}
	if m2.deleteConfirmPath != "" {
		t.Error("n in delete confirm: deleteConfirmPath should be cleared")
	}
}

func TestDeleteConfirmEscCancels(t *testing.T) {
	m := modelWithDocs()
	m = sendKey(m, "d")
	m2 := sendKey(m, "esc")
	if m2.mode != ModeNormal {
		t.Errorf("esc in delete confirm: mode = %v, want ModeNormal", m2.mode)
	}
}

func TestDeleteDoneMsgReloadsAndSetsStatus(t *testing.T) {
	m := modelWithDocs()
	m2, cmd := m.Update(deleteDoneMsg{path: "/notes/setup-nginx.md"})
	m3 := m2.(Model)
	if !strings.Contains(m3.statusMsg, "deleted") {
		t.Errorf("deleteDoneMsg: statusMsg = %q, want 'deleted'", m3.statusMsg)
	}
	if cmd == nil {
		t.Error("deleteDoneMsg should return cmdLoadDocs")
	}
}

func TestDeleteDoneMsgError(t *testing.T) {
	m := newTestModel()
	m2, _ := m.Update(deleteDoneMsg{path: "/notes/x.md", err: fmt.Errorf("permission denied")})
	m3 := m2.(Model)
	if !strings.Contains(m3.statusMsg, "delete failed") {
		t.Errorf("deleteDoneMsg error: statusMsg = %q, want 'delete failed'", m3.statusMsg)
	}
}

func TestPinKeyTogglesPin(t *testing.T) {
	m := modelWithDocs()
	path := fakeDocs[0].Path
	if m.pinnedPaths[path] {
		t.Fatal("path should not be pinned initially")
	}
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m3 := m2.(Model)
	if !m3.pinnedPaths[path] {
		t.Error("p key: path should be pinned after first press")
	}
	if cmd == nil {
		t.Error("p key: should return cmdSavePins")
	}
}

func TestPinKeyUnpins(t *testing.T) {
	m := modelWithDocs()
	m = sendKey(m, "p") // pin
	m2 := sendKey(m, "p") // unpin
	path := fakeDocs[0].Path
	if m2.pinnedPaths[path] {
		t.Error("second p press: path should be unpinned")
	}
}

func TestPinKeyNoDocSelected(t *testing.T) {
	m := newTestModel()
	m = sendMsg(m, docsLoadedMsg{docs: []search.Result{}})
	m = sendKey(m, "p")
	if !strings.Contains(m.statusMsg, "no document") {
		t.Errorf("p with no doc: statusMsg = %q, want 'no document'", m.statusMsg)
	}
}

func TestSmartFilterCycles(t *testing.T) {
	m := newTestModel()
	if m.smartFilter != SmartFilterNone {
		t.Fatal("smartFilter should start at None")
	}
	m = sendKey(m, "f")
	if m.smartFilter != SmartFilterUntagged {
		t.Errorf("after f: smartFilter = %v, want Untagged", m.smartFilter)
	}
	m = sendKey(m, "f")
	if m.smartFilter != SmartFilterStaleWeek {
		t.Errorf("after ff: smartFilter = %v, want StaleWeek", m.smartFilter)
	}
	m = sendKey(m, "f")
	if m.smartFilter != SmartFilterInbox {
		t.Errorf("after fff: smartFilter = %v, want Inbox", m.smartFilter)
	}
	m = sendKey(m, "f")
	if m.smartFilter != SmartFilterNone {
		t.Errorf("after ffff: smartFilter = %v, want None (wrapped)", m.smartFilter)
	}
}

func TestSmartFilterUntagged(t *testing.T) {
	tagged := search.Result{
		Path: "/notes/tagged.md",
		Frontmatter: document.Frontmatter{Tags: []string{"go"}},
	}
	untagged := search.Result{
		Path: "/notes/untagged.md",
		Frontmatter: document.Frontmatter{Tags: nil},
	}
	docs := []search.Result{tagged, untagged}
	result := applySmartFilter(docs, SmartFilterUntagged)
	if len(result) != 1 {
		t.Errorf("SmartFilterUntagged: got %d docs, want 1", len(result))
	}
	if result[0].Path != untagged.Path {
		t.Errorf("SmartFilterUntagged: got %q, want %q", result[0].Path, untagged.Path)
	}
}

func TestSmartFilterInbox(t *testing.T) {
	docs := []search.Result{
		{Path: "/a.md", Frontmatter: document.Frontmatter{Type: "unsorted"}},
		{Path: "/b.md", Frontmatter: document.Frontmatter{Type: "kb"}},
	}
	result := applySmartFilter(docs, SmartFilterInbox)
	if len(result) != 1 {
		t.Errorf("SmartFilterInbox: got %d docs, want 1", len(result))
	}
}

func TestPreviewReadyMsgUpdatesViewport(t *testing.T) {
	m := newTestModel()
	m.preview.Width = 60
	m.preview.Height = 20
	content := "# Hello\n\nThis is a preview."
	m2 := sendMsg(m, previewReadyMsg{content: content})
	if m2.preview.TotalLineCount() == 0 {
		t.Error("previewReadyMsg: viewport should have content")
	}
}

func TestPinsLoadedMsgUpdatesPins(t *testing.T) {
	m := newTestModel()
	pins := map[string]bool{"/notes/a.md": true}
	m2 := sendMsg(m, pinsLoadedMsg{pins: pins})
	if !m2.pinnedPaths["/notes/a.md"] {
		t.Error("pinsLoadedMsg: /notes/a.md should be pinned")
	}
}

func TestPinsLoadedMsgErrorIgnored(t *testing.T) {
	m := newTestModel()
	m2 := sendMsg(m, pinsLoadedMsg{err: fmt.Errorf("no pins file")})
	// On error, pins stay empty — no panic.
	if len(m2.pinnedPaths) != 0 {
		t.Errorf("pinsLoadedMsg error: pinnedPaths should remain empty, got %v", m2.pinnedPaths)
	}
}

func TestViewDashboardRendersWithoutPanic(t *testing.T) {
	m := modelWithDocs()
	m.width = 80
	m.height = 24
	m.activeView = ViewDashboard
	view := m.View()
	if view == "" {
		t.Error("ViewDashboard rendered empty string")
	}
}

func TestViewDashboardShowsStatusBar(t *testing.T) {
	m := modelWithDocs()
	m.width = 80
	m.height = 24
	m.activeView = ViewDashboard
	view := stripANSI(m.View())
	if !strings.Contains(view, "NORMAL") {
		t.Errorf("dashboard view missing NORMAL mode indicator:\n%s", view)
	}
}

func TestViewTagBrowserRendersWithoutPanic(t *testing.T) {
	m := modelWithDocs()
	m.width = 80
	m.height = 24
	m.activeView = ViewTags
	m.tagEntries = buildTagIndex(fakeDocs)
	view := m.View()
	if view == "" {
		t.Error("ViewTags rendered empty string")
	}
}

func TestSpinnerFrameAdvancesOnTick(t *testing.T) {
	m := newTestModel() // loading = true
	initial := m.spinnerFrame
	m2, cmd := m.Update(tickMsg{})
	m3 := m2.(Model)
	if m3.spinnerFrame == initial {
		t.Error("tickMsg should advance spinnerFrame")
	}
	// When loading, tick re-schedules.
	if cmd == nil {
		t.Error("tickMsg while loading should return cmdTick")
	}
}

func TestSpinnerStopsWhenNotLoading(t *testing.T) {
	m := modelWithDocs() // loading = false after docsLoadedMsg
	_, cmd := m.Update(tickMsg{})
	if cmd != nil {
		t.Error("tickMsg when not loading should not return a cmd")
	}
}

func TestSmartFilterStringNone(t *testing.T) {
	if SmartFilterNone.String() != "" {
		t.Errorf("SmartFilterNone.String() = %q, want empty", SmartFilterNone.String())
	}
}

func TestSmartFilterStrings(t *testing.T) {
	tests := []struct {
		f    SmartFilter
		want string
	}{
		{SmartFilterUntagged, "filter: untagged"},
		{SmartFilterStaleWeek, "filter: stale (>7 days)"},
		{SmartFilterInbox, "filter: inbox"},
	}
	for _, tt := range tests {
		if got := tt.f.String(); got != tt.want {
			t.Errorf("SmartFilter(%d).String() = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestNewModelHasThemeAndIcons(t *testing.T) {
	m := New(config.Config{Theme: "nord"})
	if m.theme.GlamourStyle == "" {
		t.Error("New() should initialize theme with non-empty GlamourStyle")
	}
	if m.icons.CursorSel == "" {
		t.Error("New() should initialize icons with non-empty CursorSel")
	}
}

func TestNewModelNerdFonts(t *testing.T) {
	plain := New(config.Config{NerdFonts: false})
	nerd := New(config.Config{NerdFonts: true})
	if plain.icons.CursorSel == nerd.icons.CursorSel {
		t.Error("NerdFonts=false and NerdFonts=true should produce different icons")
	}
}

func TestViewTagsKeyBuildTagIndex(t *testing.T) {
	m := modelWithDocs()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("6")})
	if cmd == nil {
		t.Error("key 6 should return cmdBuildTagIndex")
	}
}

func TestDeleteModeStringIs(t *testing.T) {
	if ModeDeleteConfirm.String() != "DELETE?" {
		t.Errorf("ModeDeleteConfirm.String() = %q, want DELETE?", ModeDeleteConfirm.String())
	}
}

// TestViewShowsFileNames verifies that file names appear in the file panel.
func TestViewShowsFileNames(t *testing.T) {
	m := modelWithDocs()
	m.width = 80
	m.height = 24
	view := stripANSI(m.View())
	if !strings.Contains(view, "2026-03-14.md") {
		t.Errorf("view should contain filename '2026-03-14.md':\n%s", view)
	}
}

// TestViewShowsDeleteConfirmStatus verifies the delete confirm status message.
func TestViewShowsDeleteConfirmStatus(t *testing.T) {
	m := modelWithDocs()
	m.width = 80
	m.height = 24
	m = sendKey(m, "d")
	view := stripANSI(m.View())
	if !strings.Contains(view, "Delete") {
		t.Errorf("delete confirm view should contain 'Delete':\n%s", view)
	}
}

// --- Issue #3 tests: template picker filtering ---

// TestNKeyFiltersPickerToJournalAndKB asserts that pressing "n" with all 5
// built-in templates loaded results in exactly 2 pickerTemplates: journal and kb.
func TestNKeyFiltersPickerToJournalAndKB(t *testing.T) {
	m := newTestModel()
	builtins := tmpl.BuiltinTemplates()
	for name, content := range builtins {
		m.templates = append(m.templates, tmpl.Template{Name: name, Content: content})
	}
	if len(m.templates) != 5 {
		t.Fatalf("expected 5 built-in templates, got %d", len(m.templates))
	}

	m = sendKey(m, "n")

	if m.mode != ModeTemplatePicker {
		t.Fatalf("n key: mode = %v, want ModeTemplatePicker", m.mode)
	}
	if len(m.pickerTemplates) != 2 {
		names := make([]string, len(m.pickerTemplates))
		for i, t := range m.pickerTemplates {
			names[i] = t.Name
		}
		t.Errorf("pickerTemplates count = %d, want 2; names = %v", len(m.pickerTemplates), names)
	}
	for _, pt := range m.pickerTemplates {
		if pt.Name != "journal" && pt.Name != "kb" {
			t.Errorf("unexpected template in picker: %q", pt.Name)
		}
	}
}

// TestPickerViewShowsTypeHeader verifies the picker header says "Select Type".
func TestPickerViewShowsTypeHeader(t *testing.T) {
	m := newTestModel()
	m.mode = ModeTemplatePicker
	m.pickerTemplates = []tmpl.Template{
		{Name: "journal"},
		{Name: "kb"},
	}
	m.width = 80
	m.height = 24
	view := stripANSI(m.View())
	if !strings.Contains(view, "Select Type") {
		t.Errorf("picker view should contain 'Select Type': %q", view)
	}
}

// --- Issue #2 tests: journal creation flow ---

// TestJournalTemplateSelectionBypassesVarMode verifies that selecting the journal
// template does not enter ModeTemplateVars.
func TestJournalTemplateSelectionBypassesVarMode(t *testing.T) {
	journalContent := tmpl.BuiltinTemplates()["journal"]
	m := newTestModel()
	m.mode = ModeTemplatePicker
	m.pickerTemplates = []tmpl.Template{
		{Name: "journal", Content: journalContent},
	}
	m.pickerCursor = 0

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := m2.(Model)

	if m3.mode == ModeTemplateVars {
		t.Error("journal selection should not enter ModeTemplateVars")
	}
	if m3.mode != ModeNormal {
		t.Errorf("journal selection: mode = %v, want ModeNormal", m3.mode)
	}
	if cmd == nil {
		t.Error("journal selection should return cmdJournalCreate")
	}
}

// TestKBTemplateSelectionEntersVarMode verifies that selecting the kb template
// (which has {{title}}) enters ModeTemplateVars.
func TestKBTemplateSelectionEntersVarMode(t *testing.T) {
	kbContent := tmpl.BuiltinTemplates()["kb"]
	m := newTestModel()
	m.mode = ModeTemplatePicker
	m.pickerTemplates = []tmpl.Template{
		{Name: "kb", Content: kbContent},
	}
	m.pickerCursor = 0

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := m2.(Model)

	if m3.mode != ModeTemplateVars {
		t.Errorf("kb selection: mode = %v, want ModeTemplateVars", m3.mode)
	}
	if len(m3.varNames) == 0 {
		t.Error("kb selection: varNames should not be empty (has {{title}})")
	}
}

// TestKBTemplateNoBuiltinVarsPrompted verifies that built-in vars (date_short)
// are not included in the var prompt after render-first.
func TestKBTemplateNoBuiltinVarsPrompted(t *testing.T) {
	kbContent := tmpl.BuiltinTemplates()["kb"]
	m := newTestModel()
	m.mode = ModeTemplatePicker
	m.pickerTemplates = []tmpl.Template{
		{Name: "kb", Content: kbContent},
	}
	m.pickerCursor = 0

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := m2.(Model)

	for _, v := range m3.varNames {
		if v == "date_short" || v == "date" {
			t.Errorf("built-in var %q should not appear in user prompts", v)
		}
	}
}

func TestSmartFilterStaleWeek(t *testing.T) {
	stale := search.Result{
		Path: "/notes/stale.md",
		Frontmatter: document.Frontmatter{
			Modified: time.Now().AddDate(0, 0, -10),
		},
	}
	recent := search.Result{
		Path: "/notes/recent.md",
		Frontmatter: document.Frontmatter{
			Modified: time.Now().AddDate(0, 0, -3),
		},
	}
	result := applySmartFilter([]search.Result{stale, recent}, SmartFilterStaleWeek)
	if len(result) != 1 {
		t.Fatalf("SmartFilterStaleWeek: got %d results, want 1", len(result))
	}
	if result[0].Path != stale.Path {
		t.Errorf("SmartFilterStaleWeek: got %q, want %q", result[0].Path, stale.Path)
	}
}

// reusable for other packages
var _ = filepath.Join // keep filepath import used
