# Issue #5: Restructure README for End Users

**Type:** Enhancement
**Priority:** Medium
**Labels:** `documentation`, `phase-6`

## Problem

The current README is organized as a technical reference (architecture, frontmatter contracts, project structure). It reads like internal documentation, not a user-facing guide. A new user who lands on the GitHub repo needs to know: what is this, how do I install it, how do I use it — in that order.

## New README Structure

### 1. Header & Introduction

```markdown
# MadaM — Markdown Admin Management

A keyboard-centric TUI for managing markdown documents, daily journals, and TODOs.
Inspired by lazygit, yazi, k9s, and atac.

Your editor does the editing. MadaM does the organizing.

> **Status:** Early alpha — actively developed and tested.
```

- What it is (one paragraph)
- Design inspiration (lazygit, yazi, k9s, atac)
- Emphasize TUI as primary interface; CLI as secondary for scripting
- Status: early alpha

### 2. Installation

This is the most important section after the introduction.

#### Prerequisites

- Go 1.21+
- `$EDITOR` environment variable set (e.g., `nvim`, `vim`, `nano`)
- Git
- Optional: [lazygit](https://github.com/jesseduffield/lazygit) for git handoff (`ctrl+g`)
- Optional: A [Nerd Font](https://www.nerdfonts.com/) for icon support

#### Build & Install

```bash
git clone https://github.com/AegirAexx/mdam.git
cd mdam
go build -o mdam ./cmd/mdam

# Install to your PATH (choose one):
cp mdam ~/.local/bin/       # user-local
sudo cp mdam /usr/local/bin/ # system-wide
```

Note: Go compiles a static binary with no runtime dependencies.
No cross-compilation flags are needed when building on the target machine.

#### First Run

```bash
mdam
```

On first launch, MadaM will guide you through setup: choosing a base directory, creating the folder structure, and generating a default config.

### 3. Quick Start (TUI first)

Show the TUI wireframe, then the most common keybindings. Keep it brief — link to KEYBINDINGS.md for the full reference.

Then show 4-5 CLI examples for the CLI crowd.

### 4. Configuration

The `config.yml` reference. Move from its current position higher up, since users will want to customize immediately after install.

### 5. Features

Consolidated feature descriptions — what the tool does, not how it's built. The current feature list is good but should be rewritten in terms of user benefits, not implementation details.

### 6. Everything Else

Move to linked sub-documents:

| Content | Location |
|---------|----------|
| Full keybinding reference | `docs/KEYBINDINGS.md` |
| Project specification | `docs/mdam-spec-v1.md` |
| Frontmatter contract | `docs/FRONTMATTER.md` (new) |
| TODO task format | `docs/TODO-FORMAT.md` (new) |
| CLI reference | `docs/CLI.md` (new) |
| Development guide | `docs/DEVELOPMENT.md` (new) |
| Phase reports | `docs/phase-*.md` |

The README links to these but does not duplicate their content.

### What to remove from README

- Project structure tree (move to `docs/DEVELOPMENT.md`)
- Phase status table (replace with single status line)
- Detailed frontmatter contract (move to own doc)
- Detailed TODO format (move to own doc)
- Full CLI reference (move to own doc, keep 5 examples in README)
- Development commands (move to `docs/DEVELOPMENT.md`)

## Acceptance Criteria

- [ ] README follows the new structure: intro → install → quick start → config → features → links
- [ ] TUI is presented first, CLI second
- [ ] Installation section includes prerequisites, build, install to PATH, and first run
- [ ] Detailed reference content moved to linked sub-documents
- [ ] All links resolve correctly
- [ ] No content is lost — only reorganized
