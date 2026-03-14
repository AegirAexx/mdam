# Phase 3 Report — Integration

**Completed:** 2026-03-14
**Goal:** Connect the Phase 1 engine to the Phase 2 TUI skeleton. The TUI stops showing dummy data and starts reading from the filesystem, responding to real commands, and displaying live git status.

---

## Summary

Phase 3 is complete. The TUI is now a functional tool backed by real data. All dummy content has been removed. On startup, the application fires three parallel async commands to load documents, open tasks, and git status. The file panel displays real filenames from `cfg.BaseDir` with git status markers. The status bar shows the current branch, ahead/behind count, and document count. Search, export, view switching, rescan, TODO sweep/archive, and the template picker are all wired to the engine.

**27 new tests. 54 total TUI tests passing. All 10 packages pass. `go vet` clean.**

---

## What Was Built

### New Files

| File | Lines | Purpose |
|---|---|---|
| `tui/messages.go` | 49 | Async message types for all engine operations |
| `tui/commands.go` | 134 | `tea.Cmd` factories that call engine functions on a goroutine |

### Modified Files

| File | Lines (before → after) | Changes |
|---|---|---|
| `tui/mode.go` | 61 → 77 | Added `ModeTemplatePicker`, `ModeTemplateVars`, `View` enum |
| `tui/model.go` | 338 → 615 | Complete rewrite: real fields, real Init/Update, all wiring |
| `tui/view.go` | 236 → 361 | Real file list, git markers, real status bar, picker overlays |
| `tui/tui.go` | 44 → 46 | `Run(cfg config.Config) error` signature |
| `internal/cli/root.go` | 46 → 46 | Passes loaded `cfg` to `tui.Run(cfg)` |
| `tui/model_test.go` | 366 → 745 | Updated Phase 2 tests, 27 new Phase 3 tests |

---

## Architecture

### Async Data Loading

Phase 3 follows the BubbleTea pattern for I/O: all engine calls happen in `tea.Cmd` goroutines and communicate back to the model via typed messages. The model never blocks.

```
Init() → tea.Batch(
    cmdLoadDocs(cfg.BaseDir)      → docsLoadedMsg
    cmdLoadTodos(cfg.TodoPath())  → todosLoadedMsg
    cmdLoadGitStatus(cfg.BaseDir) → gitStatusMsg
)
```

Every async operation follows the same structure in `commands.go`:

```go
func cmdLoadDocs(baseDir string) tea.Cmd {
    return func() tea.Msg {
        docs, err := search.ListAll(baseDir)
        return docsLoadedMsg{docs: docs, err: err}
    }
}
```

### Message Types (`tui/messages.go`)

| Message | Sent when |
|---|---|
| `docsLoadedMsg` | Directory scan completes |
| `todosLoadedMsg` | TODO file is read |
| `gitStatusMsg` | `git status` returns |
| `searchDoneMsg` | Fuzzy search completes |
| `exportDoneMsg` | Export to file completes |
| `sweepDoneMsg` | TODO sweep or archive completes |
| `fileCreatedMsg` | New document written to disk |

### Updated Model (`tui/model.go`)

```
Model
├── cfg           config.Config       // entire config available everywhere
├── mode          Mode                // Normal | Command | Search | TemplatePicker | TemplateVars
├── activePanel   PanelID             // Files | Preview | Todo
├── activeView    View                // All | Journal | KB | Todo | Recent
├── docs          []search.Result     // all documents from last scan
├── fileCursor    int
├── todos         []todo.Task         // open tasks from global TODO file
├── todoCursor    int
├── gitStatus     git.RepoStatus      // branch, ahead, behind
├── gitFileMap    map[string]git.FileStatus  // path → status, O(1) lookup
├── searchResults []search.Result     // results of last search
├── searchActive  bool                // true = show searchResults in file panel
├── templates     []template.Template // discovered templates for picker
├── pickerCursor  int
├── pendingTmpl   template.Template   // selected template awaiting var input
├── varNames      []string            // unresolved {{vars}} in pending template
├── varValues     []string            // user-entered values (parallel to varNames)
├── varCursor     int
├── varInput      textinput.Model
├── loading       bool                // true during initial scan
├── errorMsg      string
├── lastKey       string              // chord state
├── cmdInput      textinput.Model
├── searchInput   textinput.Model
├── showHelp      bool
├── statusMsg     string
├── width, height int
```

### Update Dispatch

