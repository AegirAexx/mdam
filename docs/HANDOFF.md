# mdam — Project Handoff

mdam is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, scratch notes.

**Current branch:** `fix/kb-tree-index-panic`

---

## Current State

All core features are implemented and working. The application is in active daily use (dogfooding). This branch fixes a critical KB tree panic found during dogfooding.

### What changed this session

- **KB tree index-out-of-bounds panic fix** — Expanding a folder below a large already-expanded folder caused `kbCursor` to go stale (the expand handler collapses all other folders but never relocated the cursor). The next `h` keypress accessed `rows[i]` out of bounds and panicked. Fixed both the `l` handler (rebuild rows and relocate cursor after expand) and the `h` handler (clamp cursor before any slice access). Two new tests in `tui/view_kb_test.go`.

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
- Pins now at `{base_dir}/.mdam/pins.json` (was `~/.config/mdam/pins.json`)

## All tests pass

```
go test ./... — PASS
go vet ./...  — clean
```
