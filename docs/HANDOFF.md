# mdam — Project Handoff

mdam (Markdown Admin Management) is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, scratch notes.

**Current branch:** `refactor/simplify-setup-and-ice-features`

---

## Current State

Major simplification to enable dogfooding, plus dashboard UX fixes and read mode improvements.

### Features put on ice (backlogged)
- **Full TODO system** — sweep, archive, categories, priorities. `internal/todo` package exists but is not wired to the TUI.
- **Import/inbox pipeline** — `internal/importer` exists but `.inbox/` not scaffolded. CLI hidden.
- **Lazygit integration** — removed from active keybindings.
- **Git config** — removed from user-facing config. Git status detection still works.
- **Delete feature** — `d` key disabled, `ModeDeleteConfirm` removed, `delete.go` removed.
- **Search** — needs overhaul, flagged as next task.

### Directory structure
- Scaffolded: `journal/`, `kb/`, `.templates/`
- Singletons at base root: `todo.md`, `scratch.md`

### Templates
- Two built-in: `journal.md`, `kb.md`

### Config
- Fields: editor, author, base_dir, export_dir, theme, nerd_fonts, journal.auto_create
- No todo, git, or import sections

### TUI Setup Wizard
- `tui/wizard.go` — separate BubbleTea program for first-run
- Steps: base_dir, editor, author, theme (live preview), nerd_fonts, export_dir, config preview

### Dashboard (key 1)
- Left: Journal (5 recent) / Pinned (up to 10, insertion order) / Recent (up to 10, excludes journal/todo/scratch)
- Right: glamour-rendered `todo.md` preview
- Enter/o on right panel opens todo.md in editor/read mode
- `p` correctly pins from dashboard using `dashCursor`
- `selectedDoc()` and `selectedDocTitle()` handle ViewDashboard

### Pins
- Ordered list in `pins.json` (insertion order, not sorted)
- FIFO eviction at max 10 pins
- `pinnedOrder []string` + `pinnedPaths map[string]bool` on Model

### Read mode (key o)
- Full-width glamour rendering (uses terminal width, not hardcoded 80)
- Navigation: j/k (line), d/u (half-page), f/b (page), g/G (top/bottom)

### Journal auto-create
- `journal.auto_create: true` (default) creates today's entry on TUI startup
- Triggers doc re-scan if new entry was created

## Next task: Search overhaul

The search feature (`/`) needs a UX redesign. This was flagged by the user as the remaining sizable issue before the app is ready for regular use.

## All tests pass
```
go test ./... — PASS
go vet ./...  — clean
```
