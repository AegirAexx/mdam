# MadaM — Keybinding Reference

> **Status:** Current through Phase 5. Full keybinding review completed at Phase 5 start. Additional Phase 5 items (TODO-specific bindings, `g` vs `ctrl+g` revisit) deferred to real usage data.

## Modes

MadaM's TUI operates in four modes. There is no insert or visual mode — all text editing is handled by `$EDITOR`.

| Mode          | Purpose                                   | Activation     | Exit              |
|---------------|-------------------------------------------|----------------|-------------------|
| Normal        | Navigate, browse, trigger actions         | Default        | —                 |
| Command       | Execute colon-prefixed commands           | `:`            | `Enter` / `Esc`   |
| Search        | Fuzzy find across the document tree       | `/`            | `Enter` / `Esc`   |
| Delete?       | Confirm or cancel a document deletion     | `d` on a doc   | `y` (confirm) / `n` / `Esc` |

---

## Normal Mode

### Navigation

| Key          | Action                              |
|--------------|-------------------------------------|
| `j`          | Move down                           |
| `k`          | Move up                             |
| `h`          | Navigate left panel                 |
| `l`          | Navigate right panel                |
| `gg`         | Jump to top of list                 |
| `G`          | Jump to bottom of list              |
| `Tab`        | Cycle panel focus                   |
| `Shift+Tab`  | Cycle panel focus (reverse)         |

### Views

| Key  | Action                                        |
|------|-----------------------------------------------|
| `1`  | Dashboard (today's context)                   |
| `2`  | Journal view                                  |
| `3`  | Knowledge base view                           |
| `4`  | TODO view (focuses TODO panel)                |
| `5`  | Recent documents (top 20 by modified date)    |
| `6`  | Tag browser                                   |

> **Note:** ViewAll (all documents) is the startup default and is accessible whenever no number view is active. Pressing `/` to search and then `Esc` returns to ViewAll.

### Document Actions

| Key     | Action                                             |
|---------|----------------------------------------------------|
| `Enter` | Open selected document in `$EDITOR`               |
| `s`     | Open scratch pad in `$EDITOR`                      |
| `n`     | New document (template picker)                     |
| `d`     | Delete selected document (prompts for confirmation)|
| `e`     | Export selected document (strip frontmatter)       |
| `p`     | Pin / unpin selected document                      |

### Filtering & Search

| Key  | Action                                                   |
|------|----------------------------------------------------------|
| `/`  | Enter search mode (fuzzy search)                         |
| `f`  | Cycle smart filter: None → Untagged → Stale → Inbox → … |

### Git

| Key      | Action                               |
|----------|--------------------------------------|
| `ctrl+g` | Open lazygit in the managed tree     |

> **Note:** `ctrl+g` is used instead of `g` to avoid conflicting with the `gg` jump-to-top chord. This will be reconsidered after Phase 5 real usage data is available.

### Application

| Key  | Action                     |
|------|----------------------------|
| `?`  | Toggle keybinding help overlay |
| `q`  | Quit MadaM                 |
| `R`  | Force re-scan directory tree |

---

## Delete Confirmation Mode

Entered when `d` is pressed on a selected document.

| Key        | Action                    |
|------------|---------------------------|
| `y`        | Confirm delete            |
| `n`        | Cancel (return to Normal) |
| `Esc`      | Cancel (return to Normal) |

---

## Command Mode

Entered via `:` in normal mode. Commands are executed on `Enter`, cancelled with `Esc`.

| Command                          | Action                                     |
|----------------------------------|--------------------------------------------|
| `:q`                             | Quit                                       |
| `:todo sweep`                    | Run TODO sweep manually                    |
| `:todo archive`                  | Archive old completed tasks                |

---

## Search Mode

Entered via `/` in normal mode. Type to fuzzy-search, results update live.

| Key        | Action                              |
|------------|-------------------------------------|
| `Enter`    | Open selected result in `$EDITOR`   |
| `Esc`      | Cancel search, return to Normal     |
| `j` / `k`  | Navigate search results             |

---

## Tag Browser (key `6`)

The tag browser replaces the normal panel layout. The left panel lists all tags (sorted by document count). The right panel shows documents carrying the selected tag.

| Key     | Action                                                |
|---------|-------------------------------------------------------|
| `j`/`k` | Navigate tags                                         |
| `Enter` | Search for documents with the selected tag            |
| `gg`/`G`| Jump to top / bottom of tag list                      |
| `6`     | Return to tag browser (press any other view key to leave) |

---

## Smart Filter (key `f`)

Smart filters are post-filters applied over the ViewAll document list. Press `f` repeatedly to cycle:

| Filter       | What it shows                                   |
|--------------|-------------------------------------------------|
| None         | All documents (default)                         |
| Untagged     | Documents with no tags                          |
| Stale        | Documents not modified in the last 7 days       |
| Inbox        | Documents with `type: unsorted`                 |

The active filter is shown as a bar below the panel header and in the status bar.

---

## Design Conventions

These conventions guide keybinding decisions and should be maintained as new actions are added:

- **Lowercase** for common, non-destructive actions (`j`, `k`, `s`, `n`, `e`, `p`, `f`)
- **Uppercase** for infrequent actions (`R` for re-scan)
- **Numbers** for view switching (`1`–`6`)
- **Single letter** preferred over chords for high-frequency actions
- **Vim muscle memory** respected — `hjkl`, `gg`/`G`, `/`, `:`, `q` behave as expected
- **No collisions** — every key has exactly one action per mode

---

## TODO (Phase 5+ deferred)

- [x] Finalize keybindings for Phase 4 features (`Enter`, `s`, `ctrl+g`)
- [x] Define delete confirmation keybindings (`d` to initiate, `y`/`n`/`Esc` to confirm/cancel)
- [x] Define tag browser keybindings
- [x] Define smart filter cycling (`f`)
- [x] Define pin/unpin (`p`)
- [ ] Decide on `j`/`k` vs arrow keys for search result navigation (deferred — real usage needed)
- [ ] Revisit `g` vs `ctrl+g` for lazygit (deferred — real usage needed)
- [ ] Define TODO-specific keybindings (mark done, change status, change category)
- [ ] Test for ergonomic conflicts with common terminal emulator shortcuts
