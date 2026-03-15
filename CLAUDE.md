# CLAUDE.md — MadaM Project Context

MadaM (Markdown Admin Management) is a Go TUI tool that manages markdown documents, journals, and TODOs. It is an admin/routing tool — it never edits documents. All editing is delegated to `$EDITOR`. The filesystem is the database.

## Commands

```bash
go test ./...                 # Run all tests (MANDATORY after every change)
go vet ./...                  # Static analysis (MANDATORY after every change)
go build -o mdam ./cmd/mdam   # Build binary
```

**Never commit with a failing test or vet warning.**

## Architecture

- `cmd/mdam/` — Application entrypoint, Cobra root command
- `internal/cli/` — Cobra subcommand wiring — no business logic here
- `internal/config/` — Config loading (Viper, `~/.config/mdam/config.yml`)
- `internal/document/` — Markdown document model, frontmatter parsing/validation
- `internal/importer/` — Import pipeline, filename and frontmatter validation
- `internal/journal/` — Journal creation, date management
- `internal/todo/` — TODO parsing, sweep logic, archive
- `internal/template/` — Template discovery and variable interpolation
- `internal/search/` — Fuzzy search across frontmatter and filenames
- `internal/export/` — Frontmatter stripping for sharing
- `internal/git/` — Git status detection (shells out to `git`)
- `tui/` — BubbleTea TUI (all TUI code lives here)

## Rules

### Code Style

- **Standard library first.** ALWAYS prefer Go stdlib over third-party packages. Only use external dependencies listed in go.mod. Do NOT add new dependencies without explicit approval.
- **Functions are small and pure.** Prefer functions that take inputs and return outputs over methods with side effects. Keep functions under 50 lines where possible.
- **Error handling.** Return errors, don't panic. Wrap errors with context using `fmt.Errorf("doing thing: %w", err)`.
- **Naming.** Follow Go conventions: `MixedCaps`, not `snake_case`. Packages are lowercase single words. Exported functions have doc comments.
- **File paths.** Always use `filepath.Join()`, never string concatenation for paths.

### Testing

- **Test everything.** Every function gets a table-driven test in a `_test.go` file.
- **Run `go test ./...` after every change.** No exceptions.
- **Tests use only the `testing` package** — no third-party assertion libraries.
- **TUI tests use `stripANSI()` helper** for assertions against rendered output.

### Data

- **No caching, no database.** Read from the filesystem on every operation. No in-memory caches that persist between function calls. The filesystem is the source of truth.
- **No editing.** The application never writes to the markdown body of a document except during TODO sweep. Frontmatter updates are the only other permitted file mutations.

## Frontmatter Contract

Every managed document has YAML frontmatter. Field order matters for consistency:

```yaml
---
type: journal              # Required. One of: journal, kb, todo, scratch, unsorted
title: 2026-03-15          # Required. Human-readable title
tags: []                   # Required. List of strings
created: 2026-03-15        # Required. YYYY-MM-DD format
modified: 2026-03-15       # Required. YYYY-MM-DD format
---
```

Additional fields are passed through without validation.

**Date format:** Use `YYYY-MM-DD` (date-only) for `created` and `modified`. The parser must also accept full ISO 8601 (`2026-03-15T12:32:26Z`) for backwards compatibility.

## Folder Structure Convention

```
{base_dir}/
├── journal/           # Daily journal entries (YYYY-MM-DD.md)
├── kb/                # Knowledge base documents
├── todo/              # Global TODO file
├── scratch/           # Scratch pad singleton
├── .inbox/            # Import inbox
└── .templates/        # User-defined and built-in templates
```

Document type determines destination directory:

- `type: journal` → `journal/YYYY-MM-DD.md`
- `type: kb` → `kb/{kebab-title}.md`
- `type: scratch` → `scratch/scratch.md`
- `type: todo` → `todo/todo.md`
- `type: unsorted` → `{base_dir}/{kebab-title}.md`

## Working on Issues

When given a GitHub issue or bug report:

1. **Read the issue completely** before making any changes.
2. **State your plan** before writing code. List the files you'll modify and why.
3. **Make the smallest change that fixes the issue.** Don't refactor unrelated code.
4. **Update or add tests** that cover the fix.
5. **Run `go test ./...` and `go vet ./...`** before declaring done.
6. **Summarize what you changed** after completing the work.

## Key Documentation

- `docs/specs/mdam-spec-v1.md` — Full project specification
- `docs/KEYBINDINGS.md` — TUI keybinding reference
- `docs/HANDOFF.md` — Complete project state for future sessions
