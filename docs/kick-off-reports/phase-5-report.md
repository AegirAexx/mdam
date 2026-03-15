# Phase 5 Report — Polish

**Completed:** 2026-03-14
**Baseline:** 64 TUI tests, Phases 1–4 complete, fully functional but unstyled TUI
**Result:** 119 TUI tests, production-quality visual design, all packages clean

---

## What Was Done

### Dependency

Added `github.com/charmbracelet/glamour` (pre-approved in spec §5) for async markdown preview rendering. All other dependencies were already present.

---

### Color Theming (`tui/theme.go`)

`Theme` struct holds all lipgloss styles computed once per palette in `New()` — never re-computed per frame. `NewTheme(name string) Theme` dispatches to one of five palette constructors:

| Name | Base | Glamour style |
|---|---|---|
| `tokyonight` (default) | `#1a1b26` background, blue/cyan/green accents | `"dark"` |
| `nord` | `#2e3440` background, nordic blue/gray | `"notty"` |
| `gruvbox` | `#282828` background, warm earth tones | `"dark"` |
| `catppuccin` | `#1e1e2e` background, pastel accents | `"dark"` |
| `dracula` | `#282a36` background, purple/pink | `"dark"` |

Unknown theme names fall back to tokyonight. The `GlamourStyle` field on `Theme` is passed directly to `glamour.Render` so the markdown preview palette matches the UI palette.

Styles cover: mode indicator (status bar), git markers, file list selection, panel headers, preview metadata, TODO items, smart filter bar, dashboard sections.

---

### Nerd Font Icons (`tui/icons.go`)

`Icons` struct with two variants selected at startup from `cfg.NerdFonts`:

- **`DefaultIcons()`** — Nerd Font glyphs (journal, KB, git markers, pin star, etc.)
- **`PlainIcons()`** — ASCII fallback (`[J]`, `[M]`, `[*]`, `> `, etc.) that works on any terminal

`NerdFonts` defaults to `false` in config so the tool works out of the box without a patched font.

---

### Pin/Unpin (`tui/pins.go`, key `p`)

Pinned document paths are stored as a sorted JSON array at `~/.config/mdam/pins.json` (via `cfg.PinsPath()`), alongside the config file rather than inside `BaseDir` so pins survive directory reconfigurations.

Three pure functions:
- `loadPins(path)` — returns empty map (not an error) if file absent
- `savePins(path, pins)` — serialises sorted slice, writes atomically
- `togglePin(pins, path)` — returns new map, never mutates the original

Flow: `p` key → `togglePin` → update `m.pinnedPaths` → `cmdSavePins` (async, errors silently dropped — pins are best-effort). `cmdLoadPins` fires in `Init()` alongside the document scan. Pinned documents are highlighted with `theme.FilePinned` style and the pin icon in all file list views, and sorted to the top of the dashboard.

---

### Delete with Confirmation (`tui/delete.go`, key `d`)

New `ModeDeleteConfirm` mode. Flow:

1. `d` on a selected document → `ModeDeleteConfirm`, `deleteConfirmPath` set, status bar shows `"Delete <filename>? (y/n)"`
2. `y` → `cmdDeleteDoc` (calls `os.Remove`) → `deleteDoneMsg` → reload docs, clear path, show `"deleted <filename>"`
3. `n` or `Esc` → `ModeNormal`, path cleared, no side effects

`deleteDoneMsg` carries both `path` and `err`. On error the status message is `"delete failed: …"`.

---

### Smart Filter (`tui/model.go`, key `f`)

`SmartFilter` int enum with four constants. `f` cycles through them in order:

| Filter | Condition |
|---|---|
| `SmartFilterNone` | No filter — all documents |
| `SmartFilterUntagged` | `len(tags) == 0` |
| `SmartFilterStaleWeek` | `Modified` before 7 days ago |
| `SmartFilterInbox` | `type: unsorted` |

`applySmartFilter` is a pure function applied inside `visibleDocs()` only when `activeView == ViewAll`. A tinted filter bar appears below the panel header when a filter is active, showing the filter name. Pressing `f` a fourth time wraps back to `SmartFilterNone` and clears the bar.

---

### Tag Browser (`tui/view_tags.go`, key `6`)

`buildTagIndex(docs []search.Result) []tagEntry` aggregates all tags from the document list, returning entries sorted by count descending then name ascending for ties.

`key 6` → `ViewTags` + `cmdBuildTagIndex(m.docs)` → `tagIndexMsg` → `m.tagEntries`. The tag browser replaces the normal panel layout:

- **Left panel**: tag list with count, cursor highlighting, `tagCursor` navigation
- **Right panel**: documents carrying the currently selected tag (title list)
- Pressing `Enter` on a tag fires `cmdSearch` with that tag as a filter

---

### Dashboard (`tui/view_dashboard.go`, key `1`)

Key `1` now maps to `ViewDashboard` (was `ViewAll` in Phase 4). `ViewAll` is the startup default and remains accessible without a key.

`renderDashboard()` shows four sections in a single-panel layout:
1. **Today** — today's date, link to today's journal entry (if it exists in docs)
2. **TODOs** — open task count + up to 5 task previews
3. **Pinned** — pinned documents with pin icon; `"(pin docs with p)"` when empty
4. **Recent** — up to 5 documents sorted by `Modified` descending

