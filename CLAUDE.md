# CLAUDE.md — MadaM Project Context

MadaM (Markdown Admin Management) is a Go TUI tool that manages markdown documents, journals, and TODOs. It is an admin/routing tool — it never edits documents. All editing is delegated to `$EDITOR`. The filesystem is the database.

## Commands

```bash
go test ./...           # Run all tests (do this after every change)
go vet ./...            # Static analysis (do this after every change)
go build -o mdam ./cmd/mdam  # Build binary
```

## Architecture

- `cmd/mdam/` — Application entrypoint, Cobra root command
- `internal/config/` — Config loading (Viper, `~/.config/mdam/config.yml`)
- `internal/document/` — Markdown document model, frontmatter parsing/validation
- `internal/import/` — Import pipeline, filename and frontmatter validation
- `internal/journal/` — Journal creation, date management
- `internal/todo/` — TODO parsing, sweep logic, archive
- `internal/template/` — Template discovery and variable interpolation
- `internal/search/` — Fuzzy search across frontmatter and filenames
- `internal/export/` — Frontmatter stripping for sharing
- `internal/git/` — Git status detection (shells out to `git`)
- `tui/` — BubbleTea TUI (Phase 2+, not yet implemented)

## Rules

- **Standard library first.** ALWAYS prefer Go stdlib over third-party packages. Only use external dependencies listed in go.mod. Do NOT add new dependencies without explicit approval.
- **Test everything.** Every function gets a table-driven test in a `_test.go` file. Run `go test ./...` after every change. Tests use only the `testing` package — no third-party assertion libraries.
- **No caching, no database.** Read from the filesystem on every operation. No in-memory caches that persist between function calls. The filesystem is the source of truth.
- **No editing.** The application never writes to the markdown body of a document except during TODO sweep. Frontmatter updates are the only other permitted file mutations.
- **Functions are small and pure.** Prefer functions that take inputs and return outputs over methods with side effects. Keep functions under 50 lines where possible.
- **Error handling.** Return errors, don't panic. Wrap errors with context using `fmt.Errorf("doing thing: %w", err)`.
- **Naming.** Follow Go conventions: `MixedCaps`, not `snake_case`. Packages are lowercase single words. Exported functions have doc comments.
- **File paths.** Always use `filepath.Join()`, never string concatenation for paths.

## Frontmatter Contract

Every managed document has YAML frontmatter with these required fields:
- `title` (string), `tags` (list), `created` (ISO 8601), `modified` (ISO 8601), `type` (string: journal/kb/todo/scratch/unsorted)

## Key Documentation

- @docs/mdam-spec-v1.md — Full project specification (features, architecture, execution plan)
- @docs/KEYBINDINGS.md — TUI keybinding reference
