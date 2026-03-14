# MadaM — Keybinding Reference

> **Status:** Draft. Keybindings are tentative and will be finalized during the Phase 3–4 transition after real usage informs the design. This document is a living reference updated throughout development.

## Modes

MadaM's TUI operates in three modes. There is no insert or visual mode — all text editing is handled by `$EDITOR`.

| Mode    | Purpose                                | Activation | Exit        |
|---------|----------------------------------------|------------|-------------|
| Normal  | Navigate, browse, trigger actions      | Default    | —           |
| Command | Execute colon-prefixed commands        | `:`        | `Enter`/`Esc` |
| Search  | Fuzzy find across the document tree    | `/`        | `Enter`/`Esc` |

---

## Normal Mode

### Navigation

| Key        | Action                              |
|------------|-------------------------------------|
| `j`        | Move down                           |
| `k`        | Move up                             |
| `h`        | Collapse / navigate left panel      |
| `l`        | Expand / navigate right panel       |
| `gg`       | Jump to top of list                 |
| `G`        | Jump to bottom of list              |
| `Tab`      | Cycle panel focus                   |
| `Shift+Tab`| Cycle panel focus (reverse)         |

### Document Actions

| Key     | Action                                        |
|---------|-----------------------------------------------|
| `Enter` | Open selected document in `$EDITOR`           |
| `s`     | Open scratch pad in `$EDITOR`                 |
| `n`     | New document (template picker)                |
| `d`     | Delete selected document (with confirmation)  |
| `m`     | Move selected document to another directory   |
| `e`     | Export selected document (strip frontmatter)  |
| `y`     | Copy document path to clipboard               |

### Views & Navigation

| Key     | Action                                        |
|---------|-----------------------------------------------|
| `/`     | Enter search mode                             |
| `:`     | Enter command mode                            |
| `?`     | Toggle keybinding help overlay                |
| `1`     | Switch to today's context / dashboard         |
| `2`     | Switch to journal view                        |
| `3`     | Switch to knowledge base view                 |
| `4`     | Switch to TODO view                           |
| `5`     | Switch to recent documents view               |

### Git

| Key     | Action                                        |
|---------|-----------------------------------------------|
| `g`     | Open lazygit in the managed tree              |

### Application

| Key     | Action                                        |
|---------|-----------------------------------------------|
| `q`     | Quit MadaM                                    |
| `R`     | Force re-scan directory tree                  |

---

## Command Mode

Entered via `:` in normal mode. Commands are executed on `Enter`, cancelled with `Esc`.

| Command                          | Action                                     |
|----------------------------------|--------------------------------------------|
| `:q`                             | Quit                                       |
| `:sync`                          | Git add, commit, pull --rebase, push       |
| `:import <path>`                 | Import file or directory                   |
| `:import <path> --auto-fix`      | Import with auto-fix                       |
| `:export`                        | Export selected document                   |
| `:export --clipboard`            | Export to clipboard                        |
| `:todo sweep`                    | Run TODO sweep manually                    |
| `:todo archive`                  | Archive old completed tasks                |
| `:template list`                 | List available templates                   |
| `:config`                        | Open config.yml in `$EDITOR`               |

---

## Search Mode

Entered via `/` in normal mode. Type to fuzzy-search, results update live.

| Key        | Action                              |
|------------|-------------------------------------|
| `Enter`    | Open selected result in `$EDITOR`   |
| `Esc`      | Cancel search, return to normal     |
| `j` / `k`  | Navigate search results (TBD: may use arrow keys instead) |
| `Tab`      | Cycle search scope (all / tags / filenames / content) |

---

## Design Conventions

These conventions guide keybinding decisions and should be maintained as new actions are added:

- **Lowercase** for common, non-destructive actions (`j`, `k`, `s`, `n`, `e`)
- **Uppercase** for infrequent or potentially destructive actions (`R` for re-scan, `D` for force-delete if added)
- **Numbers** for view switching (`1`–`5`)
- **Single letter** preferred over chords for high-frequency actions
- **Vim muscle memory** respected — `hjkl`, `gg`/`G`, `/`, `:`, `q` behave as expected
- **No collisions** — every key has exactly one action per mode

---

## TODO

- [ ] Finalize keybindings after Phase 3 integration testing
- [ ] Decide on `j`/`k` vs arrow keys for search result navigation
- [ ] Define TODO-specific keybindings (mark done, change status, change category)
- [ ] Define tag browser keybindings
- [ ] Test for ergonomic conflicts with common terminal emulator shortcuts
