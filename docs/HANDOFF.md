# mdam — Project Handoff

mdam is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, scratch notes.

**Current branch:** `test/watcher-audit`

---

## Current State

All core features are implemented and working. The application is in active daily use (dogfooding). This branch is a test quality pass addressing all 15 actions from the watcher audit (2026-04-08).

### What changed this session

**Test quality audit** — 15 actions across 9 packages, all resolved:

- **ACTION-001** `internal/document` — `Write()` round-trip now verifies Body, Type, Tags, and timestamps, not just Title.
- **ACTION-002** `internal/document` — `RenderFrontmatter()` round-trip test for `Extra` fields.
- **ACTION-003** `internal/search` — Direct table-driven tests for `scoreDocument()` and `extractSnippet()`.
- **ACTION-004** `internal/setup` — `TestEnsureGettingStarted` added (idempotency verified).
- **ACTION-005** `internal/git` — Tautological `TestIsAvailable` fixed; result now stored and used.
- **ACTION-006** `internal/todo` — `ReadTasks()` happy-path test with mixed-status tasks.
- **ACTION-007** `internal/todo` — `ParseTasks()` assertions extended to cover `.Text` field.
- **ACTION-008** `internal/journal` — `ListByMonth()` now asserts filenames and newest-first order.
- **ACTION-009** `internal/journal` — `ScaffoldFrontmatter()` now asserts Created/Modified timestamps and empty Tags slice.
- **ACTION-010** `internal/config` — `TestDefaultConfigPath` and `TestLoadFromInvalidYAML` added.
- **ACTION-011** `internal/cli` — Functional tests: `TestJournalListCommand`, `TestTodoListCommand` execute real subcommands.
- **ACTION-012** `internal/importer` — `scaffoldFrontmatter()` now asserts title derivation and non-zero timestamps.
- **ACTION-013** `tui` — `TestBuildTagIndexTieBreak` verifies alphabetical sort when tag counts are equal.
- **ACTION-014** `internal/document` — `TestParseFrontmatterTimestamps` verifies year/month/day values.
- **ACTION-015** `internal/search` — `TestSearchWithBody` snippet now asserts content contains query term.

**Coverage delta:** 64.8% → 65.5% overall. Notable package gains: `cli` 11.3%→25.7%, `config` 75.6%→84.4%, `setup` 73.1%→79.0%, `search` 88.1%→91.1%.

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
- Pins at `{base_dir}/.mdam/pins.json`

## All tests pass

```
go test ./... — PASS
go vet ./...  — clean
```
