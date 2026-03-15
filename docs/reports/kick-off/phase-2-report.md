# Phase 2 Report — TUI Skeleton

**Completed:** 2026-03-14
**Goal:** Establish the BubbleTea framework and event loop with correct architecture. A running TUI that responds to keybindings and displays placeholder content. Architecture is correct even though content is fake.

---

## Summary

Phase 2 is complete. The TUI skeleton is running. `mdam` with no subcommand now launches a full-screen interactive terminal application. All core navigation keybindings work. Command and search mode input is functional. The architecture is wired for Phase 3 to slot real data in with minimal changes.

**27 new tests. 96 total passing. `go vet` clean.**

---

## What Was Built

### `tui/` Package (1,125 lines across 6 files)

| File | Lines | Purpose |
|---|---|---|
| `mode.go` | 60 | `Mode` and `PanelID` types, string representations, cycle helpers |
| `keys.go` | 81 | `KeyMap` struct with all keybindings and `DefaultKeyMap()` |
| `model.go` | 338 | `Model` struct, `Init`/`Update`, per-mode handlers, cursor helpers |
| `view.go` | 236 | `View()`, panel rendering, status bar, help overlay, layout helpers |
| `tui.go` | 44 | `Run()` entry point |
| `model_test.go` | 366 | 27 test functions |

### Entry Point

`mdam` (invoked with no subcommand) now launches the interactive TUI via `tui.Run()`, which creates a `tea.Program` with `tea.WithAltScreen()` (uses the alternate screen buffer, restoring the terminal on exit) and mouse cell motion enabled for Phase 5.

---

## Architecture

### MVU Model

The `Model` struct owns all application state for the current render cycle. No state lives outside it.

```
Model
├── mode         Mode          // Normal | Command | Search
├── activePanel  PanelID       // Files | Preview | Todo
├── keys         KeyMap        // keybinding definitions
├── fileItems    []string      // dummy file list (Phase 3: real scan results)
├── fileCursor   int
├── todoItems    []string      // dummy todo list (Phase 3: real tasks)
├── todoCursor   int
├── lastKey      string        // chord state for gg detection
├── cmdInput     textinput     // bubbles/textinput for command mode
├── searchInput  textinput     // bubbles/textinput for search mode
├── showHelp     bool
├── statusMsg    string
├── width, height int          // updated by tea.WindowSizeMsg
```

### Update Dispatch

`Update` switches on message type, then dispatches key events to a per-mode handler:

```
Update(msg)
├── tea.WindowSizeMsg → update dimensions
└── tea.KeyMsg
    ├── ModeNormal  → updateNormal
    ├── ModeCommand → updateCommand (delegates to textinput)
    └── ModeSearch  → updateSearch  (delegates to textinput)
```

Each handler returns `(tea.Model, tea.Cmd)` — pure function, no side effects.

### Panel System

Three panels in cycle order: `PanelFiles → PanelPreview → PanelTodo → (wrap)`.

`PanelID` has `next()` and `prev()` methods. `Tab` / `Shift+Tab` and `h`/`l` advance the cycle. Each panel tracks its own cursor independently.

### Chord Handling

The `gg` chord (vim: jump to top) is implemented via `lastKey` state. If `g` is received and `lastKey == "g"`, the chord fires and `lastKey` is cleared. Any other key resets `lastKey`. This is the correct approach for arbitrary multi-key sequences.

---

## Keybindings Implemented

### Normal Mode

| Key | Action |
|---|---|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `h` / `←` | Previous panel |
| `l` / `→` | Next panel |
| `gg` | Jump to top (chord) |
| `G` | Jump to bottom |
| `Tab` | Next panel (cycle) |
| `Shift+Tab` | Previous panel (cycle) |
| `/` | Enter Search mode |
| `:` | Enter Command mode |
| `?` | Toggle help overlay |
| `1`–`5` | View switching (status stub, Phase 3) |
| `Enter` | Open in `$EDITOR` (stub, Phase 4) |
| `g` | Lazygit (stub, Phase 4) |
| `n` | New document (stub, Phase 3) |
| `s` | Scratch pad (stub, Phase 4) |
| `e` | Export (stub, Phase 3) |
| `d` | Delete (stub, Phase 3) |
| `R` | Rescan (stub, Phase 3) |
| `q` | Quit |

### Command Mode (`:`)

- `Esc` — cancel, return to Normal
- `Enter` — execute command, return to Normal
- All other keys delegated to `bubbles/textinput`
- `:q` / `:quit` returns an informational status (full wire-up in Phase 3)

### Search Mode (`/`)

- `Esc` — cancel, return to Normal
- `Enter` — submit query, return to Normal (real search in Phase 3)
- All other keys delegated to `bubbles/textinput`

