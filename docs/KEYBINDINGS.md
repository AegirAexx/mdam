# mdam — Keybinding Reference

## Modes

| Mode | Purpose | Activation | Exit |
|---|---|---|---|
| Normal | Navigate, browse, trigger actions | Default | — |
| Read | Full-screen glamour document viewer | `o` on a doc | `q` / `Esc` |
| Command | Execute colon-prefixed commands | `:` | `Enter` / `Esc` |

---

## Normal Mode

### Navigation

| Key | Action |
|---|---|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` | Switch to left panel; collapse tree folder if on a folder row |
| `l` / `→` | Switch to right panel; expand tree folder if on a folder row |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `Tab` | Cycle to next pane |
| `Shift+Tab` | Cycle to previous pane |

### Panes

| Key | Pane |
|---|---|
| `1` | Dashboard |
| `2` | Journal (month-folder tree) |
| `3` | KB (subtype-folder tree) |
| `4` | Tag Browser (with substring filter) |
| `5` / `/` | Search |

### Document Actions

| Key | Action |
|---|---|
| `Enter` | Open selected document in `$EDITOR`; on Tag Browser / Search left panel, activates filter input |
| `o` | Open selected document in read mode |
| `s` | Open scratch pad in `$EDITOR` |
| `t` | Open todo in `$EDITOR` |
| `n` | New document (template picker: journal or kb) |
| `e` | Export selected document (strip frontmatter) |
| `p` | Pin / unpin selected document |
| `R` | Force re-scan directory tree and git status |
| `Esc` | Clear filter (Tag Browser) or search results (Search pane) |

### Application

| Key | Action |
|---|---|
| `?` | Toggle keybinding help overlay |
| `:` | Command mode |
| `q` | Quit mdam |

---

## Read Mode

Entered with `o` on any document. Full-screen glamour-rendered view, frontmatter stripped.

| Key | Action |
|---|---|
| `j` / `↓` | Scroll down one line |
| `k` / `↑` | Scroll up one line |
| `d` | Scroll down half a page |
| `u` | Scroll up half a page |
| `f` | Scroll down one page |
| `b` | Scroll up one page |
| `g` | Go to top |
| `G` | Go to bottom |
| `q` / `Esc` | Close and return to previous pane |

---

## Command Mode

Entered via `:`. Commands execute on `Enter`, cancel with `Esc`.

| Command | Action |
|---|---|
| `:q` / `:quit` | Quit |

---

## Journal Pane (key `2`)

The left panel shows a month-folder tree. Only one month can be expanded at a time.

| Key | Action |
|---|---|
| `j` / `k` | Move cursor through folders and files |
| `l` on folder | Expand |
| `h` on folder | Collapse |
| `h` on file | Collapse parent folder |
| `l` on file | No-op (folder already open) |
| `Enter` on file | Open in `$EDITOR` |
| `o` on file | Open in read mode |

The current month is auto-expanded and the cursor lands on the most recent entry when entering this pane. If the current month has no entries, the most recent past month is expanded instead.

---

## KB Pane (key `3`)

The left panel shows a subtype-folder tree derived from the `type` field (e.g. `kb_summary` -> "Summary" folder).

| Key | Action |
|---|---|
| `j` / `k` | Move cursor |
| `l` on folder | Expand |
| `h` on folder | Collapse |
| `h` on file | Collapse parent folder |
| `l` on file | No-op (folder already open) |
| `Enter` on file | Open in `$EDITOR` |
| `o` on file | Open in read mode |

---

## Tag Browser (key `4`)

The left panel lists all tags sorted alphabetically. The right panel shows documents carrying the selected tag.

| Key | Panel | Action |
|---|---|---|
| `Enter` | Left | Activate the filter input — type to narrow the tag list by substring |
| `Enter` | Left (filter active) | Deactivate filter input, keep filtered list, navigate with `j`/`k` |
| `Esc` | Left | Clear the filter, restore full tag list |
| `j` / `k` | Left | Navigate tags |
| `l` / `h` | — | Switch between tag list and document list |
| `j` / `k` | Right | Navigate documents for selected tag |
| `Enter` | Right | Open highlighted document in `$EDITOR` |
| `o` | Right | Open highlighted document in read mode |

---

## Search Pane (key `5` or `/`)

Two-column layout. Left panel has a search input and three category entries (Journal, KB, Tags). Right panel shows documents for the selected category.

| Key | Panel | Action |
|---|---|---|
| `Enter` | Left | Activate the search input — type a query, press Enter to search |
| `Enter` | Left (input active) | Execute search, deactivate input |
| `Esc` | Left (input active) | Deactivate input without searching |
| `Esc` | — | Clear search results and reset to empty state |
| `j` / `k` | Left | Navigate categories (Journal, KB, Tags) |
| `l` / `h` | — | Switch between category list and document list |
| `j` / `k` | Right | Navigate documents in selected category |
| `Enter` | Right | Open highlighted document in `$EDITOR` |
| `o` | Right | Open highlighted document in read mode |
| `p` | Right | Pin / unpin highlighted document |

---

## Dashboard (key `1`)

Two-column view. Left column is navigable (journal / pinned / recent). Right column shows a glamour-rendered preview of `todo.md`.

| Key | Action |
|---|---|
| `j` / `k` | Navigate left column (skips section headers) |
| `l` | Switch focus to right (Todo) column |
| `h` | Switch focus back to left column |
| `Enter` | Left: open selected document in `$EDITOR`; Right: open `todo.md` in `$EDITOR` |
| `o` | Left: open selected document in read mode; Right: open `todo.md` in read mode |
| `p` | Pin / unpin selected document (max 10 pins, oldest evicted) |
