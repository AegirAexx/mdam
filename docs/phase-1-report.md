# Phase 1 Report — Headless Engine

**Completed:** 2026-03-14
**Goal:** Prove all core business logic with full test coverage, no UI dependency.

---

## Summary

Phase 1 is complete. The headless engine is fully implemented and tested. Every core feature from the spec is accessible as a standalone Go package and exposed via `mdam <subcommand>` CLI. The TUI has not been touched — Phase 2 starts from a clean, well-tested foundation.

**69 tests passing. 0 failures. `go vet` clean.**

---

## What Was Built

### Engine Packages (`internal/`)

| Package | Responsibility | Key exports |
|---|---|---|
| `document` | Frontmatter model, YAML parsing, filename validation, round-trip write | `ParseFile`, `ValidateFrontmatter`, `ValidateFilename`, `ToKebabCase` |
| `config` | Viper-backed config loading from `~/.config/mdam/config.yml` with defaults | `Load`, `LoadFrom`, `Config` |
| `template` | Template discovery, `{{variable}}` interpolation, 5 built-in templates | `Discover`, `Find`, `Render`, `WriteBuiltins` |
| `journal` | Daily entry creation (idempotent), listing, month filter, past entry detection | `Create`, `List`, `ListByMonth`, `PastEntries`, `ParseDate` |
| `todo` | Task line parsing, sweep logic, archive, filter | `ParseTask`, `ParseTasks`, `Sweep`, `Archive`, `FilterTasks` |
| `export` | Frontmatter stripping for sharing | `Strip`, `ToFile` |
| `git` | Git status detection via `git` subprocess | `Status`, `IsRepo`, `IsAvailable` |
| `search` | Fuzzy search over frontmatter and filenames, optional body search | `Search`, `SearchWithBody`, `ListAll` |
| `importer` | Import pipeline: filename validation, frontmatter scaffolding, duplicate detection | `ImportFile`, `ImportDir` |

### CLI Package (`internal/cli/`)

Cobra subcommand tree wiring all engine packages to the `mdam` binary.

### Binary (`cmd/mdam/`)

12-line entrypoint delegating to `internal/cli`.

---

## CLI Reference (Implemented)

```
mdam                                  # placeholder — TUI in Phase 2

mdam journal create [date]            # create entry for today or YYYY-MM-DD
mdam journal list [--month YYYY-MM]   # list entries, newest-first

mdam todo list [--status S] [--category C] [--all]
mdam todo sweep                       # promote incomplete tasks to global TODO
mdam todo archive [--older-than N]    # archive tasks completed N+ days ago

mdam search [query] [--tag T] [--type T] [--modified-after YYYY-MM-DD]

mdam import <path> [--auto-fix] [--dry-run]
mdam export <file> [--to dir]

mdam status [--porcelain]

mdam template list
mdam template show <name>

mdam config
```

---

## Test Coverage

**69 test functions across 9 packages.** All table-driven, using only `testing`.

| Package | Test functions | Notable cases |
|---|---|---|
| `document` | 7 | Missing delimiters, invalid YAML, round-trip write, extra fields |
| `config` | 4 | Defaults with no file, full config load, `~` expansion, dir helpers |
| `template` | 7 | Discovery, find, render, unresolved vars, built-in content |
| `journal` | 9 | Idempotent create, list sort, month filter, past entries |
| `todo` | 11 | All status variants, sweep (remove open / keep done), no-duplicate promotion, archive by age |
| `export` | 4 | Standard strip, no frontmatter error, leading newlines, creates dest dir |
| `git` | 5 | Repo detection, status after file add, untracked/staged/modified flags |
| `search` | 9 | Tag/type/date filters, hidden dir skip, fuzzy matching, body snippet |
| `importer` | 11 | Filename auto-fix, frontmatter scaffold, duplicate detection, dry-run |

---

## Design Decisions

### Frontmatter parsing without goldmark
The spec listed `yuin/goldmark` for frontmatter extraction. In practice, the frontmatter block is plain YAML between `---` delimiters — a simple line-scanner plus `gopkg.in/yaml.v3` handles it cleanly. Goldmark was skipped; it can be added later if full markdown AST parsing is needed.

### `internal/importer` package name
The spec uses `internal/import/` but `import` is a Go keyword. Named `importer` instead.

### TODO task format
Implemented the spec's proposed format verbatim:
```
- [ ] Task text @category !priority (YYYY-MM-DD) ✓YYYY-MM-DD
```
Status encoding: `[ ]` = open, `[x]` = done, `[-]` = cancelled. In-progress is detected by `~text~` prefix convention (can be revisited).

### Sweep logic
The sweep reads the journal's `## TODOs` section, keeps completed tasks as history, removes open tasks, and appends them to the global TODO — deduplicating by task text. Non-TODO content and subsequent heading sections are preserved.

### Search scoring
Four-tier scoring: exact tag match (100) > title substring (50) > filename match (30) > body match (10). Fuzzy matching (subsequence) applies at each tier at half score. Tie-breaks by `modified` timestamp.

---

## Dependencies Added

| Package | Version | Purpose |
|---|---|---|
| `github.com/spf13/cobra` | v1.10.2 | CLI subcommand routing |
| `github.com/spf13/viper` | v1.21.0 | Config file loading |
| `gopkg.in/yaml.v3` | v3.0.1 | Frontmatter YAML parsing |

Transitive deps pulled in by viper (mapstructure, afero, etc.) but not used directly.

---

## What Phase 1 Does Not Include

- TUI (Phase 2)
- `$EDITOR` handoff (Phase 4)
- Lazygit handoff (Phase 4)
- `mdam scratch` / `mdam new` commands (deferred — require editor handoff)
- `mdam kb` commands (deferred — no additional logic beyond search/import)
- Git auto-commit on engine mutations (config flag exists, behaviour not wired)
- Full-text body search is implemented but not the default (opt-in via `SearchWithBody`)

---

## Next: Phase 2 — TUI Skeleton

Establish the BubbleTea framework and event loop. Goals:

1. `tea.Model` with correct MVU structure
2. Basic interactive list rendering (dummy data)
3. Core navigation keybindings: `j`/`k`, `gg`/`G`, `Tab`
4. Command mode (`:` prefix) and search mode (`/` prefix) input handling
5. No styling, no real data, no editor integration

The engine is ready to be consumed — Phase 2 just needs to hook `List`, `Search`, and `ReadTasks` into `Init()` and `Update()`.
