# Development

## Build and test

```bash
go test ./...                        # Run all tests (mandatory after every change)
go vet ./...                         # Static analysis (mandatory after every change)
go build -o mdam ./cmd/mdam          # Build binary
go test -v ./tui/...                 # TUI tests with verbose output
go test -v ./internal/todo/...       # Package-specific tests
```

**Never commit with a failing test or vet warning.**

## Project structure

```
mdam/
├── cmd/mdam/          # Application entrypoint — delegates to internal/cli
├── internal/
│   ├── cli/           # Cobra subcommand wiring (no business logic)
│   ├── config/        # Configuration loading (Viper, ~/.config/mdam/config.yml)
│   ├── setup/         # First-run detection, config/dir scaffolding, template seeding
│   ├── document/      # Frontmatter model, parsing, validation, kebab-case
│   ├── importer/      # Import pipeline (backlogged — not yet active)
│   ├── journal/       # Daily journal creation, listing, date parsing
│   ├── todo/          # Task parsing (backlogged — sweep/archive not yet active)
│   ├── template/      # Template discovery, render, built-in templates (journal, kb)
│   ├── search/        # Fuzzy search (frontmatter + optional body)
│   ├── export/        # Frontmatter stripping for sharing
│   └── git/           # Git status detection (shells out to git)
├── tui/               # BubbleTea TUI
│   ├── mode.go           # Mode, PanelID, View types and constants
│   ├── keys.go           # KeyMap and DefaultKeyMap()
│   ├── messages.go       # Async message types for engine responses
│   ├── commands.go       # tea.Cmd factories wrapping engine calls
│   ├── model.go          # Model struct, Init/Update, all mode handlers
│   ├── view.go           # View(), tab bar, file/preview panels, status bar
│   ├── view_dashboard.go # Dashboard pane (key 1)
│   ├── view_journal.go   # Journal tree pane (key 2)
│   ├── view_kb.go        # KB subtype tree pane (key 3)
│   ├── view_tags.go      # Tag browser pane (key 4)
│   ├── theme.go          # Theme struct, NewTheme(), 5 palettes
│   ├── icons.go          # Icons struct, DefaultIcons(), PlainIcons()
│   ├── pins.go           # loadPins / savePins / togglePin
│   ├── delete.go         # cmdDeleteDoc, deleteDoneMsg
│   ├── wizard.go         # First-run TUI setup wizard
│   └── tui.go            # Run(cfg) and RunWizard(cfgPath) entry points
└── docs/
    ├── KEYBINDINGS.md    # TUI keybinding reference
    ├── HANDOFF.md        # Current project state for session continuity
    ├── FRONTMATTER.md    # Frontmatter contract and field reference
    ├── TODO-FORMAT.md    # TODO task syntax and sweep/archive behaviour
    ├── CLI.md            # Full CLI subcommand reference
    ├── DEVELOPMENT.md    # This file
    ├── issues/           # Bug reports and feature requests
    ├── specs/            # Feature specifications
    └── reports/          # Session and issue reports
```

## Code style

- **Standard library first.** Only use external packages already in `go.mod`. No new dependencies without explicit approval.
- **Functions are small and pure.** Prefer inputs → outputs over methods with side effects. Keep functions under 50 lines.
- **Error handling.** Return errors, don't panic. Wrap with context: `fmt.Errorf("doing thing: %w", err)`.
- **Naming.** Follow Go conventions: `MixedCaps`, not `snake_case`. Packages are lowercase single words.
- **File paths.** Always use `filepath.Join()`, never string concatenation.

## Testing rules

- Every function gets a table-driven test in a `_test.go` file.
- Use only the `testing` package — no third-party assertion libraries.
- TUI tests use `stripANSI()` for assertions against rendered output (lipgloss adds ANSI codes).

## External dependencies (go.mod)

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI subcommand routing |
| `github.com/spf13/viper` | Config file loading |
| `gopkg.in/yaml.v3` | YAML frontmatter parsing/rendering |
| `github.com/charmbracelet/bubbletea` | TUI event loop |
| `github.com/charmbracelet/lipgloss` | TUI styling and theming |
| `github.com/charmbracelet/bubbles` | TUI components (textinput, viewport) |
| `github.com/charmbracelet/glamour` | Markdown rendering for preview and read mode |
