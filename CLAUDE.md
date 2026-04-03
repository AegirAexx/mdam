# CLAUDE.md — mdam Project Context

mdam is a Go TUI tool for managing markdown documents, journals, and TODOs.
It is an admin/routing tool — it never edits document bodies. All editing is delegated to `$EDITOR`. The filesystem is the database. No caching, no in-memory state that persists between function calls.

**Status: v0.1.0 — early alpha. No features have been individually tested.**

## Mandatory After Every Change

- `go test ./...` — no exceptions, never commit a failing test
- `go vet ./...` — no exceptions, never commit a vet warning

## Code Rules

- **Stdlib first.** Never add a new dependency without explicit approval.
- **No panics.** Return errors. Wrap with context: `fmt.Errorf("doing thing: %w", err)`
- **File paths** always use `filepath.Join()`, never string concatenation.
- **Functions** are small and pure. Under 50 lines where possible.
- **Tests** are table-driven, in `_test.go` files, using only the `testing` package.
- **TUI tests** use `stripANSI()` helper for assertions against rendered output.

## Non-Obvious Constraints

- The app never writes to a document's markdown body except during TODO sweep.
- Frontmatter updates are the only other permitted file mutations.
- No editing. No caching. The filesystem is always the source of truth.

## Frontmatter Contract

Field order is mandatory for consistency:

```yaml
---
type: journal        # FIRST. One of: journal, kb, todo, scratch, unsorted
title: 2026-03-15
tags: []
created: 2026-03-15
modified: 2026-03-15
---
```

- Dates emit as `YYYY-MM-DD` unquoted YAML `!!timestamp` via `*yaml.Node{Tag: "!!timestamp"}`
- Parser also accepts full ISO 8601 for backwards compatibility
- Additional fields pass through without validation

## Versioning

- **Patch** — bug fixes, small corrections during testing
- **Minor** — completing a testing pass, notable behavior/CLI/config changes
- **1.0.0** — all features tested, CLI and config format locked

## Working on Issues

1. Read the issue completely before touching anything.
2. State your plan and list files you'll modify before writing code.
3. Make the smallest change that fixes the issue. Don't touch unrelated code.
4. Add or update tests covering the fix.
5. Run `go test ./...` and `go vet ./...` before declaring done.
6. Summarize what changed.

## Session Continuity

`docs/HANDOFF.md` is the current project state for the next session.
Read it at the start. Rewrite it at the end — current state only, no history.