```
Update(msg)
├── tea.WindowSizeMsg    → update dimensions
├── docsLoadedMsg        → populate docs, load templates, clear loading
├── todosLoadedMsg       → populate todos (open tasks only)
├── gitStatusMsg         → store status, rebuild gitFileMap
├── searchDoneMsg        → populate searchResults, set searchActive
├── exportDoneMsg        → update statusMsg
├── sweepDoneMsg         → update statusMsg, re-run cmdLoadTodos
├── fileCreatedMsg       → update statusMsg, re-run cmdLoadDocs
└── tea.KeyMsg
    ├── ModeNormal          → updateNormal
    ├── ModeCommand         → updateCommand → executeCommand
    ├── ModeSearch          → updateSearch → cmdSearch
    ├── ModeTemplatePicker  → updateTemplatePicker
    └── ModeTemplateVars    → updateTemplateVars → cmdCreateDoc
```

### View Filtering (`visibleDocs()`)

The file panel displays different document sets depending on `activeView` and `searchActive`:

| State / View | Documents shown |
|---|---|
| `searchActive = true` | `searchResults` |
| `ViewAll` | all `docs` |
| `ViewJournal` | `docs` filtered by `type == "journal"` |
| `ViewKB` | `docs` filtered by `type == "kb"` |
| `ViewRecent` | top 20 `docs` sorted by `Modified` descending |
| `ViewTodo` | focuses `PanelTodo` (no file list change) |

### Git File Map

On receipt of `gitStatusMsg`, `buildGitFileMap()` indexes `RepoStatus.Files` by absolute path for O(1) lookup during rendering:

```go
func (m *Model) buildGitFileMap() {
    m.gitFileMap = make(map[string]git.FileStatus, len(m.gitStatus.Files))
    for _, f := range m.gitStatus.Files {
        absPath := filepath.Join(m.cfg.BaseDir, f.Path)
        m.gitFileMap[absPath] = f
    }
}
```

---

## Features Wired

### File Panel

Real document names from `search.ListAll(cfg.BaseDir)`. Each entry includes a git status marker when the file appears in `git status --porcelain` output:

```
> 2026-03-14.md    [M]
  2026-03-13.md
  setup-nginx.md   [?]
  deploy-runbook.md
```

Markers: `[M]` = modified, `[?]` = untracked, `[A]` = staged.

Loading state shows `scanning…` before the first `docsLoadedMsg` arrives. Empty directories show `(no documents)`.

### Status Bar

```
 NORMAL │ main ↑2 │ 24 docs              /search  :cmd  ?help  q:quit
```

- Mode indicator on the left
- Branch name from `git.RepoStatus.Branch`
- `↑N` if ahead of remote, `↓N` if behind
- Document count from `len(m.docs)`
- `scanning…` replaces doc count during initial load
- Current status message replaces the right-side hints when set

### View Switching (1–5)

| Key | View |
|---|---|
| `1` | All documents |
| `2` | Journal entries only |
| `3` | Knowledge base only |
| `4` | TODO panel focus |
| `5` | 20 most recently modified |

Each switch resets the file cursor to 0, clears any active search, and immediately applies `visibleDocs()` filtering — no re-scan required since `docs` is already in memory.

### Search (`/`)

Entering a query and pressing `Enter` dispatches `cmdSearch`, which calls `search.Search()` on a goroutine. On return, `searchDoneMsg` populates `searchResults` and sets `searchActive = true`. The file panel immediately shows only matching documents. `Esc` in search mode also clears `searchActive`, restoring the full list.

### Export (`e`)

Pressing `e` on a selected document dispatches `cmdExport(selectedPath, cfg.ExportDir)`, which calls `export.ToFile()`. The result is shown in the status bar: `exported → /path/to/file.md` or an error message.

### Rescan (`R`)

Sets `loading = true` and dispatches `cmdLoadDocs` + `cmdLoadGitStatus` in a batch. The file panel immediately shows `scanning…` until results arrive.

### Command Mode (`:`)

| Command | Action |
|---|---|
| `:q` / `:quit` | `tea.Quit` |
| `:todo sweep` | `cmdSweep(cfg.JournalDir(), cfg.TodoPath())` |
| `:todo archive` | `cmdArchive(todoPath, archivePath, days)` |
| anything else | `":<cmd> — unknown command"` in status bar |

After a sweep or archive, `cmdLoadTodos` is automatically dispatched to refresh the TODO panel.

### Template Picker (`n`)

Pressing `n` in Normal mode enters `ModeTemplatePicker`. Available templates are loaded from `cfg.TemplatesDir()`; if none are found, the five built-in templates are used as fallback.

```
  New Document — Select Template

  > journal
    kb
    howto
    meeting
    scratch

  j/k to navigate, Enter to select, Esc to cancel
```

Selecting a template checks for unresolved `{{variables}}` (excluding builtins handled by `template.Render`). If none remain, the document is created immediately. If variables exist, `ModeTemplateVars` prompts for each one in sequence:

```
  New howto — Enter Details

  title:        Setup Nginx
  author:       |
  tags:

  Enter to confirm, Esc to go back
```

On completion, `cmdCreateDoc` renders the template, determines the destination directory by document type, writes the file, and dispatches a `docsLoadedMsg` refresh.

---

## Preview Panel

