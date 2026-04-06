# mdam — Project Handoff

mdam (Markdown Admin Management) is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, scratch notes.

**Current branch:** `fix/dogfooding-git-markers-root-ignore`

---

## Current State

Dogfooding fixes: git markers wired to all views, nerd font icons restored, root-level repo files ignored, dashboard UX fixes, recent list uses filesystem mtime.

### What changed this session
- **Git markers in all views** — `[M]`/`[?]`/`[A]` (or nerd font equivalents) now show in Journal, KB, Tags, and Dashboard views. Previously only wired to an unused flat file list.
- **Nerd font icons restored** — all `DefaultIcons()` glyphs were `U+0020` (space). Replaced with correct codepoints using `\U` escape sequences that survive any editor/encoding.
- **Pin markers use icon system** — hardcoded `[*]` replaced with `m.icons.Pinned` across all views. Dashboard section header also uses the icon.
- **Root-level repo files ignored** — `README.md`, `LICENSE.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md` silently skipped when in base_dir root. Prevents "skipped — invalid frontmatter" noise.
- **Recent list uses filesystem mtime** — `recentDocs()` now sorts by `Result.ModTime` (from `os.Stat`) instead of `Frontmatter.Modified`. Editing a file immediately bumps it in the list.
- **Dashboard right panel j/k disabled** — cursor no longer moves a hidden global file list when focus is on the todo panel.
- **Status bar singleton indicators** — scratch count removed. Red warning icon shown when `todo.md` or `scratch.md` is missing.
- **Roadmap cleaned** — lazygit integration and git auto-commit permanently removed from README roadmap and HANDOFF on-ice list.

### Features put on ice (backlogged)
- **Full TODO system** — sweep, archive, categories, priorities. `internal/todo` package exists but is not wired to the TUI.
- **Import/inbox pipeline** — `internal/importer` exists but `.inbox/` not scaffolded. CLI hidden.
- **Git config** — removed from user-facing config. Git status detection still works.
- **Delete feature** — `d` key disabled, `ModeDeleteConfirm` removed, `delete.go` removed.
- **Search** — needs overhaul, flagged as next task.

### Git status integration
- `internal/git/git.go` — branch, ahead/behind, per-file status, stash count via `git` binary
- Loaded async on: startup, editor return, manual `R` key
- Status bar: branch name + `↑N`/`↓N` sync indicators
- Per-file: styled markers in all tree views (Journal, KB, Tags, Dashboard)
- No periodic polling, no `git fetch` — intentional

### Nerd font icons
- All codepoints in `tui/icons.go` use `\U` escape sequences (e.g., `\U000F0415`)
- Includes: document types, git status, pin, missing, cursor, dashboard, tag, filter

### Directory structure
- Scaffolded: `journal/`, `kb/`, `.templates/`
- Singletons at base root: `todo.md`, `scratch.md`

### Config
- Fields: editor, author, base_dir, export_dir, theme, nerd_fonts, journal.auto_create
- No todo, git, or import sections

## Next task: Search overhaul

The search feature (`/`) needs a UX redesign. This was flagged by the user as the remaining sizable issue before the app is ready for regular use.

## All tests pass
```
go test ./... — PASS
go vet ./...  — clean
```
