# MadaM — Markdown Admin Management

A keyboard-centric TUI tool for managing markdown documents, daily journals, and TODOs. The filesystem is the database. Your editor does the editing.

MadaM is an administration and routing tool — it organizes, navigates, and automates workflows around your markdown files. All viewing and authoring is delegated to your `$EDITOR`.

## Status

**All 5 phases complete — production-ready**

| Phase | Scope | Status |
|---|---|---|
| 1 | Headless engine (config, scan, parse, TODO, templates, search, export, git, CLI) | Complete |
| 2 | BubbleTea TUI skeleton (event loop, keybindings, panel layout, dummy data) | Complete |
| 3 | Integration (real data, git status bar, search, export, template picker wired) | Complete |
| 4 | Editor handoff (`$EDITOR` + lazygit via `tea.ExecProcess`, suspend/resume) | Complete |
| 5 | Polish (lipgloss styling, glamour preview, theming, ambient findability views) | Complete |

All core engine operations work headlessly via CLI subcommands. The TUI is fully polished with theming, markdown preview, and ambient findability views. Press `Enter` on any document to open it in your editor.

## Features

- **Markdown management** — Frontmatter validation, filename enforcement, folder-based categorization
- **Daily journals** — Auto-created from templates, named `YYYY-MM-DD.md`
- **TODO system** — Two-tier task management with automatic sweep from journal entries to a global backlog
- **Knowledge base** — Organized reference documents with user-defined taxonomy
- **Scratch pad** — Persistent clipboard for ephemeral content, always one keybinding away
- **Templates** — User-extensible document scaffolding with variable interpolation
- **Search** — Fuzzy finding across frontmatter, filenames, and document content
- **Export** — Strip frontmatter and share clean markdown
- **Git integration** — Status awareness in the TUI, lazygit handoff via `ctrl+g`
- **Dual interface** — Full TUI and headless CLI with UNIX-style flags and subcommands
- **Color theming** — Five built-in palettes: tokyonight, nord, gruvbox, catppuccin, dracula
- **Markdown preview** — Glamour-rendered live preview in the viewport panel
- **Nerd Font icons** — Optional icon set for file types, git status, and TODO states (`nerd_fonts: false` default)
- **Dashboard** (`1`) — Today's journal, open TODO count, pinned docs, and recent documents
- **Tag browser** (`6`) — All tags with document counts; drill into any tag to list its documents
- **Smart filter** (`f`) — Cycle through Untagged / Stale / Inbox post-filters on the document list
- **Pin / unpin** (`p`) — Bookmark any document; pins persist to `~/.config/mdam/pins.json`
- **Delete with confirmation** (`d` → `y`/`n`) — Safe delete mode prevents accidental removal

## Requirements

