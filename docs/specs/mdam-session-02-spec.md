# mdam — Testing Session 02 Specification

> **Session focus:** UX and UI. Navigation, layout, visual consistency, and one engine-level expansion (KB subtypes).
> **Stack:** Go, BubbleTea, lipgloss, glamour. No stack changes in this session.

---

## 0. Pane inventory change

Currently there are 6 panes. After this session there will be 4.

**Remove panes:**
- Pane 4 — TODOs (standalone pane)
- Pane 5 — Recent (standalone pane)

**Remaining panes (renumbered):**

| # | Name        |
|---|-------------|
| 1 | Dashboard   |
| 2 | Journal     |
| 3 | KB          |
| 4 | Tag Browser |

All tab bar indicators, keybindings, help menu text, and any internal routing must reflect 4 panes only.

---

## 1. Global — Tab Bar

### Problem

The user has no ambient awareness of which pane they are on or how many panes exist without opening the help menu.

### Current state

```
┌────────────────────────────────────────────────────────────────┐
│  Today — 2026-04-03                                            │
│                                                                │
│  Journal                                                       │
│    (no entry for today)                                        │
│  ...                                                           │
```

No tab indicator is visible at the top of the application.

### Required state

A persistent tab bar must be rendered at the top of every pane. It must show all 4 tabs. The active tab is visually distinct using **inverted colors** (foreground and background swapped — the same pattern tmux uses for the active window indicator in its status bar).

```
┌────────────────────────────────────────────────────────────────┐
│ [█ Dashboard █]  [ Journal ]  [ KB ]  [ Tag Browser ]         │
├────────────────────────────────────────────────────────────────┤
│  Today — 2026-04-03                                            │
│  ...                                                           │
```

**Specification:**
- Tab bar occupies one fixed line at the very top of the terminal.
- Each tab label is the pane name: `Dashboard`, `Journal`, `KB`, `Tag Browser`.
- Active tab: inverted colors (lipgloss `Reverse(true)`).
- Inactive tabs: normal text, no decoration.
- Tabs are separated by two spaces.
- Tab bar is always visible regardless of which pane is active.
- Navigating between panes with the existing pane-switch keybindings updates the active tab indicator.

---

## 2. Global — Focus Indicator

### Problem

Focus state is communicated inconsistently across panes. Some sections use indentation, some use `|>`, making it unclear where focus is at any time.

### Required behavior

Use a single consistent focus pattern throughout the application: **inverted colors on the focused row**, identical to how tmux highlights the active pane tab.

**Specification:**
- Any list row that has focus must be rendered with inverted colors (lipgloss `Reverse(true)`).
- This applies uniformly to: file lists, tag lists, document lists, and any other navigable list in the application.
- Remove all existing `|>` arrow indicators and indentation-based focus patterns.
- When a panel does not have focus (e.g. the right panel while the left panel has focus), no row in that panel should show the inverted highlight.

---

## 3. Global — Footer

### Problem

The footer shows `85 docs` — a single total count. The user cannot tell how many journal entries, KB docs, etc. exist without navigating to each pane.

### Current state

```
┌─ NORMAL ──┐  85 docs          /search  :cmd  f:filter  ?help  q:quit
```

### Required state

```
┌─ NORMAL ──┐  42 journal · 31 kb · 12 scratch    journals/2026-04-01.md    /search  :cmd  f:filter  ?help  q:quit
```

**Specification:**
- Replace the single doc count with a breakdown by document type (folder/type): e.g. `42 journal · 31 kb · 12 scratch`.
- Separator between counts is ` · ` (space, middle dot, space).
- To the right of the count breakdown, show the **relative path** of the currently highlighted file. Relative to the configured `base_dir`. Example: `journals/2026-04-01.md`.
- If no file is highlighted, omit the path segment.
- The keybinding hints remain on the far right, unchanged.

---

## 4. Global — Remove lazygit

All references to lazygit must be removed from the application.

**Remove:**
- The `ctrl+g` keybinding (or whichever key triggers the lazygit handoff).
- All lazygit mentions in the help menu (`?`).
- The `git.lazygit` config key handling (the config field can remain for forward-compatibility but the TUI must not expose or act on it).

The user will use lazygit independently outside of mdam. No replacement UI is needed.

---

## 5. Global — Open document with glamour (full-screen read mode)

### Problem

There is currently no way to read a document inside mdam without handing off to `$EDITOR`.

### Solution

Add a full-screen glamour read mode triggered by `o`. This uses the **glamour** library already compiled into mdam — no external `glow` binary is required or invoked.

**Specification:**

