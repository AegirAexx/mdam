# mdam — /ˈmæd.əm/

> **Madam** by name, `mdam` by command — a keyboard-centric TUI tool for managing markdown documents, daily journals, and TODOs.

Inspired by [lazygit](https://github.com/jesseduffield/lazygit) and [atac](https://github.com/Julien-cpsMusic/ATAC) (keyboard-driven TUI design) and [zk](https://github.com/zk-org/zk) (plain-file notebook management).

**Design philosophy:**

- The filesystem is the database — no SQL, no cache, no sync service.
- Your editor does the editing — mdam never touches document bodies.
- mdam handles organization, navigation, and workflow automation.

> **Status: v0.1.0 — early, untested alpha.**
> All planned features are implemented but **none have been fully tested or verified**. Do not rely on any part of mdam for important data yet. The CLI surface, config format, and behavior may all change. Currently in active testing — features will be marked as tested individually as they are confirmed working.

---

## Installation

### Prerequisites

- Go 1.21+
- `$EDITOR` environment variable set (e.g., `nvim`)
- Git

### Recommended companion tools

- [neovim](https://neovim.io/) — mdam delegates all editing to `$EDITOR`; nvim pairs well
- [lazygit](https://github.com/jesseduffield/lazygit) — mdam can hand off to lazygit for git operations (`ctrl+g`)

### Build from source

```bash
git clone https://github.com/AegirAexx/mdam.git
cd mdam
go build -o mdam ./cmd/mdam
# Optionally move to your PATH:
mv mdam ~/.local/bin/
```

### First run

On first launch, mdam detects that no `base_dir` has been configured and starts an interactive setup flow. This will:

1. Prompt you for a base directory path (e.g., `~/notes`) — this becomes the root of your managed document tree.
2. Create the directory structure: `journal/`, `kb/`, `todo/`, `scratch/`, and `.templates/`.
3. Generate a default config file at `~/.config/mdam/config.yml`.
4. Seed the `.templates/` directory with built-in templates for journals, knowledge base documents, and TODOs.
5. Create an empty scratch pad at `scratch/scratch.md`.

This setup is fully idempotent — running it again won't overwrite existing files or directories.

---

## Quick Start

```bash
# Launch the TUI
mdam

# Create today's journal entry (CLI)
mdam journal create

# List open TODOs
mdam todo list

# Fuzzy search across all documents
mdam search "nginx"

# Import a markdown file
mdam import ~/Downloads/notes.md
```

### TUI layout

```
▶ Files ──────────────────────│─ Preview ──────────────────────────
> 2026-03-14.md          [M]  │  Journal 2026-03-14
  2026-03-13.md               │
  setup-nginx.md         [?]  │  type: journal
  deploy-runbook.md           │  tags: daily
  meeting-notes-…             │  modified: 2026-03-14
                              │─ TODOs ─────────────────────────────
                              │> - [ ] Review PR #42 @work
                              │  - [ ] Buy groceries @personal
──────────────────────────────────────────────────────────────────
 NORMAL │ main ↑2 │ 5 docs              /search  :cmd  ?help  q:quit
```

### Keybindings (summary)

| Key | Action |
|---|---|
| `j` / `k` | Move down / up |
| `1` | Dashboard (today's context) |
| `2` / `3` / `4` / `5` / `6` | Journal / KB / TODOs / Recent / Tags |
| `n` | New document (journal or KB) |
| `Enter` | Open selected document in `$EDITOR` |
| `s` | Open scratch pad in `$EDITOR` |
| `d` | Delete with confirmation (`y` / `n`) |
| `p` | Pin / unpin |
| `f` | Cycle smart filter (Untagged → Stale → Inbox) |
| `/` | Fuzzy search |
| `:` | Command mode (`:todo sweep`, `:q`) |
| `?` | Help overlay |
| `ctrl+g` | Open lazygit |
| `q` | Quit |

See [docs/KEYBINDINGS.md](docs/KEYBINDINGS.md) for the full reference.

---

## Configuration

mdam reads `~/.config/mdam/config.yml`. Run `mdam config --edit` to open it.

```yaml
editor: nvim                      # falls back to $EDITOR env var
author: "Your Name"
base_dir: ~/notes                 # root of your managed document tree
export_dir: ~/Downloads

theme: tokyonight                 # nord, gruvbox, catppuccin, dracula
nerd_fonts: false                 # true if your terminal font has Nerd Font glyphs

git:
  enabled: true
  lazygit: true                   # ctrl+g handoff to lazygit

todo:
  archive_after_days: 30

journal:
  auto_create: true
  sweep_on_create: true
```

---

## Features

| Feature | Status | Description |
|---|---|---|
| Daily journals | ⚠️ untested | Auto-created from templates, named `YYYY-MM-DD.md`, sweep TODOs on create |
| Knowledge base | ⚠️ untested | Organized reference documents with user-defined taxonomy and tags |
| TODO system | ⚠️ untested | Task backlog with category, priority, and date fields; sweep from journals |
| Scratch pad | ⚠️ untested | Persistent clipboard singleton, one keypress away (`s`) |
| Templates | ⚠️ untested | User-extensible document scaffolding; add `.md` files to `{base_dir}/.templates/` |
| Fuzzy search | ⚠️ untested | Across frontmatter fields, filenames, and document content |
| Export | ⚠️ untested | Strip frontmatter and share clean markdown (`e` or `mdam export`) |
| Git integration | ⚠️ untested | Per-file status markers in the TUI, lazygit handoff via `ctrl+g` |
| Dashboard | ⚠️ untested | Today's journal, open TODO count, pinned docs, and recent activity (`1`) |
| Tag browser | ⚠️ untested | All tags with document counts; drill into any tag (`6`) |
| Smart filter | ⚠️ untested | Post-filter by Untagged / Stale (>7 days) / Inbox (`f`) |
| Pin / unpin | ⚠️ untested | Bookmark documents; pins persist to `~/.config/mdam/pins.json` (`p`) |
| Color theming | ⚠️ untested | Five built-in palettes: tokyonight, nord, gruvbox, catppuccin, dracula |
| Markdown preview | ⚠️ untested | Glamour-rendered live preview in the right panel |

---

## Documentation

| Document | Contents |
|---|---|
| [docs/KEYBINDINGS.md](docs/KEYBINDINGS.md) | Full TUI keybinding reference |
| [docs/CLI.md](docs/CLI.md) | All CLI subcommands and flags |
| [docs/FRONTMATTER.md](docs/FRONTMATTER.md) | Frontmatter field contract |
| [docs/TODO-FORMAT.md](docs/TODO-FORMAT.md) | TODO task syntax and sweep/archive |
| [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) | Project structure, code style, testing |
| [docs/specs/mdam-spec-v1.md](docs/specs/mdam-spec-v1.md) | Full project specification |

---

## License

MIT — see [LICENSE](LICENSE) for details.
