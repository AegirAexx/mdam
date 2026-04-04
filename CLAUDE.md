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

## TUI and UI Work

**Before making any change to the TUI — new view, new component, modified layout,
new panel, or any styling change — read `docs/TUI-UX.md` in full.**

`docs/TUI-UX.md` is the single source of truth for all interface decisions. It is
not optional and not a suggestion. Every UI output must comply with it. The document
defines spacing, focus indicators, empty states, text hierarchy, status bar layout,
tree anatomy, and a hard list of things the agent must never do.

Key rules that apply to every UI task without exception:
- Every list or tree that can be empty must show a placeholder (§5).
- Focus is always indicated by full-width `Reverse(true)`, never color alone (§3.1).
- All panels have 1-cell horizontal internal padding (§2.1).
- Section headers are always visually distinct from body rows (§2.3, §4).
- Colors always come from `theme.*` — never hardcoded (§4).

## Frontmatter Contract

Field order is mandatory for consistency:

```yaml
---
type: journal        # FIRST.
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

Read `docs/HANDOFF.md` at the start of every session.
Rewrite it at the end of every session — current state only, no history.
Never append. Always replace.