All sections use `theme.DashboardHeader` and `theme.DashboardItem` styles. The status bar is rendered normally at the bottom.

---

### Glamour Markdown Preview (`tui/commands.go`, `tui/model.go`)

`cmdLoadPreview(path, glamourStyle, width)` runs in a goroutine: reads the file, calls `glamour.Render`, sends `previewReadyMsg`. On glamour failure it falls back to raw file content.

`preview viewport.Model` is stored on `Model`. `WindowSizeMsg` resizes the viewport to match the current right panel dimensions. `previewReadyMsg` calls `m.preview.SetContent()`.

When the viewport has content (`TotalLineCount() > 0`), `renderPreviewPanel` shows the viewport output. When the viewport is empty (initial state, before the first async load completes), it falls back to the styled frontmatter metadata display from Phase 3.

The preview is re-loaded whenever the file cursor moves.

---

### Spinner (`tui/model.go`, `tui/commands.go`)

`cmdTick()` returns a `tea.Tick(100ms)` command that sends `tickMsg`. On `tickMsg`, `spinnerFrame` advances through `spinnerFrames` (10 braille frames). While `m.loading` is true, each `tickMsg` re-schedules the next tick. When loading completes the chain stops.

The current frame is rendered in the loading state of the file panel and in the status bar scanning indicator.

---

### `view.go` Lipgloss Rewrite

`view.go` was fully rewritten. Key changes from Phase 4's plain-text rendering:

- `padRight` now uses `lipgloss.Width()` (ANSI-aware) instead of `len()` so styled strings pad correctly
- `panelHeader` (plain, for tests) kept alongside new `styledPanelHeader` (lipgloss, for rendering)
- `gitMarkerStyled` returns coloured, icon-based markers; `gitMarkerFor`/`gitMarkerForStatus` kept for test assertions
- Status bar mode indicator uses palette-matched background colour per mode
- File list shows: cursor glyph, pin icon, styled filename, coloured git marker — all with proper visual-width accounting
- `View()` dispatches to `renderDashboard()` or `renderTagBrowser()` for their respective view types before the normal two-panel layout

---

### Config additions (`internal/config/config.go`)

```go
NerdFonts bool `mapstructure:"nerd_fonts"` // default: false
```

```go
func (c Config) PinsPath() string // → ~/.config/mdam/pins.json
```

---

### Test Suite

`model_test.go` gained `stripANSI(s string) string` (regex strips `\x1b[...m` sequences) and all `View()` string assertions were updated to use it. This makes tests terminal-agnostic — they pass whether or not lipgloss emits colour codes.

**Test counts:**

| Package | Before | After |
|---|---|---|
| `tui` | 64 | 119 |
| All packages | 64 TUI + rest | 119 TUI + rest |

New test categories: theme smoke tests, icon completeness, pin round-trip, delete confirm flow, smart filter cycling, tag index aggregation, dashboard rendering, viewport resize, spinner behaviour, ANSI stripping.

---

## Files Changed

### New
- `tui/theme.go` + `tui/theme_test.go`
- `tui/icons.go` + `tui/icons_test.go`
- `tui/pins.go` + `tui/pins_test.go`
- `tui/delete.go`
- `tui/view_dashboard.go`
- `tui/view_tags.go` + `tui/view_tags_test.go`

### Modified
- `go.mod` / `go.sum` — added `glamour`
- `internal/config/config.go` + `config_test.go` — `NerdFonts`, `PinsPath()`
- `tui/mode.go` — `ModeDeleteConfirm`, `ViewDashboard`, `ViewTags`, `SmartFilter`
- `tui/keys.go` — `6`, `p`, `f`, `y`, `DeleteConfirm`, `Pin`, `SmartFilter` bindings
- `tui/messages.go` — `previewReadyMsg`, `pinsLoadedMsg`, `tagIndexMsg`, `tickMsg`
- `tui/commands.go` — `cmdLoadPreview`, `cmdLoadPins`, `cmdSavePins`, `cmdBuildTagIndex`, `cmdTick`
- `tui/model.go` — new fields, `Init`, `Update`, all handlers, `visibleDocs`, `applySmartFilter`
- `tui/view.go` — full lipgloss rewrite, viewport, spinner, smart filter bar
- `tui/model_test.go` — `stripANSI`, updated assertions, 55 new tests
- `docs/KEYBINDINGS.md` — Phase 5 complete

---

## Decisions Made

- **`ctrl+g` stays** for lazygit — `g` conflicts with `gg` chord. Deferred to real usage data post-Phase 5.
- **`1` → ViewDashboard** (spec §4.2). ViewAll is the startup default; no key activates it directly.
- **Nerd Fonts default false** — ASCII fallback ensures the binary works on any terminal.
- **Pins best-effort** — `cmdSavePins` errors are silently dropped; pin state is cosmetic, not critical.
- **Smart filter ViewAll only** — other views already filter by type; stacking filters would produce confusing empty states.
- **Viewport falls back to frontmatter metadata** when glamour content hasn't loaded yet — no blank preview on startup.
