# TUI-UX.md — mdam Interface Standards

This document defines the UX rules for the mdam TUI. It exists to give Claude Code
inherent design judgment so that every UI change, new view, or component produces
consistent, readable, low-cognitive-load output — without needing per-task guidance.

The north star: **information should be instantly scannable, focus should be
unambiguous, and the interface should never make the user think about the UI itself.**

Reference applications: lazygit, k9s, atac, yazi, glow.
Charm stack: Bubble Tea, Lip Gloss, Bubbles, Glamour.

---

## 1. Core Principles

### 1.1 Hierarchy before decoration
Every screen has exactly one primary element, one secondary element, and everything
else is tertiary. Never style two things the same way if one is more important than
the other. Decoration that does not carry information must be removed.

### 1.2 Focus is sacred
The user must always be able to answer "where am I?" in under one second without
reading anything. Cursor position, active panel, and active pane are always visually
unambiguous. When in doubt, make the focus indicator stronger, not weaker.

### 1.3 Empty states are not an afterthought
Every list, tree, and panel that can be empty must render a placeholder. Blank space
is not a valid empty state. A blank screen communicates failure, not emptiness.

### 1.4 Breathing room is load-bearing
Margins and padding exist to group related items and separate unrelated ones.
Tightly packed content forces the eye to work harder. Every panel has internal
padding. Related items share spacing. Unrelated sections have visible separation.

### 1.5 Muted text earns its keep
Dim/muted styling is reserved for genuinely secondary information (dates, counts,
paths, hints). If something is muted it must be less important than everything around
it. Never mute the thing the cursor is on.

---

## 2. Spacing and Layout

### 2.1 Panel internal padding
All bordered or visually delimited panels use **1 cell of horizontal padding** on
each side. Do not render content flush against a border or the terminal edge.

```
// correct
style := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)

// wrong — content pressed against border
style := lipgloss.NewStyle()
```

### 2.2 Section separation
When a panel contains multiple logical sections (e.g. Dashboard: Journal / Pinned /
Recent), separate them with a **blank line** between the last item of one section and
the header of the next. Never separate with a horizontal rule unless it is a named
border in the theme.

### 2.3 Section headers
Section headers inside a panel use `Bold(true)` and are followed by a blank line
before the first item. They are never the same style as body rows.

```
// header row style
theme.SectionHeader = lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)

// never this — header indistinguishable from body
theme.SectionHeader = lipgloss.NewStyle()
```

### 2.4 Indentation in trees
Tree children (files under a folder) are indented by **2 spaces** relative to their
parent. Folder rows themselves are not indented.

```
▶ 2026-03        ← folder, no indent
  2026-03-15     ← child, 2-space indent
  2026-03-14
▸ 2026-02
```

### 2.5 Status bar spacing
The status bar is a single line. Fields within it are separated by the theme
separator glyph with a space on each side (` │ `). Never concatenate fields without
a delimiter.

---

## 3. Focus and Cursor

### 3.1 Active cursor row
The selected row in any list or tree uses **full-width reverse video** (`Reverse(true)`
on a style that spans the full panel width). This is non-negotiable. Subtle underline
or color-only indicators are insufficient in a terminal environment.

```go
// correct
rowStyle = rowStyle.Reverse(true).Width(availableWidth)

// wrong — color change only, loses contrast on many terminals
rowStyle = rowStyle.Foreground(theme.Accent)
```

### 3.2 Active panel indicator
The active panel's border or header uses `theme.Accent` color. Inactive panels use
`theme.Subtle`. This must be applied consistently across all panels in all panes.

### 3.3 Active pane tab
The active tab in the tab bar uses `Bold(true)` and `theme.Accent` foreground.
Inactive tabs use `theme.Subtle`. The tab bar always renders — it is never hidden.

### 3.4 Focus never disappears
When switching panes or panels, the cursor position in the newly focused panel must
be visible immediately. If re-entering a pane, restore the last cursor position.
Never land the user on a panel with no visual cursor.

### 3.5 Read mode
Read mode is a full-screen overlay. There is no panel border, no status bar section
ambiguity. A single-line header shows the document title. The mode indicator in the
footer changes to `READ`. The cursor is the viewport scroll position, which must
always be visibly tracked (Bubbles viewport handles this).

---

## 4. Typography and Text Hierarchy

Four levels. Use exactly these — do not invent new ones.

| Level | Usage | Style |
|---|---|---|
| **Accent** | Active tab, active border, cursor highlight | `theme.Accent` fg, Bold |
| **Primary** | File names, tag names, document titles, interactive items | default fg |
| **Secondary** | Dates, counts, type labels, git status | `theme.Subtle` fg |
| **Muted** | Placeholder text, hints, disabled states | `theme.Muted` fg, optionally Italic |

Rules:
- Never use more than two text styles on the same row.
- The cursor row always renders in Reverse — do not additionally Bold the text inside it.
- Paths shown in the status bar are Secondary.
- Counts in brackets (e.g. `[3]`) are Secondary.

---

## 5. Empty States

Every view that renders a list, tree, or panel must handle the empty case explicitly.

### 5.1 Required empty states

