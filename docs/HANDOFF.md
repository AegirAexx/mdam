# mdam — Project Handoff

> Complete project state as of Phase 5 + Issues #1, #4, #3, #2, #5 fixed. All five phases are implemented and passing. No features have been individually tested or verified yet. This document is the single source of truth for any future session or collaborator picking up the project.

---

## 1. What mdam Is

mdam (Markdown Admin Management) is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base documents, TODOs, and scratch notes. It is an **administration and routing tool**, not an editor. All text editing is delegated to `$EDITOR`. The filesystem is the database; there is no SQL layer, no cache, no sync service. Git provides version control and multi-device sync via a lazygit handoff (`ctrl+g`).

---

## 2. Build, Test, Run

```bash
go test ./...                  # run all tests (all 12 packages must pass)
go vet ./...                   # static analysis (must be clean)
go build -o mdam ./cmd/mdam    # build the binary
./mdam                         # launch the TUI — first run prompts for base dir and scaffolds folders
```

These three commands are the mandatory gate after every change. Never commit with a failing test or vet warning.

---

## 3. Complete Feature Inventory

> **Note:** All features below are implemented but ⚠️ untested. None have been individually verified as working correctly.

### Engine (Phases 1–3, headless, all CLI-accessible)

| Feature | How to invoke |
|---|---|
| Config loading | Auto on startup; `mdam config` / `mdam config --edit` |
| Directory scanning | Runs on every operation — no cache |
| Frontmatter parsing & validation | `internal/document` — validates all 5 required fields |
| Filename validation (kebab-case) | Part of import pipeline and document creation |
| Import pipeline | `mdam import <path> [--auto-fix] [--dry-run]` |
| Daily journal creation | `mdam journal create [date]` |
| Journal listing | `mdam journal list [--month YYYY-MM]` |
| TODO parsing | `mdam todo list [--status S] [--category C] [--all]` |
| TODO sweep | `mdam todo sweep` or `:todo sweep` in TUI |
| TODO archive | `mdam todo archive [--older-than N]` |
| Template discovery | `mdam template list` / `mdam template show <name>` |
| Template rendering | Variable interpolation: `{{date_short}}`, `{{title}}`, `{{author}}`, etc. Caller-supplied vars take precedence over built-ins. |
| Fuzzy search | `mdam search "query" [--tag T] [--type T] [--modified-after D]` |
| Export (strip frontmatter) | `mdam export <file> [--to DIR]` |
| Git status detection | `mdam status [--porcelain]` |
| Scratch pad | `mdam scratch` |
| New document | `mdam new [--template T] [--title "T"]` |

### TUI (Phases 2–5)

