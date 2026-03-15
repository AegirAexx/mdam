# Issue #4: Standardize Frontmatter Format

**Type:** Enhancement
**Priority:** Medium
**Labels:** `enhancement`, `frontmatter`, `ux`, `phase-6`

## Problem

The current frontmatter generation produces timestamps with full ISO 8601 precision (`2026-03-15T12:32:26Z`) and uses an ordering that buries the most important field (`type`) at the bottom.

### Current output (wrong)

```yaml
---
title: Journal 2026-03-15
tags: []
created: 2026-03-15T12:32:26Z
modified: 2026-03-15T12:32:26Z
type: journal
---
```

### Desired output (correct)

```yaml
---
type: journal
title: 2026-03-15
tags: []
created: 2026-03-15
modified: 2026-03-15
---
```

## Changes

### 1. Field ordering

Standardize frontmatter field order across all document creation paths:

```yaml
type:      # What is this document? Most important — comes first.
title:     # Human-readable title
tags:      # Categorization
created:   # When it was made
modified:  # When it was last touched
```

### 2. Date format

Use date-only format (`YYYY-MM-DD`) instead of full ISO 8601 timestamps for `created` and `modified`.

**Justification:** The core engine uses dates at day granularity for all current features:
- TODO sweep operates on date boundaries (today vs. past)
- Search `--modified-after` filter uses dates
- Archive threshold is in days
- Journal filenames are `YYYY-MM-DD.md`

Sub-day precision adds visual noise with no functional benefit. If sub-day precision is needed later, the filesystem's mtime is always available as a fallback.

### 3. Journal title format

Journal entries should use the date as the title (`2026-03-15`), not a prefixed string (`Journal 2026-03-15`). The `type: journal` field already identifies it as a journal entry. The title should be the date itself for clean display in the file list and search results.

## Scope

This affects:

- `internal/document/` — Frontmatter serialization (field ordering, date format)
- `internal/journal/` — Journal template rendering (title format)
- `internal/template/` — Built-in template content (field order in templates)
- `internal/todo/` — TODO file creation if it uses frontmatter
- `tui/commands.go` — `cmdEnsureAndOpenScratch` frontmatter generation

This does **not** affect frontmatter parsing — the parser should accept both formats (full ISO 8601 and date-only) for backwards compatibility with existing documents.

## Acceptance Criteria

- [ ] All newly created documents use the standardized field order
- [ ] All newly created documents use `YYYY-MM-DD` date format
- [ ] Journal titles use date-only format (no "Journal " prefix)
- [ ] Existing documents with full ISO 8601 timestamps still parse correctly
- [ ] `go test ./...` passes
