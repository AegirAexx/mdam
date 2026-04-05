# mdam — Keybinding Reference

## Modes

| Mode | Purpose | Activation | Exit |
|---|---|---|---|
| Normal | Navigate, browse, trigger actions | Default | — |
| Read | Full-screen glamour document viewer | `o` on a doc | `q` / `Esc` |
| Command | Execute colon-prefixed commands | `:` | `Enter` / `Esc` |
| Search | Fuzzy find across the document tree | `/` | `Enter` / `Esc` |
| Delete? | Confirm or cancel a document deletion | `d` on a doc | `y` / `n` / `Esc` |

---

## Normal Mode

### Navigation

| Key | Action |
|---|---|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` | Switch to left panel; collapse tree folder if on a folder row |
| `l` / `→` | Switch to right panel; expand tree folder if on a folder row |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `Tab` | Cycle to next pane |
| `Shift+Tab` | Cycle to previous pane |

### Panes

| Key | Pane |
|---|---|
| `1` | Dashboard |
| `2` | Journal (month-folder tree) |
| `3` | KB (subtype-folder tree) |
| `4` | Tag Browser |

### Document Actions

| Key | Action |
|---|---|
| `Enter` | Open selected document in `$EDITOR` |
| `o` | Open selected document in read mode |
| `s` | Open scratch pad in `$EDITOR` |
| `n` | New document (template picker) |
| `d` | Delete selected document (prompts for confirmation) |
| `e` | Export selected document (strip frontmatter) |
| `p` | Pin / unpin selected document |
| `R` | Force re-scan directory tree |

### Search

| Key | Action |
|---|---|
| `/` | Enter search mode |

### Application

| Key | Action |
|---|---|
| `?` | Toggle keybinding help overlay |
| `q` | Quit mdam |

---

## Read Mode

Entered with `o` on any document. Full-screen glamour-rendered view, frontmatter stripped.

| Key | Action |
|---|---|
| `j` / `↓` | Scroll down one line |
| `k` / `↑` | Scroll up one line |
| `Space` | Scroll down one page |
| `b` | Scroll up one page |
| `q` / `Esc` | Close and return to previous pane |

---

## Delete Confirmation Mode

| Key | Action |
|---|---|
| `y` | Confirm delete |
| `n` / `Esc` | Cancel |

---

## Command Mode

Entered via `:`. Commands execute on `Enter`, cancel with `Esc`.

| Command | Action |
|---|---|
| `:q` / `:quit` | Quit |
| `:todo sweep` | Run TODO sweep manually |
| `:todo archive` | Archive old completed tasks |

---

## Search Mode

| Key | Action |
|---|---|
| `Enter` | Confirm search, return to normal with results |
| `Esc` | Cancel search |

After searching, use `j`/`k` to navigate results and `Enter` to open.

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

The left panel shows a subtype-folder tree derived from the `type` field (e.g. `kb_summary` → "Summary" folder).

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

The left panel lists all tags sorted by document count. The right panel shows documents carrying the selected tag.

| Key | Panel | Action |
|---|---|---|
| `j` / `k` | Left | Navigate tags |
| `l` / `h` | — | Switch between tag list and document list |
| `j` / `k` | Right | Navigate documents for selected tag |
| `Enter` | Right | Open highlighted document in `$EDITOR` |

---

## Dashboard (key `1`)

Two-column view. Left column is navigable (journal / pinned / recent). Right column shows open TODOs (display only).

| Key | Action |
|---|---|
| `j` / `k` | Navigate left column (skips section headers) |
| `l` | Switch focus to right (TODO) column |
| `h` | Switch focus back to left column |
| `Enter` | Open selected document in `$EDITOR` |
| `o` | Open selected document in read mode |
