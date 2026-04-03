# mdam — Project Handoff

mdam (Markdown Admin Management) is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, TODOs, scratch notes. The project is at v0.1.0 alpha. All five development phases are implemented and all bug fixes (issues #1, #2, #3, #4, #5) have been applied and merged. No systematic testing pass has been run yet — every feature is implemented but none has been individually verified as working correctly.

---

## Known Issues

_(none logged — populate as testing reveals problems)_

---

## Next Session: Start Here

Begin the systematic testing pass. Work through packages in order:

1. `internal/document` — frontmatter parsing, validation, kebab-case
2. `internal/template` — rendering, variable precedence, `TemplateType()`
3. `internal/journal` — create, list, date parsing
4. `internal/todo` — parse, sweep, archive, filter
5. `internal/search` — fuzzy search, tag/type/date filters
6. `internal/importer` — validate, auto-fix, duplicate detection
7. `internal/export` — frontmatter stripping
8. `internal/setup` — first-run detection, scaffolding, idempotence
9. `internal/config` — path helpers, Viper loading
10. `internal/git` — git status detection
11. `tui/` — keybindings, view switching, modal flows, editor handoff

For each package: read existing tests, identify gaps, add missing table-driven tests, run `go test ./... && go vet ./...`.

Gate for v0.2.0: all packages have meaningful test coverage and `go test ./...` is green.

---

## Deferred (out of scope for v1)

- AI / Agent integration
- Multi-device conflict resolution
- Structured TODO format
- File watchers (`fsnotify`)
- TODO-specific keybindings
- `g` vs `ctrl+g` for lazygit
- Arrow key navigation
