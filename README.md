# MadaM — Markdown Admin Management

A keyboard-centric TUI tool for managing markdown documents, daily journals, and TODOs.

**Design philosophy:** The filesystem is the database. Your editor does the editing. MadaM handles organization, navigation, and workflow automation — it never edits document bodies.

> **Status: early alpha.** Core functionality is complete and tested, but the CLI surface and config format may change before 1.0.

---

## Installation

**Prerequisites:** Go 1.21+, `$EDITOR` set (e.g., `nvim`), Git

```bash
git clone https://github.com/AegirAexx/mdam.git
cd mdam
go build -o mdam ./cmd/mdam
# Optionally move to your PATH:
mv mdam ~/.local/bin/
```

**First run:** On first launch, MadaM detects a missing `base_dir` and runs an interactive setup to create your managed document tree and config file.

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

MadaM reads `~/.config/mdam/config.yml`. Run `mdam config --edit` to open it.

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

- **Daily journals** — Auto-created from templates, named `YYYY-MM-DD.md`, sweep TODOs on create
- **Knowledge base** — Organized reference documents with user-defined taxonomy and tags
- **TODO system** — Task backlog with category, priority, and date fields; sweep from journals
- **Scratch pad** — Persistent clipboard singleton, one keypress away (`s`)
- **Templates** — User-extensible document scaffolding; add `.md` files to `{base_dir}/.templates/`
- **Fuzzy search** — Across frontmatter fields, filenames, and document content
- **Export** — Strip frontmatter and share clean markdown (`e` or `mdam export`)
- **Git integration** — Per-file status markers in the TUI, lazygit handoff via `ctrl+g`
- **Dashboard** — Today's journal, open TODO count, pinned docs, and recent activity (`1`)
- **Tag browser** — All tags with document counts; drill into any tag (`6`)
- **Smart filter** — Post-filter by Untagged / Stale (>7 days) / Inbox (`f`)
- **Pin / unpin** — Bookmark documents; pins persist to `~/.config/mdam/pins.json` (`p`)
- **Color theming** — Five built-in palettes: tokyonight, nord, gruvbox, catppuccin, dracula
- **Markdown preview** — Glamour-rendered live preview in the right panel

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
