# mdam — /ˈmæd.əm/

> **Madam** by name, `mdam` by command — a keyboard-centric TUI tool for managing markdown documents and daily journals.

Inspired by [lazygit](https://github.com/jesseduffield/lazygit) and [atac](https://github.com/Julien-cpsMusic/ATAC) (keyboard-driven TUI design) and [zk](https://github.com/zk-org/zk) (plain-file notebook management).

**Design philosophy:**

- The filesystem is the database — no SQL, no cache, no sync service.
- Your editor does the editing — mdam never touches document bodies.
- mdam handles organization, navigation, and workflow automation.

> **Status: alpha — usable but unverified.**
> Everything works for daily use. The CLI surface and config format may still change.

---

## Installation

### Prerequisites

- Go 1.21+
- `$EDITOR` environment variable set (e.g., `nvim`)
- Git

### Build from source

```bash
git clone https://github.com/AegirAexx/mdam.git
cd mdam
go build -o mdam ./cmd/mdam
mv mdam ~/.local/bin/   # optional
```

### First run

On first launch, mdam detects no configuration and starts a TUI setup wizard that walks you through:

1. Base directory path (e.g., `~/notes`).
2. Editor, author name, color theme, Nerd Font preference, export directory.
3. Shows a preview of the generated config before saving.

The wizard creates:

- Directory structure: `journal/`, `kb/`, `.templates/`, `.mdam/`
- Singleton files: `todo.md`, `scratch.md` (at base root)
- A getting-started KB document in `kb/` covering onboarding and usage
- Config at `~/.config/mdam/config.yml`
- Built-in templates: `journal.md`, `kb.md` in `.templates/`

Setup is fully idempotent — re-running won't overwrite existing files or templates.

---

## Quick Start

```bash
mdam                      # Launch the TUI
mdam journal create       # Create today's journal entry (CLI)
mdam search "nginx"       # Fuzzy search across all documents
```

### TUI layout

```
 1: Dashboard   2: Journal   3: KB   4: Tag Browser   5: Search
▶ Overview ─────────────────│─ Todo ─────────────────────────────
 Journal                    │  - [ ] Review PR #42
 2026-04-03                 │  - [ ] Deploy staging
 2026-04-01                 │
 Pinned                     │
 Recent                     │
 Setup Nginx                │
────────────────────────────────────────────────────────────────
 NORMAL │ main ↑2 │ 3 journal · 1 kb │ /  :  o:read  t:todo  s:scratch  ?  q
```

The tab bar at the top shows all five panes. The left column is navigable; the right shows a live glamour-rendered preview. On the Dashboard, the right panel renders `todo.md`.

### Keybindings (summary)

| Key | Action |
|---|---|
| `j` / `k` | Move down / up |
| `h` / `l` | Switch panels or expand/collapse tree folders |
| `Tab` / `Shift+Tab` | Cycle panes forward / backward |
| `g` / `G` | Jump to top / bottom |
| `1` | Dashboard |
| `2` | Journal (month-folder tree) |
| `3` | KB (subtype-folder tree) |
| `4` | Tag Browser (with substring filter) |
| `5` / `/` | Search pane |
| `Enter` | Open in `$EDITOR` / activate filter input |
| `o` | Full-screen read mode (glamour) |
| `s` | Scratch pad |
| `t` | Todo |
| `n` | New document (template picker) |
| `p` | Pin / unpin |
| `e` | Export (strip frontmatter) |
| `Esc` | Clear filter / search results |
| `:` | Command mode (`:q`) |
| `?` | Help overlay |
| `q` | Quit |

See [docs/KEYBINDINGS.md](docs/KEYBINDINGS.md) for the full reference.

---

## Configuration

mdam reads `~/.config/mdam/config.yml`. Run `mdam config --edit` to open it.

```yaml
editor: "nvim"          # defaults to $EDITOR
author: "YourName"
base_dir: ~/notes       # root of your document tree
export_dir: ~/Downloads
theme: tokyonight       # tokyonight | nord | gruvbox | catppuccin | dracula
nerd_fonts: false       # set true if your terminal uses a Nerd Font

journal:
  auto_create: true     # create today's entry on startup
```

---

## Features

| Feature | Description |
|---|---|
| Daily journals | Auto-created from templates, named `YYYY-MM-DD.md`, grouped in a month-folder tree |
| Knowledge base | Subtype folders derived from `kb_*` type prefix (e.g. `kb_summary` -> Summary folder) |
| Todo list | Simple `todo.md` at base root, rendered on the dashboard, opened in `$EDITOR` (`t`) |
| Scratch pad | Persistent singleton, one keypress away (`s`) |
| Templates | Two built-in templates (journal, kb); add custom `.md` files to `{base_dir}/.templates/` |
| Search pane | Dedicated pane (`5`/`/`) — results categorized by Journal, KB, Tags; documents openable from results |
| Export | Strip frontmatter and share clean markdown (`e` or `mdam export`) |
| Git integration | Branch + sync status in the status bar; per-file markers (modified / untracked / staged) in all views |
| Dashboard | Navigable two-column view: recent journal / pinned / recent docs + glamour-rendered todo.md |
| Tag browser | All tags with document counts; substring filter to narrow the list; navigate into any tag to see its documents |
| Read mode | Full-screen glamour-rendered overlay (`o`); vim navigation (`j`/`k`/`d`/`u`/`f`/`b`/`g`/`G`) |
| Pin / unpin | Bookmark documents (max 10, FIFO eviction); pins persist to `{base_dir}/.mdam/pins.json` (`p`) |
| Color theming | Five built-in palettes: tokyonight, nord, gruvbox, catppuccin, dracula |
| Markdown preview | Live glamour-rendered preview in the right panel |

---

## Documentation

| Document | Contents |
|---|---|
| [docs/KEYBINDINGS.md](docs/KEYBINDINGS.md) | Full TUI keybinding reference |
| [docs/CLI.md](docs/CLI.md) | All CLI subcommands and flags |
| [docs/FRONTMATTER.md](docs/FRONTMATTER.md) | Frontmatter field contract |
| [docs/TODO-FORMAT.md](docs/TODO-FORMAT.md) | TODO task format |
| [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) | Project structure, code style, testing |

---

## License

MIT — see [LICENSE](LICENSE) for details.