- Keybinding: `o` on any highlighted file opens the glamour read overlay.
- The overlay occupies the full terminal — no panels, no footer, no tab bar.
- Glamour renders the document using the style defined in `config.yml` (`theme:` field). Map the config theme name to the closest glamour style (e.g. `tokyonight` → glamour's `tokyo-night` style).
- Word wrap at terminal width (glamour `WordWrap: 0` equivalent — fill available width).
- The view is scrollable with `j`/`k` or arrow keys (use a BubbleTea viewport).
- `q` or `Escape` closes the overlay and returns the user to exactly where they were (same pane, same highlighted file, same scroll position).
- Add `o:read` to the help menu and footer hint line.

---

## 6. Dashboard — Pane 1

### Current layout

```
┌────────────────────────────────────────────────────────────────┐
│ Today — 2026-04-03                                             │
│                                                                │
│ Journal                                                        │
│   (no entry for today)                                         │
│                                                                │
│ TODOs (0 open)                                                 │
│   (no open tasks)                                              │
│                                                                │
│ Pinned (0)                                                     │
│   (pin docs with p)                                            │
│                                                                │
│ Recent                                                         │
│   04-02  Neovim and LazyVim                                    │
│   04-02  aexx.dev — Hetzner VPS Setup...                       │
│   ...                                                          │
└────────────────────────────────────────────────────────────────┘
```

### Required layout

The dashboard is split into two equal halves — left and right — divided by a vertical separator.

```
┌─ Dashboard ──────────────────────┬─────────────────────────────┐
│  Journal (last 5 days)           │  TODOs                      │
│    (no entry for today)          │    (coming soon)            │
│                                  │                             │
│  Pinned (0)                      │                             │
│    (pin docs with p)             │                             │
│                                  │                             │
│  Recent (last 20)                │                             │
│    04-02  Neovim and LazyVim     │                             │
│    04-02  aexx.dev — Hetzner...  │                             │
│    ...                           │                             │
└──────────────────────────────────┴─────────────────────────────┘
```

**Left side:**
- `Journal (last 5 days)` — the 5 most recent journal entries by date.
- `Pinned (max 20)` — pinned documents.
- `Recent (max 20)` — recently accessed/modified documents.

**Right side:**
- `TODOs` — placeholder panel. Display a static message such as `(coming soon)`. The panel is inert — no navigation, no interaction. This will be implemented in a future session.

**Navigation:**
- `j` / `k` moves the cursor up and down through the items in all left-side lists (Journal, Pinned, Recent are one navigable list).
- `h` / `l` moves focus between the left and right halves.
- When a file is highlighted, `o` opens it in glamour read mode (see §5).
- When a file is highlighted, `Enter` opens it in `$EDITOR`.
- The right half (TODOs placeholder) is not navigable.

---

## 7. Journal — Pane 2

### Current layout

```
┌─ Files ──────────────────┬─ Preview ──────────────────────────┐
│  2026-03-30.md           │                                    │
│  2026-04-01.md           │  --------                          │
│  2026-03-31.md           │                                    │
│  2026-03-27.md           │  ## type: journal                  │
│  2026-03-26.md           │  title: Monday - January 19 2026   │
│  ...                     │  tags: [onboarding, admin]         │
│                          │  created: 2026-01-19               │
│                          │  modified: 2026-03-25              │
│                          │                                    │
│                          │   Monday - January 19 2026        │
│                          │                                    │
│                          │  --------                          │
│                          │                                    │
│                          │  ## Hours                          │
│                          │  ...                               │
├──────────────────────────┴────────────────────────────────────┤
│ TODOs                                                          │
└────────────────────────────────────────────────────────────────┘
```

### Required layout

```
┌─ Files ──────────────────┬─ Preview ──────────────────────────┐
│  ▶ March - 2026          │  tags: [onboarding, admin]         │
│  ▼ April - 2026          │                                    │
│      2026-04-01.md       │  # Monday - January 19 2026        │
│    █ 2026-04-03.md █     │                                    │
│                          │  ---                               │
│                          │                                    │
│                          │  ## Hours                          │
│                          │  ...                               │
└──────────────────────────┴────────────────────────────────────┘
```

### Changes

**Remove TODOs section:** The TODOs panel at the bottom right must be removed entirely. This pane is two halves only: Files (left) and Preview (right).

**File tree — left side:**

- Entries are grouped into virtual month folders.
- Folder label format: `Month - YYYY` (e.g. `April - 2026`, `March - 2026`).
- Folders are sorted in reverse chronological order (most recent month at the top).
- All folders are **collapsed by default**.
- Only one folder can be expanded at a time. Opening a new folder collapses the previously open one.
- Collapsed folder indicator: `▶ April - 2026`
- Expanded folder indicator: `▼ April - 2026`
- `l` or `Enter` on a collapsed folder expands it. `h` or `Enter` on an expanded folder collapses it.
- When a folder is expanded, `j` / `k` moves between its entries.
- File entries within a folder are sorted in reverse chronological order (newest first).

**Preview — right side:**

The YAML frontmatter must **not** be rendered as raw text. Apply the following rules:

- **Strip** all frontmatter fields except `tags`.
- **Render only the tags line** at the top of the preview, e.g.: `tags: [onboarding, admin]`
- Then render the markdown body below the tags line, starting immediately — no blank lines before the first heading.
- The `---` separators and all other frontmatter fields (`type`, `title`, `created`, `modified`) must not appear in the preview.

**Expected preview output for a typical journal entry:**

```
tags: [sss, pos, admin, open-api]

# Monday - March 30 2026

---

## Hours
09:00 - 17:50
...
```

---

## 8. KB (Knowledge Base) — Pane 3

### Layout changes

Same layout changes as Journal (§7):
- Remove the TODOs panel from bottom right.
- Two halves only: Files (left) and Preview (right).
- Same preview frontmatter rendering rules as §7 (tags only, no raw YAML block, no leading whitespace).

### KB document subtypes

**Engine change:** Expand the type system for KB documents.

Currently the application identifies KB documents by `type: kb` in frontmatter. This must be extended to support subtypes.

**New rule:** Any document whose `type` value begins with `kb` is a KB document. This includes:
- `kb` (the existing base type)
- `kb_summary`
- `kb_domain`
- `kb_meeting`
- `kb_cars`
- `kb_ancient-history`
- Any `kb_*` value the user defines

The subtype is the string after the underscore. Display name is derived by replacing hyphens with spaces and title-casing. Examples:

| `type` value       | Display name      |
|--------------------|-------------------|
| `kb`               | KB (no subtype)   |
| `kb_summary`       | Summary           |
| `kb_domain`        | Domain            |
| `kb_meeting`       | Meeting           |
| `kb_cars`          | Cars              |
| `kb_ancient-history` | Ancient History |

**File tree — left side:**

Subtypes are grouped into virtual folders in the left panel, identical in behavior to the journal month folders (§7):
- Each subtype gets its own folder.
- `kb` (no subtype) documents go into a folder labeled `KB`.
- Folders are sorted alphabetically by display name.
- Folders are collapsed by default. One open at a time.
- Toggle behavior: `l` / `Enter` expands, `h` / `Enter` collapses.
- Subtype folders are **derived at runtime** by scanning the `type` frontmatter field of all files. No static configuration.

---

## 9. Tag Browser — Pane 4

### Current layout

```
┌─ Tags ───────────────────┬─ Documents ────────────────────────┐
│  pos (28)                │  Tuesday - March 31 2026           │
│  sss (26)                │  Wednesday - February 18 2026      │
│  authentigo (16)         │  Friday - February 27 2026         │
│  onboarding (15)         │  Friday - January 23 2026          │
│  admin (12)              │  Wednesday - February 4 2026       │
│  ...                     │  Thursday - March 19 2026          │
│                          │  Monday - March 16 2026            │
│                          │  Friday - March 13 2026            │
└──────────────────────────┴────────────────────────────────────┘
```

### Navigation bugs to fix

**Bug 1 — Focus not defaulting to the left panel on pane entry:**

When the user arrives at the Tag Browser from another pane where focus was on the right panel (preview or todos), the Tag Browser inherits that focus state. `hjkl` does nothing because focus is on the right panel which has no navigable content.

Fix: When the user navigates to the Tag Browser pane, **always reset focus to the left panel** (Tags list) regardless of which panel had focus on the previous pane.

**Bug 2 — Navigation breaks after pressing `h` or `l`:**

When the tags list is working and the user presses `h` or `l` to switch panels, all `hjkl` keys become unresponsive.

Fix: `h` and `l` must correctly transfer focus between the left (Tags) and right (Documents) panels without breaking the key handler. After pressing `l`, focus must be on the Documents panel with `j`/`k` responsive. After pressing `h`, focus must return to the Tags panel with `j`/`k` responsive.

### Required navigation behavior

| Key | Context | Action |
|-----|---------|--------|
| `j` / `k` | Focus on Tags (left) | Move up/down the tag list. Right panel updates to show documents for the highlighted tag. |
| `l` | Focus on Tags (left) | Move focus to Documents (right). |
| `h` | Focus on Documents (right) | Move focus back to Tags (left). |
| `j` / `k` | Focus on Documents (right) | Scroll up/down the document list for the selected tag. |
| `o` | Focus on Documents (right), file highlighted | Open highlighted document in glamour read mode (§5). |
| `Enter` | Focus on Documents (right), file highlighted | Open highlighted document in `$EDITOR`. |

---

## 10. Out of scope for this session

The following are noted but explicitly deferred:

- Deep TODO system implementation (engine, sweeps, persistence).
- Git integration beyond lazygit removal.
- Search refinement.
- Export feature.
- Terminal resize handling edge cases.
- Color theme loading from config (beyond existing behavior).
- The cursor aesthetic — how it looks is deferred; the inverted-color behavior specified in §2 is the implementation target for this session.
