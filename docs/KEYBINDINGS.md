# MadaM — Keybinding Reference

> **Status:** Current through Phase 4. Keybindings for Phase 5 features (tag browser, delete confirmation, ambient findability views) are not yet defined. The full keybinding pass described in the spec will happen at the Phase 5 start after real usage with the Phase 4 tool informs the design.

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

| Key      | Action                                       |
|----------|----------------------------------------------|
| `ctrl+g` | Open lazygit in the managed tree             |

> **Note:** The spec lists `g` for lazygit, but `g` is also the first key of the `gg` jump-to-top chord. `ctrl+g` avoids the conflict. This will be revisited during the Phase 5 keybinding pass.

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

- [x] Finalize keybindings for Phase 4 features (`Enter`, `s`, `ctrl+g`)
- [ ] Decide on `j`/`k` vs arrow keys for search result navigation
- [ ] Revisit `g` vs `ctrl+g` for lazygit after real usage in Phase 4
- [ ] Define TODO-specific keybindings (mark done, change status, change category)
- [ ] Define tag browser keybindings
- [ ] Define delete confirmation keybindings (`d` to initiate, `y`/`n` or `Enter`/`Esc` to confirm)
- [ ] Test for ergonomic conflicts with common terminal emulator shortcuts
- [ ] Full keybinding review and rationalization at Phase 5 start