---

## View Layout

```
▶ Files ─────────────────────│─ Preview ────────────────────────────
> 2026-03-14.md              │
  2026-03-13.md              │  # 2026-03-14.md
  2026-03-12.md              │
  setup-nginx.md             │  (document preview — Phase 3)
  deploy-runbook.md          │
  meeting-notes-…            │  Open in $EDITOR: press Enter (Phase 4)
  scratch.md                 │
                             │─ TODOs ──────────────────────────────
                             │> [ ] Review PR #42 @work
                             │  [ ] Update DNS records @work
                             │  [ ] Buy groceries @personal
────────────────────────────────────────────────────────────────────
 NORMAL                                    /search  :cmd  ?help  q:quit
```

Phase 2 uses plain text with box-drawing characters. No lipgloss styling — that is Phase 5 work. The layout proportions (left ~33%, right ~67%) and panel structure are the same ones Phase 5 will style.

---

## Test Coverage

27 test functions in `tui/model_test.go`.

| Test | What it verifies |
|---|---|
| `TestNewModel` | Initial state: Normal mode, PanelFiles, non-empty items |
| `TestModeString` | String representations of all three modes |
| `TestPanelCycle` | `next()`/`prev()` wrapping, full cycle returns to start |
| `TestCursorMovement` | j/k move cursor, clamps at 0 |
| `TestCursorMovementArrowKeys` | Arrow keys work identically to j/k |
| `TestJumpBottom` | G moves cursor to last item |
| `TestGGChord` | Two-key gg chord jumps to top; single g does not |
| `TestTabCyclesPanels` | Tab advances through all three panels and wraps |
| `TestShiftTabCyclesPanelsReverse` | Shift+Tab goes to previous panel |
| `TestEnterSearchMode` | `/` sets mode to ModeSearch |
| `TestEnterCommandMode` | `:` sets mode to ModeCommand |
| `TestEscapeFromSearch` | Esc returns to ModeNormal from search |
| `TestEscapeFromCommand` | Esc returns to ModeNormal from command |
| `TestHelpToggle` | `?` toggles showHelp; second `?` closes it |
| `TestWindowResize` | `tea.WindowSizeMsg` updates width/height |
| `TestTodoCursorMovement` | Cursor movement on PanelTodo after `4` key |
| `TestViewRendersWithoutPanic` | `View()` returns non-empty string |
| `TestViewContainsModeIndicator` | View shows "NORMAL" in status bar |
| `TestViewHelpOverlay` | Help overlay contains "Keybindings" |
| `TestViewInSearchMode` | View shows "SEARCH" after `/` |
| `TestViewInCommandMode` | View shows "COMMAND" after `:` |
| `TestSearchEnterReturnsToNormal` | Enter from search mode → Normal |
| `TestCommandEnterReturnsToNormal` | Enter from command mode → Normal |
| `TestInit` | `Init()` returns nil (no startup I/O in Phase 2) |
| `TestPadRight` | Padding and truncation to exact width |
| `TestTruncate` | String truncation with `…` suffix |
| `TestPanelHeader` | Focused header uses `▶`, unfocused uses `─` |

---

## Dependencies Added

| Package | Purpose |
|---|---|
| `charmbracelet/bubbletea` v1.3.10 | MVU event loop, `tea.Program`, `tea.Msg`, `tea.Cmd` |
| `charmbracelet/bubbles` v1.0.0 | `textinput` component for command/search input |
| `charmbracelet/lipgloss` v1.1.0 | Pulled in transitively by bubbles (unused until Phase 5) |

---

## Phase 3 Integration Checklist

The following changes will connect the TUI skeleton to the real engine:

- [ ] `Init()` — return a `tea.Cmd` that scans `cfg.BaseDir` and returns a results message
- [ ] Replace `fileItems []string` with `[]search.Result` — feed from `search.ListAll()`
- [ ] Replace `todoItems []string` with `[]todo.Task` — feed from `todo.ReadTasks()`
- [ ] Wire `1`–`5` number keys to actual view switching (journal list, KB list, etc.)
- [ ] Wire `/` search to `search.Search()` and display results in the file panel
- [ ] Wire `:import`, `:todo sweep`, `:config`, `:q` command mode commands
- [ ] Implement git status bar (branch, ahead/behind, change count) via `git.Status()`
- [ ] Show per-file git status indicators (`[M]`, `[?]`) in the file list
- [ ] Wire `R` rescan to re-run `Init()` scan command
- [ ] Wire `e` export to `export.ToFile()` on selected document
- [ ] Wire `n` new document to template picker flow
- [ ] Show real document count in status bar