| Feature | Key / Activation |
|---|---|
| Launch TUI | `mdam` (no subcommand) |
| Navigate list | `j`/`k`, `gg`/`G` |
| Panel focus | `Tab` / `Shift+Tab`, `h`/`l` |
| Dashboard (today's context) | `1` |
| Journal view | `2` |
| Knowledge base view | `3` |
| TODO view | `4` |
| Recent documents view | `5` |
| Tag browser | `6` |
| ViewAll (all docs, startup default) | No key — active when no number view selected |
| Fuzzy search | `/` |
| Command mode | `:` |
| Keybinding help overlay | `?` |
| Open in `$EDITOR` | `Enter` |
| Open scratch pad in `$EDITOR` | `s` |
| New document (template picker) | `n` |
| Delete with confirmation | `d` → `y` (confirm) / `n` or `Esc` (cancel) |
| Export selected document | `e` |
| Pin / unpin selected document | `p` |
| Cycle smart filter | `f` — None → Untagged → StaleWeek → Inbox → None |
| Force re-scan filesystem | `R` |
| Open lazygit | `ctrl+g` |
| Quit | `q` |
| Glamour markdown preview | Automatic in right viewport panel |
| Color theme | Set `theme:` in config.yml |
| Nerd Font icons | Set `nerd_fonts: true` in config.yml |
| Spinner (loading indicator) | Automatic while async commands run |

### Phase 5 — Ambient Findability Views

| View | Key | What it shows |
|---|---|---|
| Dashboard | `1` | Today's journal entry, open TODO count + top items, pinned docs, recently modified |
| Tag browser | `6` | All tags sorted by document count; select a tag to list its documents |
| Smart filter | `f` | Untagged docs / Stale (not modified in 7d) / Inbox (`type: unsorted`) |
| Pin/unpin | `p` | Bookmarks persist to `~/.config/mdam/pins.json` |

---

## 4. Architecture Map

| Package | Responsibility |
|---|---|
| `cmd/mdam/` | Binary entrypoint, delegates to `internal/cli` |
| `internal/cli/` | Cobra subcommand wiring — no business logic |
| `internal/config/` | Viper config loading from `~/.config/mdam/config.yml`; corrected path helpers |
| `internal/setup/` | First-run detection, config scaffolding, directory creation, template seeding |
| `internal/document/` | `Document` model, frontmatter parsing/validation, kebab-case check |
| `internal/importer/` | Import pipeline: validate, auto-fix, duplicate detection |
| `internal/journal/` | Daily journal creation from template, listing, date parsing |
| `internal/todo/` | Task parsing, sweep logic, archive, status/category filter |
| `internal/template/` | Template discovery, render with variable interpolation, `TemplateType()` extracts type field from frontmatter |
| `internal/search/` | Fuzzy search over frontmatter + optional body content |
| `internal/export/` | Frontmatter stripping for sharing/clipboard |
| `internal/git/` | Shells out to `git status --porcelain` and `git rev-list` |
| `tui/mode.go` | `Mode`, `PanelID`, `View` type definitions |
| `tui/keys.go` | `KeyMap` and `DefaultKeyMap()` |
| `tui/messages.go` | All async message types (`docsLoadedMsg`, `previewReadyMsg`, `tickMsg`, etc.) |
| `tui/commands.go` | `tea.Cmd` factories wrapping engine calls; `cmdJournalCreate` delegates to `journal.Create` and returns `scratchReadyMsg` |
| `tui/model.go` | `Model` struct, `Init`/`Update`, all mode handlers; `pickerTemplates` holds filtered (journal+kb) picker subset |
| `tui/view.go` | `View()`, panel rendering, status bar |
| `tui/view_dashboard.go` | `renderDashboard()` — dashboard layout |
| `tui/view_tags.go` | `buildTagIndex()`, `renderTagBrowser()`, `renderTagPanel()` |
| `tui/theme.go` | `Theme` struct, `NewTheme(name)`, 5 palettes (lipgloss styles) |
| `tui/icons.go` | `Icons` struct, `DefaultIcons()` (Nerd Font), `PlainIcons()` (ASCII) |
| `tui/pins.go` | `loadPins`, `savePins`, `togglePin` — pure functions, JSON persistence |
| `tui/delete.go` | `cmdDeleteDoc`, `deleteDoneMsg` |
| `tui/tui.go` | `Run(cfg)` entry point |

---

## 5. Configuration Reference

File: `~/.config/mdam/config.yml`

```yaml
# Core
editor: nvim                          # $EDITOR override (falls back to env var)
author: "Your Name"

# Directories
base_dir: ~/notes
export_dir: ~/Downloads
import:
  inbox_dir: ~/notes/.inbox
  auto_fix: false

# Visual
theme: tokyonight                     # nord, gruvbox, catppuccin, dracula
nerd_fonts: false                     # true if terminal font has Nerd Font glyphs

# Git
git:
  enabled: true
  auto_commit: false                  # auto-commit on sweep/journal-create/import
  lazygit: true

# TODO
todo:
  default_category: personal
  archive_after_days: 30

# Journal
journal:
  auto_create: true
  sweep_on_create: true
```

**Phase 5 additions** not in earlier configs:
- `theme` — selects the lipgloss color palette
- `nerd_fonts` — switches between `DefaultIcons()` and `PlainIcons()`

**Computed paths (not in config):**
- `cfg.PinsPath()` → `~/.config/mdam/pins.json`
- `cfg.TemplatesDir()` → `{base_dir}/.templates`
- `cfg.TodoDir()` / `cfg.TodoPath()` → `{base_dir}/todo/todo.md`
- `cfg.ScratchDir()` / `cfg.ScratchPath()` → `{base_dir}/scratch/scratch.md`
- `cfg.ArchivePath()` → `{base_dir}/todo/archive.md`

**First-run behavior:** `base_dir` defaults to `""`. On startup, `setup.IsFirstRun` detects a missing config or empty/absent `base_dir` and runs the guided setup flow (prompts for path, scaffolds 6 dirs, seeds templates, creates scratch pad). Fully idempotent.

---

## 6. TUI Keybinding Quick-Reference

Full reference: `docs/KEYBINDINGS.md`

| Key | Mode | Action |
|---|---|---|
| `j` / `k` | Normal | Down / up |
| `h` / `l` | Normal | Left / right panel |
| `gg` / `G` | Normal | Top / bottom of list |
| `Tab` / `Shift+Tab` | Normal | Cycle panel focus |
| `1`–`6` | Normal | Switch view |
| `Enter` | Normal | Open in `$EDITOR` |
| `s` | Normal | Open scratch pad |
| `n` | Normal | New document (journal or KB picker) |
| `d` | Normal | Delete (enter confirm mode) |
| `y` / `n` / `Esc` | DeleteConfirm | Confirm / cancel delete |
| `e` | Normal | Export |
| `p` | Normal | Pin / unpin |
| `f` | Normal | Cycle smart filter |
| `R` | Normal | Force re-scan |
| `/` | Normal | Enter search mode |
| `:` | Normal | Enter command mode |
| `ctrl+g` | Normal | Open lazygit |
| `?` | Normal | Toggle help overlay |
| `q` | Normal | Quit |
| `Enter` / `Esc` | Search | Confirm / cancel |
| `j` / `k` | Search | Navigate results |
| `Enter` / `Esc` | Command | Execute / cancel |

---

## 7. Known Deferred Items

These items are explicitly out of scope for v1. See spec §9 for full discussion.

| Item | Notes |
|---|---|
| AI / Agent integration | `.ai/` directory, LLM query interface over the managed tree |
| Multi-device conflict resolution | Git merge conflict detection and TUI surfacing |
| Structured TODO format | YAML task blocks if flat `- [ ] text` format proves insufficient |
| File watchers (`fsnotify`) | Replace full re-scan with OS-level change events |
| TODO-specific keybindings | Mark done, change status/category — deferred to real usage |
| `g` vs `ctrl+g` for lazygit | Revisit after Phase 5 real usage data |
| Arrow key navigation | `j`/`k` vs arrow keys for search results |

---

## 8. Report Index

### Phase kick-off reports (`docs/reports/kick-off/`)

| Phase | Report | Key deliverable |
|---|---|---|
| 1 | `phase-1-report.md` | Headless engine, full test suite |
| 2 | `phase-2-report.md` | BubbleTea TUI skeleton |
| 3 | `phase-3-report.md` | Real data wired, git status bar |
| 4 | `phase-4-report.md` | `$EDITOR` + lazygit handoff |
| 5 | `phase-5-report.md` | Theming, glamour preview, ambient findability |

### Issue fix reports (`docs/reports/issues/`)

| Issue | Report | Summary |
|---|---|---|
| #1 | `issue-001-report.md` | First-run initialization, path bug fixes, `internal/setup` package |
| #4 | `issue-002-report.md` | Frontmatter field order, date-only `!!timestamp` emit, docs reorganization |
| #3, #2, #5 | `issue-003-report.md` | Template picker filter, journal creation flow fix, README restructure |

---

## 9. Frontmatter Contract

Every managed document requires these YAML frontmatter fields. **Field order matters** — `RenderFrontmatter` always emits them in this order:

```yaml
---
type: journal
title: 2026-03-15
tags: []
created: 2026-03-15
modified: 2026-03-15
---
```

| Field | Type | Valid values |
|---|---|---|
| `type` | string | `journal`, `kb`, `todo`, `scratch`, `unsorted` — **first field** |
| `title` | string | Any non-empty string |
| `tags` | list | Empty list `[]` or string values |
| `created` | date | `YYYY-MM-DD` (date-only, emitted as YAML `!!timestamp`) |
| `modified` | date | `YYYY-MM-DD` (date-only, emitted as YAML `!!timestamp`) |

**Parser accepts both formats:** `YYYY-MM-DD` and full ISO 8601 (`2026-03-15T12:32:26Z`) for backwards compatibility. New documents always emit date-only.

Additional frontmatter fields are passed through without validation.

---

## 10. TODO Task Format

```
- [ ] Review PR #42 @work !high (2026-03-14)
- [x] Update DNS records @work (2026-03-12) ✓2026-03-13
- [-] Cancelled task @personal (2026-03-10) ✓2026-03-11
```

Fields: `@category`, `!priority`, `(created-date)`, `✓completed-date`.
Status: `[ ]` = open, `[/]` = in-progress, `[x]` = done, `[-]` = cancelled.
