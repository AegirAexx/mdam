# mdam — /ˈmæd.əm/

> **Madam** by name, `mdam` by command — a keyboard-centric TUI tool for managing markdown documents, daily journals, and TODOs.

Inspired by [lazygit](https://github.com/jesseduffield/lazygit) and [atac](https://github.com/Julien-cpsMusic/ATAC) (keyboard-driven TUI design) and [zk](https://github.com/zk-org/zk) (plain-file notebook management).

**Design philosophy:**

- The filesystem is the database — no SQL, no cache, no sync service.
- Your editor does the editing — mdam never touches document bodies.
- mdam handles organization, navigation, and workflow automation.

> **Status: v0.1.0 — early alpha.**
> The CLI surface, config format, and behavior may change before v1.0.

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

On first launch, mdam detects no `base_dir` has been configured and starts an interactive setup flow:

1. Prompts for a base directory path (e.g., `~/notes`).
2. Creates the directory structure: `journal/`, `kb/`, `todo/`, `scratch/`, `.templates/`.
3. Writes a default config to `~/.config/mdam/config.yml`.
4. Seeds `.templates/` with built-in templates and creates an empty scratch pad.

Setup is fully idempotent — re-running won't overwrite existing files.

---

## Quick Start

```bash
mdam                      # Launch the TUI
mdam journal create       # Create today's journal entry (CLI)
mdam todo list            # List open TODOs
mdam search "nginx"       # Fuzzy search across all documents
mdam import ~/notes.md    # Import a markdown file
```

### TUI layout

```
 Dashboard   Journal   KB   Tag Browser
▶ Overview ─────────────────│─ Preview ──────────────────────────
 Journal                    │  Journal 2026-04-03
 2026-04-03                 │
 2026-04-01                 │  tags: daily
 Pinned                     │
 Recent                     │
 Setup Nginx                │
────────────────────────────────────────────────────────────────
 NORMAL │ main ↑2 │ 3 journal · 1 kb · 0 scratch   /  :  o:read  ?  q
```

The tab bar at the top shows all four panes. The left column is navigable; the right shows a live glamour-rendered preview of the selected document.

### Keybindings (summary)

| Key | Action |
|---|---|
| `j` / `k` | Move down / up |
| `h` / `l` | Switch panels or expand/collapse tree folders |
| `Tab` / `Shift+Tab` | Cycle panes forward / backward |
| `1` | Dashboard |
| `2` | Journal (month-folder tree) |
| `3` | KB (subtype-folder tree) |
| `4` | Tag Browser |
| `Enter` | Open selected document in `$EDITOR` |
| `o` | Open selected document in full-screen read mode |
| `s` | Open scratch pad in `$EDITOR` |
| `n` | New document (template picker) |
| `d` | Delete with confirmation (`y` / `n`) |
| `p` | Pin / unpin |
| `e` | Export (strip frontmatter) |
| `/` | Fuzzy search |
| `:` | Command mode (`:todo sweep`, `:q`) |
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

todo:
  archive_after_days: 30

journal:
  auto_create: true     # create today's entry on startup
  sweep_on_create: true # carry forward open tasks from yesterday
```

---

## Features

| Feature | Description |
|---|---|
| Daily journals | Auto-created from templates, named `YYYY-MM-DD.md`, grouped in a month-folder tree |
| Knowledge base | Subtype folders derived from `kb_*` type prefix (e.g. `kb_summary` → Summary folder) |
| TODO system | Task backlog with category, priority, and date fields; sweep from journal entries |
| Scratch pad | Persistent singleton, one keypress away (`s`) |
| Templates | User-extensible scaffolding; add `.md` files to `{base_dir}/.templates/` |
| Fuzzy search | Across frontmatter fields, filenames, and document bodies |
| Export | Strip frontmatter and share clean markdown (`e` or `mdam export`) |
| Git integration | Per-file status markers (modified / untracked) in the file panel |
| Dashboard | Navigable two-column view: recent journal / pinned / recent docs + open TODOs |
| Tag browser | All tags with document counts; navigate into any tag to see its documents |
| Read mode | Full-screen glamour-rendered overlay (`o`); scroll with `j`/`k`/`Space` |
| Pin / unpin | Bookmark documents; pins persist to `~/.config/mdam/pins.json` (`p`) |
| Color theming | Five built-in palettes: tokyonight, nord, gruvbox, catppuccin, dracula |
| Markdown preview | Live glamour-rendered preview in the right panel |

---

## Documentation

| Document | Contents |
|---|---|
| [docs/KEYBINDINGS.md](docs/KEYBINDINGS.md) | Full TUI keybinding reference |
| [docs/CLI.md](docs/CLI.md) | All CLI subcommands and flags |
| [docs/FRONTMATTER.md](docs/FRONTMATTER.md) | Frontmatter field contract |
| [docs/TODO-FORMAT.md](docs/TODO-FORMAT.md) | TODO task syntax and sweep/archive |
| [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) | Project structure, code style, testing |

---

## License

MIT — see [LICENSE](LICENSE) for details.
