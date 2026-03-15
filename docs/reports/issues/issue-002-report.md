# Issue #4 Fix Report — Frontmatter Format + Docs Reorganization

**Branch:** `fix/issue-04`
**Report index:** 002 (second issue fix session)
**Date:** 2026-03-15

---

## Problem Statement

Two separate but related housekeeping issues addressed in one session:

1. **Frontmatter format inconsistency** — All document creation paths emitted full ISO 8601 timestamps (`2026-03-15T12:32:26Z`) for `created`/`modified` and placed `type` at the bottom of the frontmatter block. The desired format was date-only (`2026-03-15`) with `type` as the first field.

2. **Docs folder clutter** — `docs/kick-off-reports/` and `docs/issue-reports/` were flat siblings of `docs/specs/`. Reorganizing them into a typed hierarchy (`docs/reports/kick-off/`, `docs/reports/issues/`) keeps the tree extensible.

---

## Changes Made

### `internal/document/document.go`

**`RenderFrontmatter`** — rewrote the local serialization struct:

- Field order changed from `title → tags → created → modified → type` to `type → title → tags → created → modified`.
- Date fields changed from `time.Time` (which yaml.v3 marshals as RFC3339) to `*yaml.Node` with `Tag: "!!timestamp"`. This is the critical detail: if you marshal a Go `string` that looks like a date, yaml.v3 **quotes it** (`created: "2026-03-14"`), and then on read-back yaml.v3 treats the quoted value as `!!str` and tries `time.Parse(time.RFC3339, ...)`, which fails for date-only strings. Using a `*yaml.Node` with `!!timestamp` bypasses the quoting logic and emits `created: 2026-03-14` (unquoted), which yaml.v3 can round-trip into `time.Time` via its internal `timestampFormats` list (which includes `"2006-01-02"`).
- Added `dateNode(t time.Time) *yaml.Node` helper.

```go
// Before
type fmYAML struct {
    Title    string    `yaml:"title"`
    Tags     []string  `yaml:"tags"`
    Created  time.Time `yaml:"created"`   // → "2026-03-15T00:00:00Z"
    Modified time.Time `yaml:"modified"`
    Type     string    `yaml:"type"`
    Extra    map[string]interface{} `yaml:",inline"`
}

// After
type fmYAML struct {
    Type     string     `yaml:"type"`
    Title    string     `yaml:"title"`
    Tags     []string   `yaml:"tags"`
    Created  *yaml.Node `yaml:"created"`   // → 2026-03-15 (unquoted timestamp)
    Modified *yaml.Node `yaml:"modified"`
    Extra    map[string]interface{} `yaml:",inline"`
}
```

The `Frontmatter` struct and `parseFrontmatter` are unchanged. The parser continues to accept both formats.

---

### `internal/template/template.go`

**`Render()`** — fixed a precedence bug: built-in vars (`{{date}}`, `{{date_short}}`) were resolved **before** caller-supplied vars, which meant `renderJournalTemplate`'s journal-specific `date_short` was silently ignored (the built-in already replaced the placeholder with today's date). Caller vars are now applied first; built-ins fill any remaining unresolved placeholders.

```go
// Before: built-ins first, caller second → caller can never override date vars
// After:  caller first, built-ins second → caller wins
```

**Built-in templates (all 5)** — two changes each:
- `type` field moved to first position.
- `{{date}}` → `{{date_short}}` for `created` and `modified` lines.
- Journal template title: `Journal {{date_short}}` → `{{date_short}}`.

---

### `internal/journal/journal.go`

**`ScaffoldFrontmatter`** — title changed from `"Journal " + date.Format(DateFormat)` to `date.Format(DateFormat)`. Journal entries are already identified by their `type: journal` field; the "Journal " prefix was redundant.

---

### `tui/commands.go`

**`cmdEnsureAndOpenScratch`** — inline frontmatter string updated to new field order and `"2006-01-02"` date format.

---

### Docs reorganization

```
Before:
  docs/kick-off-reports/phase-{1-5}-report.md
  docs/issue-reports/issue-001-report.md

After:
  docs/reports/kick-off/phase-{1-5}-report.md
  docs/reports/issues/issue-001-report.md
  docs/reports/issues/issue-002-report.md  ← this file
```

`README.md` documentation links updated accordingly (individual kick-off report links removed, replaced with a single `Reports` directory link). `HANDOFF.md` and `CLAUDE.md` frontmatter contract sections updated to document the new field order, date-only format, and `!!timestamp` emit behavior.

---

## Tests Fixed

The `Render()` precedence fix resolved a latent test fragility: `TestCreate` in `internal/journal` was checking for the journal-specific date in the rendered file content, but the old code had built-in `{{date_short}}` (today's date) silently winning over the caller's date. The test was only passing on the day it was written (2026-03-14). After the fix, caller-supplied dates propagate correctly and the test is date-independent.

---

## Verification

```
go test ./...   → all 12 packages pass
go vet ./...    → clean
go build        → binary builds
```
