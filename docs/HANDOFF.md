# mdam — Project Handoff

mdam is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, scratch notes.

**Current branch:** `feature/search-overhaul`

---

## Current State

All core features are implemented and working. The application is in active daily use (dogfooding).

### What changed this session
- **Search pane (View 5)** — new dedicated pane replacing the old search mode. Two-column layout: left panel has search input + three category entries (Journal, KB, Tags), right panel shows documents for the selected category. Enter activates the input, Escape clears results.
- **Tag browser filter** — substring filter input on the Tag Browser left panel. Enter activates, keystrokes narrow the list, Enter again deactivates (keeps filtered list), Escape clears.
- **`g` keybinding simplified** — `gg` chord replaced with single `g` for jump-to-top in all contexts. `lastKey` chord state removed.
- **Help menu overhaul** — two-column layout using full viewport width. Added Read Mode keybindings, Git Markers legend, updated all bindings to match current state.
- **Documentation overhaul** — removed roadmap, future promises, and references to removed features. All docs now reflect current reality.

### Features on ice
- **Full TODO system** — sweep, archive, categories, priorities. `internal/todo` package exists but is not wired to the TUI.
- **Import/inbox pipeline** — `internal/importer` exists but `.inbox/` not scaffolded. CLI hidden.
- **Delete feature** — `d` key disabled, delete confirmation flow removed.

### Five TUI panes
1. Dashboard — journal / pinned / recent + todo.md preview
2. Journal — month-folder tree
3. KB — subtype-folder tree
4. Tag Browser — tag list with substring filter + documents per tag
5. Search — query input + results categorized by Journal/KB/Tags

### Git integration
- `internal/git/git.go` — branch, ahead/behind, per-file status via `git` binary
- Loaded async on: startup, editor return, manual `R`
- Per-file markers in all views (Journal, KB, Tags, Dashboard, Search)

### Config
- Fields: editor, author, base_dir, export_dir, theme, nerd_fonts, journal.auto_create

## All tests pass
```
go test ./... — PASS
go vet ./...  — clean
```
