# MadaM — Markdown Admin Management

A keyboard-centric TUI tool for managing markdown documents, daily journals, and TODOs. The filesystem is the database. Your editor does the editing.

MadaM is an administration and routing tool — it organizes, navigates, and automates workflows around your markdown files. All viewing and authoring is delegated to your `$EDITOR`.

## Status

**Pre-release — Phase 1 (Headless Engine)**

## Features

- **Markdown management** — Frontmatter validation, filename enforcement, folder-based categorization
- **Daily journals** — Auto-created from templates, named `YYYY-MM-DD.md`
- **TODO system** — Two-tier task management with automatic sweep from journal entries to a global backlog
- **Knowledge base** — Organized reference documents with user-defined taxonomy
- **Scratch pad** — Persistent clipboard for ephemeral content, always one keybinding away
- **Templates** — User-extensible document scaffolding with variable interpolation
- **Search** — Fuzzy finding across frontmatter, filenames, and document content
- **Export** — Strip frontmatter and share clean markdown
- **Git integration** — Status awareness in the TUI, lazygit handoff for actions
- **Dual interface** — Full TUI and headless CLI with UNIX-style flags and subcommands

## Requirements

- Go 1.26+
- `$EDITOR` set (e.g., `nvim`, `vim`, `nano`)
- Git (for version control and sync features)
- [lazygit](https://github.com/jesseduffield/lazygit) (optional, for git handoff)

## Installation

```bash
go install github.com/AegirAexx/mdam@latest
```

Or build from source:

```bash
git clone https://github.com/AegirAexx/mdam.git
cd mdam
go build -o mdam ./cmd/mdam
```

## Quick Start

```bash
# Initialize configuration
mdam config --init

# Create today's journal entry
mdam journal create

# Open the scratch pad
mdam scratch

# List open TODOs
mdam todo list

# Search for a document
mdam search "nginx setup"

# Launch the TUI
mdam
```

## Configuration

MadaM reads its configuration from `~/.config/mdam/config.yml`. Run `mdam config --init` to generate a default configuration, or `mdam config --edit` to open it in your editor.

See [docs/mdam-spec-v1.md](docs/mdam-spec-v1.md) for the full configuration reference.

## CLI Reference

```
mdam                                     Launch TUI (default)
mdam ui                                  Launch TUI (explicit)

mdam journal create [date]               Create journal entry
mdam journal list [--month YYYY-MM]      List journal entries

mdam todo list [--status S] [--category C] [--all]
mdam todo sweep                          Run TODO sweep
mdam todo archive [--older-than N]       Archive completed tasks

mdam kb list [--type T]                  List KB documents
mdam kb create --template T --title "T"  Create KB doc

mdam search "query" [--tag T] [--type T] [--modified-after D]

mdam scratch                             Open scratch pad
mdam new [--template T] [--title "T"]    Create document from template

mdam import <path> [--auto-fix] [--dry-run]
mdam export <file> [--to DIR] [--clipboard]

mdam status [--porcelain]                Git status summary

mdam template list                       List templates
mdam template show <name>                Show template content

mdam config                              Show configuration
mdam config --edit                       Open config in $EDITOR
mdam config --init                       Generate default config
```

## Documentation

- [Project Specification](docs/mdam-spec-v1.md) — Full feature spec, architecture, and execution plan
- [Keybindings](docs/KEYBINDINGS.md) — TUI keybinding reference

## Project Structure

```
mdam/
├── cmd/mdam/          # Application entrypoint
├── internal/
│   ├── config/        # Configuration loading and validation
│   ├── document/      # Markdown document model and frontmatter parsing
│   ├── import/        # Import pipeline and validation
│   ├── journal/       # Journal creation and management
│   ├── todo/          # TODO parsing, sweep, and archive
│   ├── template/      # Template discovery and interpolation
│   ├── search/        # Search and fuzzy matching
│   ├── export/        # Frontmatter stripping and export
│   └── git/           # Git status detection
├── tui/               # BubbleTea TUI (Phase 2+)
├── docs/              # Project documentation
├── CLAUDE.md          # Agent context file
├── README.md
└── go.mod
```

## Development

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/todo/...

# Build
go build -o mdam ./cmd/mdam

# Vet
go vet ./...
```

## License

MIT — see [LICENSE](LICENSE) for details.
