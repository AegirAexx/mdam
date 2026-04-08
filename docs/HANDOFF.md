# mdam — Project Handoff

mdam is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, scratch notes.

**Current branch:** `fix/dogfooding-round-2`

---

## Current State

All core features are implemented and working. The application is in active daily use (dogfooding). Second round of dogfooding fixes landed on this branch.

### What changed this session

- **Template overwriting fix** — `WriteBuiltins()` now skips any template file that already exists on disk, never overwriting user customizations.
- **New default journal template** — Hours, Work Done, Daily Log/Notes, TODO sections.
- **KB docs saved to correct directory** — `cmdCreateDoc` now uses `tmpl.TemplateType(t.Content)` to extract the document type from template frontmatter instead of `vars["type"]` (which was always empty). KB docs now correctly go to `kb/`.
- **Read mode off-by-one fix** — Viewport height changed from `height-3` to `height-2`. Also added resize handling for the read viewport on terminal resize.
- **pins.json moved to `{base_dir}/.mdam/`** — Lives in version control now. `PinsPath()` returns `{BaseDir}/.mdam/pins.json`. Stale pins (deleted files) are auto-pruned on load.
- **Getting-started KB doc** — `EnsureGettingStarted()` seeds `kb/getting-started-with-mdam.md` on first run. Covers onboarding, frontmatter, directory layout, KB subtypes, templates, keybinding quick reference.
- **Docs updated** — README, FRONTMATTER.md updated for `.mdam/` dir, pins location, KB subtypes, getting-started doc.

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