| Panel / View | Empty condition | Placeholder text |
|---|---|---|
| Journal tree | No journal entries | `No journal entries.` |
| KB tree | No KB documents | `No knowledge base documents.` |
| Tag browser (left) | No tags | `No tags found.` |
| Tag browser (right) | Tag has no docs / no tag selected | `Select a tag to see documents.` |
| Dashboard — Journal section | No journal entries | `No recent journal entries.` |
| Dashboard — Pinned section | No pins | `No pinned documents.` |
| Dashboard — Recent section | No recent docs | `No recent documents.` |
| Dashboard — TODOs | No open tasks | `No open tasks.` |
| Preview panel | No document selected | `Select a document to preview.` |
| Search results | Query returns nothing | `No results for "{query}".` |

### 5.2 Placeholder style
Placeholder text uses the **Muted** style and is rendered at the top of the content
area with 1 line of top padding. It is never centered vertically. It is never blank.

```go
placeholder := theme.Muted.Render("No journal entries.")
content = lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(placeholder)
```

---

## 6. Status Bar

The status bar is a single line at the bottom of the screen. It has three zones:

```
 NORMAL │ main ↑2 │ 3 journal · 1 kb · 0 scratch   /  :  o:read  ?  q
 ←left─────────────────────────────────────────────right→
```

- **Left zone**: mode indicator, git branch + status. Mode uses `Bold`.
- **Centre zone**: document counts, rendered in Secondary style.
- **Right zone**: key hints, always rendered in Muted style.

Rules:
- Key hints on the right show only the most context-relevant bindings (max 5).
- The mode indicator changes for every mode: `NORMAL`, `READ`, `SEARCH`, `COMMAND`, `DELETE?`.
- Never truncate the mode indicator. Truncate key hints first if space is tight.
- The status bar is always exactly 1 line. Never wrap it.

---

## 7. Confirmations and Overlays

### 7.1 Delete confirmation
Delete confirmation renders inline in the status bar area, not as a floating modal.
Format: `Delete "{title}"? (y/n)` in `theme.Warning` color.

### 7.2 Help overlay
The help overlay is a centered box rendered over the current view. It must:
- Have a visible border in `theme.Accent`.
- Show a title: `Keybindings`.
- Group bindings by section (Navigation / Actions / Application).
- Use Secondary style for key names, Primary for descriptions.
- Close on `?` or `Esc`.

### 7.3 Command mode
Command mode renders an input line in the status bar area, prefixed with `:` in
`theme.Accent`. The rest of the status bar content is hidden while command mode is
active.

---

## 8. Preview Panel

The right-side preview panel renders glamour markdown output for the selected
document. Rules:

- Frontmatter is always stripped before rendering.
- If no document is selected, show the placeholder (see §5).
- The panel title (top border label) shows the document title, not the filename.
- The panel does not scroll independently in normal navigation — it snaps to top
  when selection changes. Scroll is only available in Read mode (`o`).
- Glamour style should match `theme.Name` where a matching glamour style exists
  (`tokyonight` → `tokyo-night`, `dracula` → `dracula`, others → `dark`).

---

## 9. Trees (Journal and KB)

### 9.1 Folder row anatomy
```
▶ 2026-03   [4]
```
- Glyph: `▶` collapsed, `▼` expanded (or configured icon from `icons.go`).
- Folder name: Primary style.
- Count in brackets: Secondary style, right-aligned within the label — not padded
  to full width.

### 9.2 File row anatomy
```
  2026-03-15   ✦ pinned · modified
```
- Two-space indent.
- Filename or title: Primary.
- Status markers (pinned, git modified/untracked): Secondary, separated by ` · `.
- Markers appear after the name, never before.

### 9.3 One folder open at a time
Only one folder may be expanded at a time. Opening a new folder collapses the
previous one. This is already implemented — maintain it.

---

## 10. Dashboard

Two columns. Left is navigable. Right is static (TODO display only).

### 10.1 Left column sections (in order)
1. **Journal** — most recent 5 journal entries.
2. **Pinned** — all pinned documents.
3. **Recent** — most recently modified KB/scratch docs, deduplicated against above.

Each section has a Bold section header followed by a blank line. Section headers are
skipped by `j`/`k` navigation (cursor lands only on document rows).

### 10.2 Right column — TODOs
Displays open tasks from `todo/todo.md`. Tasks are grouped by priority:
`!high` first, then `!medium`, then `!low`, then unprioritised.

Priority label rows use Secondary style. Task rows use Primary. Completed marker `✓`
uses Muted.

### 10.3 Column divider
A vertical `│` separator in `theme.Subtle` divides the two columns. It spans the
full content height, not just the text rows.

---

## 11. What the Agent Must Never Do

- **Never** render a list or tree without an empty state placeholder.
- **Never** use color alone to indicate focus — always pair with Reverse or Bold.
- **Never** render content flush against a border without horizontal padding.
- **Never** use the same text style for a section header and a body row.
- **Never** truncate the mode indicator in the status bar.
- **Never** add a new panel or view without a defined empty state.
- **Never** hardcode colors — always use `theme.*` fields.
- **Never** invent a fifth text hierarchy level — use Muted for anything that does
  not fit the four defined levels.
- **Never** render two panels with identical border colors — active must be visually
  distinct from inactive.
