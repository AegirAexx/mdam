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
│   ├── importer/      # Import pipeline — validate, auto-fix, duplicate detection
│   ├── journal/       # Daily journal creation, listing, date parsing
│   ├── todo/          # Task parsing, sweep logic, archive, filter
│   ├── template/      # Template discovery, render, built-in templates
│   ├── search/        # Fuzzy search (frontmatter + optional body)
│   ├── export/        # Frontmatter stripping for sharing
│   └── git/           # Git status detection (shells out to git)
├── tui/               # BubbleTea TUI
│   ├── mode.go           # Mode, PanelID, View types and String() methods
│   ├── keys.go           # KeyMap and DefaultKeyMap()
│   ├── messages.go       # Async message types for engine responses
│   ├── commands.go       # tea.Cmd factories wrapping engine calls
│   ├── model.go          # Model struct, Init/Update, all mode handlers
│   ├── view.go           # View(), panel rendering, status bar
│   ├── view_dashboard.go # Dashboard view (key 1)
│   ├── view_tags.go      # Tag browser view (key 6)
│   ├── theme.go          # Theme struct, NewTheme(), 5 palettes
│   ├── icons.go          # Icons struct, DefaultIcons(), PlainIcons()
│   ├── pins.go           # loadPins / savePins / togglePin
│   ├── delete.go         # cmdDeleteDoc, deleteDoneMsg
│   └── tui.go            # Run(cfg) entry point
└── docs/
    ├── KEYBINDINGS.md    # TUI keybinding reference
    ├── HANDOFF.md        # Complete project state for future sessions
    ├── FRONTMATTER.md    # Frontmatter contract and field reference
    ├── TODO-FORMAT.md    # TODO task syntax and sweep/archive behaviour
    ├── CLI.md            # Full CLI subcommand reference
    ├── DEVELOPMENT.md    # This file
    ├── issues/           # Bug reports and feature requests
    ├── reports/
    │   ├── kick-off/     # Phase implementation reports (1–5)
    │   └── issues/       # Per-issue fix reports
    └── specs/
        └── mdam-spec-v1.md  # Full project specification
```

## Code style

- **Standard library first.** Only use external packages already in `go.mod`. Do not add new dependencies without explicit approval.
- **Functions are small and pure.** Prefer functions that take inputs and return outputs over methods with side effects. Keep functions under 50 lines.
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
| `github.com/charmbracelet/glamour` | Markdown rendering for preview panel |
