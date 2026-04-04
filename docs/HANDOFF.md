# mdam — Project Handoff

mdam (Markdown Admin Management) is a keyboard-driven terminal TUI for managing a personal markdown document tree — journals, knowledge base, TODOs, scratch notes.

**Current branch:** `fix/ui-ux-refactoring`

---

## Session 03 — COMPLETE

TUI-UX compliance refactoring pass. All 13 gaps between the existing implementation and `docs/TUI-UX.md` have been addressed. All tests pass (`go test ./...` and `go vet ./...` both green).

### What was done

| Gap | Description | Status |
|----|-------------|--------|
| 1 | Theme: Added `Accent`, `Subtle`, `Muted`, `Warning` semantic fields to all 5 palettes | DONE |
| 2 | Tab bar: Changed `TabActive` from `Reverse(true)` to `Bold+Accent fg` per §3.3 | DONE |
| 3 | Read mode: Added document title header; footer now uses `renderStatusBar()` showing `READ` | DONE |
| 4 | Preview panel: Dynamic title from `Frontmatter.Title`; empty state "Select a document to preview." | DONE |
| 5 | Help overlay: Centered box with `RoundedBorder` in Accent color via `lipgloss.Place` | DONE |
| 6 | Status bar hints: Tightened to `/  :  o:read  ?  q` using `m.theme.Muted` | DONE |
| 7 | Delete confirm: Status bar right zone shows `Delete "title"? (y/n)` in Warning color | DONE |
| 8 | Dashboard: Blank rows between sections; per-section empty states (§2.2/2.3/5) | DONE |
| 9 | Dashboard TODOs: Priority-grouped display (!high→!medium→!low→unprioritised) | DONE |
| 10 | Journal folders: `[N]` count in Subtle style outside Reverse block; YYYY-MM labels | DONE |
| 11 | KB folders: Same count treatment as journal | DONE |
| 12 | Tags: Hardcoded focused=true bug fixed; all empty states use Muted + spec text | DONE |
| 13 | Model: Added `readDocTitle`/`deleteConfirmTitle` fields; cursor skips blanks/placeholders | DONE |

### Files modified
- `tui/theme.go` — Accent/Subtle/Muted/Warning fields + TabActive updated in all 5 palettes
- `tui/model.go` — 2 new fields, `selectedDocTitle()` helper, o/d handlers updated, cursor skip logic
- `tui/view.go` — Tab bar, read mode, preview title, status bar, viewHelp box
- `tui/view_dashboard.go` — isBlank/isPlaceholder on dashItem, section structure, priority TODOs
- `tui/view_journal.go` — icon/count fields on journalRow, YYYY-MM folder labels, empty state
- `tui/view_kb.go` — icon/count fields on kbRow, empty state
- `tui/view_tags.go` — focused bug fix, Muted empty states
- `tui/model_test.go` — TestBuildDashItems updated to skip blank/placeholder rows

---

## Next Session: Start Here

Branch `fix/ui-ux-refactoring` is ready for review and PR to master.

After merge, manual smoke test all 4 panes with real document data:
1. **Dashboard** — verify blank lines between sections; Journal/Pinned/Recent placeholders when empty; TODO priority groups
2. **Journal** — verify YYYY-MM folder labels with `[N]` counts; empty state text
3. **KB** — verify subtype folder labels with `[N]` counts; empty state text
4. **Tag Browser** — verify focus resets to Tags on entry; empty state texts; Documents panel focus bug fixed
5. **Read mode** — verify document title header at top; READ mode in status bar; `q` closes properly
6. **Help** — verify centered bordered box with Accent color
7. **Delete** — verify `Delete "title"? (y/n)` in Warning color in status bar

---

## Known Issues

_(none logged — populate as testing reveals problems)_

---

## Deferred (out of scope for v1)

- AI / Agent integration
- Multi-device conflict resolution
- File watchers (`fsnotify`)
- Arrow key navigation (partially done — down/up keys work)