Replaced the Phase 2 placeholder text with real frontmatter fields from the selected document:

```
▶ Preview ────────────────────────────
  Setup Nginx

  type: kb
  tags: devops, nginx
  modified: 2026-03-10

  setup-nginx.md
```

---

## Config Threading

`tui.Run()` now takes `config.Config` as its sole argument. `internal/cli/root.go` passes the already-loaded `cfg` variable:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    return tui.Run(cfg)
},
```

The `Model` stores `cfg` and uses it in every command factory. No global state. No re-loading of config inside the TUI.

---

## Test Coverage

54 test functions in `tui/model_test.go` (27 Phase 2, 27 Phase 3).

### Phase 3 tests

| Test | What it verifies |
|---|---|
| `TestDocsLoaded` | `docsLoadedMsg` populates `docs`, clears `loading` |
| `TestDocsLoadedError` | Error in `docsLoadedMsg` sets `errorMsg` |
| `TestTodosLoaded` | `todosLoadedMsg` populates `todos` |
| `TestGitStatusLoaded` | `gitStatusMsg` stores branch and ahead count |
| `TestSearchResults` | `searchDoneMsg` sets `searchActive` and `searchResults` |
| `TestVisibleDocsAll` | `ViewAll` returns all docs |
| `TestVisibleDocsJournal` | `ViewJournal` returns only type=journal |
| `TestVisibleDocsKB` | `ViewKB` returns only type=kb |
| `TestVisibleDocsRecent` | `ViewRecent` is sorted by `Modified` descending |
| `TestVisibleDocsSearch` | `searchActive=true` returns `searchResults` |
| `TestGitMarkerModified` | `Y='M'` → `[M]` |
| `TestGitMarkerUntracked` | `X='?', Y='?'` → `[?]` |
| `TestGitMarkerStaged` | `X='A'` → `[A]` |
| `TestGitMarkerNone` | Clean file → empty string |
| `TestCommandQuit` | `:q` returns `tea.Quit` |
| `TestCommandTodoSweep` | `:todo sweep` returns a command |
| `TestCommandUnknown` | Unknown command sets status with command name |
| `TestTemplatePickerNavigation` | j/k moves `pickerCursor` |
| `TestTemplatePickerEscape` | Esc returns to `ModeNormal` |
| `TestStatusBarShowsBranch` | Branch name appears in rendered view |
| `TestStatusBarShowsDocCount` | Doc count appears in rendered view |
| `TestStatusBarLoadingState` | `scanning…` appears while `loading=true` |
| `TestViewSwitching` | Keys 1–3, 5 set correct `activeView` |
| `TestView4SwitchesToTodoPanel` | Key 4 sets `activeView=ViewTodo`, `activePanel=PanelTodo` |
| `TestRescanSetsLoading` | `R` sets `loading=true` and returns a command |
| `TestToKebabCase` | Title-to-filename conversion |
| `TestExportNoDocSelected` | `e` with empty list sets appropriate status message |

### Updated Phase 2 tests

- `TestNewModel` — checks `loading=true` instead of non-empty `fileItems`
- `TestInit` — now asserts that `Init()` returns a non-nil command (Phase 3 changed this)
- `TestModeString` — extended with `ModeTemplatePicker` and `ModeTemplateVars`
- Cursor and jump tests — use `modelWithDocs()` helper to inject fake data first

---

## Phase 3 Checklist (from Phase 2 report)

- [x] `Init()` — returns a `tea.Batch` scanning `cfg.BaseDir`, loading TODOs, loading git status
- [x] Replace `fileItems []string` with `[]search.Result` from `search.ListAll()`
- [x] Replace `todoItems []string` with `[]todo.Task` from `todo.ReadTasks()`
- [x] Wire `1`–`5` number keys to view switching
- [x] Wire `/` search to `search.Search()` and display results in the file panel
- [x] Wire `:q`, `:todo sweep`, `:todo archive` in command mode
- [x] Implement git status bar (branch, ahead/behind, doc count)
- [x] Show per-file git status indicators (`[M]`, `[?]`, `[A]`)
- [x] Wire `R` rescan to re-run scan commands
- [x] Wire `e` export to `export.ToFile()` on selected document
- [x] Wire `n` new document to template picker flow
- [x] Show real document count in status bar

---

## Phase 4 Handoff

The following remain as stubs, intentionally deferred to Phase 4:

- **`Enter`** — editor handoff via `tea.ExecProcess($EDITOR, selectedDoc)`
- **`s`** — scratch pad (open `cfg.ScratchPath()` in `$EDITOR`)
- **`g`** — lazygit handoff via `tea.ExecProcess("lazygit", cfg.BaseDir)`
- **`d`** — delete selected document (with confirmation prompt)

Phase 4's scope is the `tea.ExecProcess` suspend/resume pattern — the single most failure-prone interaction in the application.
