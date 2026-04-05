# mdam — Project Handoff

mdam (Markdown Admin Management) is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, TODOs, scratch notes.

**Current branch:** `fix/ui-ux-refactoring`

---

## Session 04 — COMPLETE

Second UX/UI refactoring pass covering navigation, preview, templates, and polish.

### Changes

| Area | Description |
|------|-------------|
| h/l tree navigation | Journal (2) and KB (3): h collapses folder (or parent if on file row), l expands folder (no-op on file row). Panes 1/4 unchanged. |
| Journal/KB preview | Strips YAML frontmatter; shows tags line + glamour-rendered body |
| `{{date:FORMAT}}` templates | Parameterised date via Go time layout strings; `template.RenderAt(t, vars, now)` for backdating |
| Stale preview fix | `m.preview.SetContent("")` on view switch to Journal/KB |
| Journal cursor fallback | `initJournalView` expands most recent past-month when current month is empty |
| Scan error surfacing | `search.ListAll` returns skip count; shown in status bar |
| `WriteBuiltins` overwrite | Stale on-disk templates overwritten when content differs from built-in |
| Pin indicators | `[*]` marker on pinned docs in all file lists |
| Create doc cursor | `viewTemplatePicker` selected item uses `Reverse(true)` |
| Dashboard alignment | Removed today-prefix `[*]` — all rows aligned |
| Tab bar | Labels now `1: Dashboard`, `2: Journal`, `3: KB`, `4: Tag Browser`; active uses `Reverse(true)` |
| KB/Journal preview placeholder | Shows "Select a document to preview." when cursor on folder |

### Files modified
- `tui/model.go` — view-switch preview clear, h/l tree nav, docsLoadedMsg re-init
- `tui/view.go` — tab bar labels+style, template picker cursor, preview panel tree-view branch
- `tui/view_dashboard.go` — removed today-prefix
- `tui/view_journal.go` — initJournalView fallback to past month
- `tui/view_kb.go` — h/l tree navigation matching journal
- `tui/commands.go` — cmdLoadPreviewDoc strips frontmatter; cmdLoadDocs captures skipCount
- `internal/search/search.go` — ListAll returns skip count
- `internal/template/template.go` — RenderAt, dateFormatRe, WriteBuiltins content-diff overwrite
- `internal/journal/journal.go` — uses template.RenderAt for backdating
- `tui/model_test.go` — updated for new signatures and journal view setup
- `docs/KEYBINDINGS.md` — h/l behaviour for journal/KB updated
- `docs/FRONTMATTER.md` — template variables and `{{date:FORMAT}}` reference
- `README.md` — tab bar ASCII art updated, template feature description

All tests pass: `go test ./...` and `go vet ./...` green.

---

## Next Session: Start Here

Branch `fix/ui-ux-refactoring` is ready for review and PR to master.

Manual smoke test checklist:
1. **Tab bar** — numbered labels (`1: Dashboard` etc.); active tab inverted
2. **Create doc screen** (`n`) — selected item inverted, no cursor icon
3. **Dashboard** — all items aligned (no stray `[*]` prefix)
4. **Journal pane (2)** — preview blank until file row selected; h collapses parent folder from file row
5. **KB pane (3)** — placeholder when cursor on folder; correct title on file row
6. **Templates** — `{{date:FORMAT}}` works (e.g. `{{date:Monday - January 02 2006}}`)
7. **Status bar** — shows skip count if any files fail to parse

---

## Known Issues

_(none logged — populate as testing reveals problems)_

---

## Deferred (out of scope for v1)

- AI / Agent integration
- Multi-device conflict resolution
- File watchers (`fsnotify`)
- Arrow key navigation (partially done — down/up keys work)
