# mdam — Project Handoff

mdam (Markdown Admin Management) is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, TODOs, scratch notes.

**Current branch:** `fix/ui-ux-updates`

---

## Session 02 — COMPLETE

All 9 work units of the mdam-session-02-spec have been implemented and all tests pass (`go test ./... && go vet ./...` both green).

### What was done

| WU | Description | Status |
|----|-------------|--------|
| WU1 | Structural cleanup — removed ViewAll/ViewTodo/ViewRecent/PanelTodo/SmartFilter/lazygit; 4 named views (Dashboard/Journal/KB/Tags); Tab/ShiftTab cycle panes | DONE |
| WU2 | Persistent tab bar + `contentHeight()` helper | DONE |
| WU3 | Full-row Reverse(true) focus indicator (file panel, tag panel) | DONE |
| WU4 | Footer doc count breakdown (journal/kb/scratch) + highlighted file path | DONE |
| WU9 | Tag Browser: `tagDocCursor` for right panel; j/k navigation when PanelPreview active; enter opens tagged doc | DONE |
| WU7 | Journal tree view: month-folder collapse/expand (one open at a time); auto-expand current month on entry; `journalCursor` navigation; `tui/view_journal.go` | DONE |
| WU8 | KB subtypes + tree: `kbSubtype()` derives folder labels from type prefix; `filterKBDocs()` uses `HasPrefix("kb")`; `kbCursor` navigation; `tui/view_kb.go` | DONE |
| WU5 | Glamour read mode: `o` opens full-screen ModeRead overlay; q/Esc closes; `stripFrontmatter()`; `cmdLoadRead()`; scrollable viewport | DONE |
| WU6 | Dashboard two-column: navigable left (journal/pinned/recent); static right (TODOs); `dashCursor`/`dashRight`; `buildDashItems()` with dedup | DONE |

### New files
- `tui/view_journal.go` — journalRow, buildJournalRows, renderJournalView, initJournalView
- `tui/view_kb.go` — kbRow, kbSubtype, filterKBDocs, buildKBRows, renderKBView
- `tui/view_journal_test.go` — journal tree tests
- `tui/view_kb_test.go` — KB subtype tests

### Key model fields added
```
journalExpanded  map[string]bool
journalCursor    int
kbExpanded       map[string]bool
kbCursor         int
tagDocCursor     int
dashCursor       int
dashRight        bool
readViewport     viewport.Model
readReturnView   View
readReturnPanel  PanelID
```

---

## Next Session: Start Here

Branch `fix/ui-ux-updates` needs to be reviewed, then merged to master via PR.

After merge, start the systematic testing pass (v0.2.0 gate):

1. Manual smoke test all 4 panes: Dashboard → Journal → KB → Tags
2. Verify Journal auto-expand with real data; test tree expand/collapse
3. Verify KB subtype folders with real `kb_*` typed docs
4. Test Tag Browser left/right panel navigation and tagDocCursor
5. Test read mode (`o`) on a real doc; verify scrolling; q/Esc closes
6. Test Dashboard navigation; verify pinned/recent dedup
7. If all smoke tests pass, consider bumping to v0.1.1

---

## Known Issues

_(none logged — populate as testing reveals problems)_

---

## Deferred (out of scope for v1)

- AI / Agent integration
- Multi-device conflict resolution
- File watchers (`fsnotify`)
- Arrow key navigation (partially done — down/up keys work)
