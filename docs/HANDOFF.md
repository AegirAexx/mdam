# mdam — Project Handoff

mdam is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, scratch notes.

**Current branch:** `fix/document-pins`

---

## Current State

All core features are implemented and working. The application is in active daily use (dogfooding).

### What changed this session

**Pin path portability fix** — `pins.json` now stores paths relative to `base_dir` instead of absolute paths. This fixes pins being silently wiped when the document repo is synced across machines with different home directories.

- `tui/pins.go` — `loadPins` and `savePins` accept a `baseDir` parameter. Save converts absolute → relative via `filepath.Rel`. Load converts relative → absolute via `filepath.Join`. Absolute entries pass through unchanged (backward compat with pre-fix `pins.json` files).
- `tui/commands.go` — `cmdLoadPins` / `cmdSavePins` thread `baseDir` through.
- `tui/model.go` — both call sites pass `m.cfg.BaseDir`.
- `tui/pins_test.go` — existing tests updated; two new tests: `TestSaveAndLoadPinsRelativePaths` (on-disk JSON is relative, round-trip is correct) and `TestLoadPinsAbsolutePathsBackwardCompat` (old absolute-path files load correctly).

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
- Pins at `{base_dir}/.mdam/pins.json` (stored as relative paths as of this session)

## All tests pass

```
go test ./... — PASS
go vet ./...  — clean
```
