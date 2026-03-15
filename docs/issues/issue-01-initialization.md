# Issue #1: First-Run Initialization & Scaffolding

**Type:** Enhancement
**Priority:** High
**Labels:** `ux`, `onboarding`, `phase-6`

## Problem

MadaM assumes the base directory, subfolder structure, and config file all exist before launch. A new user who builds from source and runs `mdam` gets undefined behavior — missing directories, no config, no templates. There is no guided setup.

## Expected Behavior

On first launch (or when required infrastructure is missing), MadaM should detect what's missing and offer to scaffold it. The user should go from `./mdam` to a working setup in under 30 seconds.

## Requirements

### 1. Config detection and creation

- On startup, check if `~/.config/mdam/config.yml` exists.
- If missing, create a default config with **verbose inline comments** explaining every field, its default value, and what it controls.
- If present, validate it and report any issues (unknown keys, invalid types, missing required fields).

### 2. Base directory setup

- `base_dir` should default to empty string (`""`), not a hardcoded path.
- If `base_dir` is empty or does not exist, prompt the user: "Where should MadaM manage your documents? Enter a path or press Enter for ~/notes"
- In TUI mode, show a blocking first-run screen. In CLI mode, print to stderr and read from stdin.

### 3. Folder structure scaffolding

After `base_dir` is confirmed, offer to create the expected subfolder structure:

```
~/notes/               # base_dir (user-defined)
├── journal/           # daily journal entries
├── kb/                # knowledge base documents
├── todo/              # global TODO file lives here
├── scratch/           # scratch pad singleton
├── .inbox/            # import inbox
└── .templates/        # user-defined templates (builtins copied here)
```

- Ask: "Create default folder structure? (Y/n)"
- Create only what's missing — don't overwrite existing directories.
- Copy built-in templates to `.templates/` if the directory is empty.

### 4. Scratch pad creation

- If `scratch/scratch.md` does not exist, create it with valid frontmatter.
- The scratch pad is a singleton — it should always exist after initialization.

## Acceptance Criteria

- [ ] Running `mdam` with no config produces a valid `config.yml` with comments
- [ ] Running `mdam` with empty `base_dir` prompts for a path
- [ ] Folder scaffolding creates all expected subdirectories
- [ ] Built-in templates are copied to `.templates/`
- [ ] Scratch pad is created with valid frontmatter
- [ ] All initialization is idempotent — running it twice changes nothing
- [ ] Existing files and directories are never overwritten
- [ ] `go test ./...` passes after changes