- Go 1.21+
- `$EDITOR` set (e.g., `nvim`, `vim`, `nano`)
- Git (for version control and sync features)
- [lazygit](https://github.com/jesseduffield/lazygit) (optional, for git handoff via `ctrl+g`)

## Installation

Build from source:

```bash
git clone https://github.com/AegirAexx/mdam.git
cd mdam
go build -o mdam ./cmd/mdam
```

## Quick Start

```bash
# Create today's journal entry
mdam journal create

# List open TODOs
mdam todo list

# Search for a document
mdam search "nginx setup"

# Import a markdown file
mdam import ~/Downloads/notes.md

# Launch the TUI
mdam
```

## TUI

Run `mdam` with no subcommand to launch the interactive TUI.

```
▶ Files ──────────────────────│─ Preview ──────────────────────────
> 2026-03-14.md          [M]  │  Journal 2026-03-14
  2026-03-13.md               │
  setup-nginx.md         [?]  │  type: journal
  deploy-runbook.md           │  tags: daily
  meeting-notes-…             │  modified: 2026-03-14
                              │
                              │  2026-03-14.md
                              │─ TODOs ─────────────────────────────
                              │> - [ ] Review PR #42 @work
                              │  - [ ] Buy groceries @personal
──────────────────────────────────────────────────────────────────
 NORMAL │ main ↑2 │ 5 docs              /search  :cmd  ?help  q:quit
```

### Keybindings

| Key | Action |
|---|---|
| `j` / `k` | Move down / up |
| `h` / `l` | Previous / next panel |
| `Tab` / `Shift+Tab` | Cycle panel focus |
| `gg` / `G` | Jump to top / bottom |
| `1` | Dashboard (today's context) |
| `2` | Journal entries |
| `3` | Knowledge base |
| `4` | TODO panel |
| `5` | Recently modified |
| `6` | Tag browser |
| `/` | Fuzzy search |
| `:` | Command mode |
| `?` | Help overlay |
| `n` | New document (template picker) |
| `d` | Delete selected document (prompts `y`/`n`) |
| `e` | Export selected document |
| `p` | Pin / unpin selected document |
| `f` | Cycle smart filter (None → Untagged → Stale → Inbox) |
| `R` | Re-scan filesystem |
| `Enter` | Open selected document in `$EDITOR` |
| `ctrl+g` | Open lazygit |
| `s` | Open scratch pad in `$EDITOR` |
| `q` | Quit |

### Command Mode (`:`)

| Command | Action |
|---|---|
| `:q` / `:quit` | Quit |
| `:todo sweep` | Run TODO sweep manually |
| `:todo archive` | Archive old completed tasks |

### Git Status Markers

Files in the managed tree show their git status inline:

| Marker | Meaning |
|---|---|
| `[M]` | Modified (working tree) |
| `[A]` | Staged |
| `[?]` | Untracked |

## Configuration

MadaM reads `~/.config/mdam/config.yml`. If the file does not exist, sensible defaults are used.

```yaml
editor: nvim                          # falls back to $EDITOR env var
author: "Your Name"
base_dir: ~/notes                     # root of your managed document tree
export_dir: ~/Downloads

import:
  inbox_dir: ~/notes/.inbox
  auto_fix: false

theme: tokyonight                     # nord, gruvbox, catppuccin, dracula
nerd_fonts: false                     # set true if your terminal font has Nerd Font glyphs

git:
  enabled: true
  auto_commit: false
  lazygit: true

todo:
  default_category: personal
  archive_after_days: 30

journal:
  auto_create: true
  sweep_on_create: true
```

Open the config in your editor:

```bash
mdam config --edit
```

## Frontmatter Contract

Every managed document requires these YAML frontmatter fields:

```yaml
---
type: kb   # journal | kb | todo | scratch | unsorted
title: "My Document"
tags: [devops, nginx]
created: 2026-03-14
modified: 2026-03-14
---
```

## TODO Task Format

```
- [ ] Review PR #42 @work !high (2026-03-14)
- [x] Update DNS records @work (2026-03-12) ✓2026-03-13
```

Fields: `@category`, `!priority`, `(created-date)`, `✓completed-date`.

## CLI Reference

```
mdam                                     Launch TUI (default)

mdam journal create [date]               Create journal entry (today if no date)
mdam journal list [--month YYYY-MM]      List journal entries

mdam todo list [--status S] [--category C] [--all]
mdam todo sweep                          Run TODO sweep manually
mdam todo archive [--older-than N]       Archive completed tasks

mdam search "query" [--tag T] [--type T] [--modified-after D]

mdam import <path> [--auto-fix] [--dry-run]
mdam export <file> [--to DIR]

mdam status [--porcelain]                Git status summary for managed tree

mdam template list                       List available templates
mdam template show <name>                Display template content

mdam config                              Show current configuration
mdam config --edit                       Open config.yml in $EDITOR
```

## Project Structure

```
mdam/
├── cmd/mdam/          # Application entrypoint
├── internal/
│   ├── config/        # Configuration loading (Viper, config.yml)
│   ├── document/      # Markdown document model, frontmatter parsing/validation
│   ├── importer/      # Import pipeline, filename and frontmatter validation
│   ├── journal/       # Journal creation, date management
│   ├── todo/          # TODO parsing, sweep logic, archive
│   ├── template/      # Template discovery and variable interpolation
│   ├── search/        # Fuzzy search across frontmatter and filenames
│   ├── export/        # Frontmatter stripping for sharing
│   └── git/           # Git status detection (shells out to git)
├── tui/               # BubbleTea TUI
│   ├── mode.go           # Mode, PanelID, View types
│   ├── keys.go           # KeyMap and DefaultKeyMap()
│   ├── messages.go       # Async message types for engine responses
│   ├── commands.go       # tea.Cmd factories wrapping engine calls
│   ├── model.go          # Model struct, Init/Update, all mode handlers
│   ├── view.go           # View(), panel rendering, status bar
│   ├── view_dashboard.go # Dashboard view (key 1)
│   ├── view_tags.go      # Tag browser view (key 6)
│   ├── theme.go          # Theme struct, NewTheme(), 5 palettes
│   ├── icons.go          # Icons struct, DefaultIcons(), PlainIcons()
│   ├── pins.go           # loadPins/savePins/togglePin
│   ├── delete.go         # cmdDeleteDoc, deleteDoneMsg
│   └── tui.go            # Run(cfg) entry point
├── docs/
│   ├── KEYBINDINGS.md          # TUI keybinding reference
│   ├── HANDOFF.md              # Complete project state for future sessions
│   ├── issues/                 # Bug reports and feature requests
│   ├── reports/
│   │   ├── kick-off/           # Phase implementation reports (1–5)
│   │   └── issues/             # Per-issue fix reports
│   └── specs/
│       └── mdam-spec-v1.md     # Full project specification
├── CLAUDE.md          # Agent context and project rules
└── go.mod
```

## Development

```bash
go test ./...                        # Run all tests (119 TUI tests + engine tests)
go vet ./...                         # Static analysis
go build -o mdam ./cmd/mdam          # Build binary
go test -v ./tui/...                 # TUI tests with output
go test -v ./internal/todo/...       # Package-specific tests
```

## Documentation

- [Project Specification](docs/specs/mdam-spec-v1.md) — Full feature spec, architecture, and execution plan
- [Keybindings](docs/KEYBINDINGS.md) — TUI keybinding reference
- [Handoff](docs/HANDOFF.md) — Complete project state for future sessions
- [Reports](docs/reports/) — Phase kick-off and issue fix reports

## License

MIT — see [LICENSE](LICENSE) for details.
