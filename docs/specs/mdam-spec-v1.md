# mdam — /ˈmæd.əm/ — Markdown Admin Management

## Project Specification v1.0

> **Version: v0.1.0 — early, untested alpha.** All planned features are implemented but none have been individually tested or verified. The CLI surface, config format, and behavior may all change.

---

## 1. Overview

mdam is a terminal-based markdown administration and routing tool for personal knowledge management. It manages, organizes, navigates, and automates workflows around markdown documents. It does not edit them. All viewing and authoring is delegated to the user's external `$EDITOR`.

Inspired by [lazygit](https://github.com/jesseduffield/lazygit) and [atac](https://github.com/Julien-cpsMusic/ATAC) (keyboard-driven TUI design) and [zk](https://github.com/zk-org/zk) (plain-file notebook management).

The filesystem is the absolute source of truth. There is no database, no caching layer, no state synchronization. The application relies entirely on raw I/O performance to scan and parse the managed directory tree on every relevant operation.

The command is `mdam`. It operates in two modes: as an interactive TUI and as a headless CLI utility with standard UNIX flags and subcommands, suitable for scripting and pipeline integration.

---

## 2. Design Philosophy

- **Admin tool, not an editor.** The TUI handles navigation, selection, file creation, organization, and automation. The moment a document needs to be read or written by a human, the application suspends and hands control to `$EDITOR`. On editor exit, the application resumes and re-scans.
- **Filesystem is the database.** Folder structure defines categorization. File content and YAML frontmatter define metadata. Nothing is derived, cached, or stored outside the file tree.
- **Stateless by default.** Every operation reads from disk. No in-memory state persists between operations except what BubbleTea's model holds for the current render cycle.
- **Standard library first.** The Go standard library is preferred over third-party packages wherever feasible. External dependencies are adopted only when they provide substantial value that cannot be reasonably achieved with the stdlib (e.g., BubbleTea for the TUI event loop, Cobra for CLI subcommand routing).
- **Test-first.** Standard Go table-driven tests alongside every function from the first line of source code. The headless engine is fully verified before any UI exists.
- **Dual interface.** Every core engine feature is accessible both through the interactive TUI and as a headless CLI subcommand. The TUI is the primary interface but the engine never depends on it. A sysadmin should be able to run `mdam todo list --status open --category work` in a shell script without launching the TUI.
- **Agent-friendly by convention.** Consistent frontmatter contracts, predictable folder conventions, and machine-readable metadata make the managed tree naturally consumable by external tools and LLM agents without coupling the engine to any specific framework.

---

## 3. Core Engine

The engine is the headless layer: pure Go functions for file operations, validation, parsing, TODO management, template interpolation, and git status. Every feature described in this section is testable and runnable without the TUI.

### 3.1 Markdown Documents

The fundamental unit. Every managed file is a markdown document with YAML frontmatter, organized in a folder tree rooted at a user-configured base directory.

**Frontmatter contract — required fields:**

| Field      | Type     | Description                                               |
|------------|----------|-----------------------------------------------------------|
| `title`    | string   | Human-readable document title                             |
| `tags`     | list     | Tags for search and categorization                        |
| `created`  | datetime | ISO 8601 timestamp of creation                            |
| `modified` | datetime | ISO 8601 timestamp of last modification                   |
| `type`     | string   | Document type (journal, kb, todo, scratch, unsorted)      |

Additional fields may be defined per document type (e.g., `kb_type: howto`). The engine validates required fields but passes through any additional frontmatter without complaint.

**Filename standards:** All filenames must be kebab-case, POSIX-friendly, and URL-safe. The sole exception is daily journal entries which use the `YYYY-MM-DD.md` format.

**Frontmatter parsing:** The engine parses YAML between the opening and closing `---` delimiters. It validates structure and required fields. It does not parse or validate the markdown body in any way — only the frontmatter and the filename.

### 3.2 Import & Validation Pipeline

A dedicated inbox directory where the user drops markdown files for the application to ingest.

**Validation checks:**

| Check               | Criteria                                                     |
|----------------------|--------------------------------------------------------------|
| Filename format      | Kebab-case, POSIX/URL-safe characters only                  |
| Frontmatter exists   | File contains opening and closing `---` delimiters           |
| YAML validity        | Frontmatter parses as valid YAML                             |
| Required fields      | All required frontmatter fields present and correctly typed  |
| Duplicate detection  | No existing file with the same name in the managed tree      |

**Resolution flow on validation failure:**

- **Auto-fix:** The engine renames the file to kebab-case, scaffolds missing frontmatter with sensible defaults (`created` from file mtime, `title` derived from filename, empty `tags`, `type: unsorted`), and re-validates.
- **Manual fix:** The file is opened in `$EDITOR` for the user to correct. On editor exit, the engine re-validates.

The choice between auto-fix and manual is presented per-file in the TUI, or controlled by the `import.auto_fix` configuration flag in headless mode.

**Batch import:** For onboarding existing markdown collections. Point `mdam import <directory>` at a folder, batch-validate all files, present a summary of issues, and offer to auto-fix all or review individually.

**CLI access:**

```
mdam import <path>              # validate and import a file or directory
mdam import <path> --auto-fix   # auto-fix all validation issues
mdam import <path> --dry-run    # report issues without modifying anything
```

### 3.3 Daily Journal Entries

Managed creation of daily journal documents named `YYYY-MM-DD.md`, stored in the configured journal directory.

**Creation:** When the user requests a new journal entry (or on application startup if configured and today's entry doesn't exist), the engine creates it from the journal template with the current date interpolated.

**Structure:** Each journal entry includes a dedicated TODO section for that day's tasks. See section 3.4 for the sweep mechanics.

**Editability:** All journal entries — past and present — are fully editable via the editor handoff. The engine's only automated mutation of journal content is the TODO sweep (removing incomplete tasks from past entries). The user can always open any journal entry to add tags, fix content, append notes, or make any other changes.

**CLI access:**

```
mdam journal create             # create today's entry (no-op if exists)
mdam journal create 2026-03-10  # create entry for a specific date
mdam journal list               # list journal entries
mdam journal list --month 2026-03  # list entries for a specific month
```

### 3.4 TODO System

A two-tier task management system bridging daily journaling with persistent task tracking.

#### 3.4.1 Global TODO Document

A single global TODO file that serves as the master backlog. Tasks live here until pulled into a daily journal, completed, or cancelled.

**Task attributes:**

| Attribute   | Required | Values / Format                                  |
|-------------|----------|--------------------------------------------------|
| `status`    | yes      | `open`, `in-progress`, `done`, `cancelled`       |
| `category`  | no       | User-defined string (work, personal, project-x)  |
| `priority`  | no       | User-defined (high/medium/low or numeric)        |
| `created`   | yes      | Date the task was created                         |
| `completed` | no       | Date completed or cancelled                       |

**In-file format:** Tasks use standard markdown checkbox syntax with metadata encoded in a consistent, parseable format:

```markdown
- [ ] Review PR #42 @work !high (2026-03-14)
- [x] Update DNS records @work (2026-03-12) ✓2026-03-13
- [ ] Buy groceries @personal (2026-03-14)
```

The exact syntax is a design decision to finalize during Phase 1 implementation. The engine must be able to parse status, category, priority, created date, and completion date from each line using `bufio.Scanner` and `strings`/`regexp`.

#### 3.4.2 Daily Journal TODO Section

Each journal entry includes a TODO section representing the tasks the user intends to work on that day. Tasks can be pulled from the global TODO or added directly in the journal.

#### 3.4.3 The Sweep

The core mechanism that prevents TODO decay. Triggered when today's date advances past a journal entry (typically on application startup or when creating a new journal entry):

1. **Scan** the past journal entry's TODO section line by line.
2. **Completed tasks** (`done`, `cancelled`) remain in the journal as a historical record. Their status is synced to the global TODO if they exist there.
3. **Incomplete tasks** (`open`, `in-progress`) are removed from the journal entry and promoted back to the global TODO if not already present.
4. **New tasks** (added directly in the journal, not present in the global TODO) are added to the global TODO during the sweep.

After the sweep, the global TODO is the current source of truth for all open tasks, and the journal entry is a historical record of what was planned and accomplished that day.

**Implementation note:** This is in-memory text mutation — read the full file into a buffer, parse the TODO section, mutate, write back with `os.WriteFile`. A core Phase 1 feature requiring careful testing of edge cases: empty TODO sections, malformed checkboxes, tasks with multi-line descriptions, tasks that exist in both journal and global.

#### 3.4.4 Archive

Completed and cancelled tasks are retained in the global TODO with their status and completion date, but filtered out of the default view. A periodic archive operation moves done/cancelled items older than a configurable threshold to a separate archive file to keep the global TODO lean.

**CLI access:**

```
mdam todo list                          # list open tasks (default view)
mdam todo list --status done            # list completed tasks
mdam todo list --category work          # filter by category
mdam todo list --all                    # show all including done/cancelled
mdam todo sweep                         # run the sweep manually
mdam todo archive                       # archive old completed tasks
mdam todo archive --older-than 30       # archive tasks completed 30+ days ago
```

### 3.5 Knowledge Base Documents

Long-term reference material: HowTos, runbooks, reports, meeting notes, reference documentation, or any content meant to be found and re-read later.

**Organization:** Folder-based taxonomy within the KB directory. The user defines the structure by creating subdirectories (`kb/howto/`, `kb/reports/`, `kb/meeting-notes/`). The engine does not enforce a taxonomy — folder structure is entirely user-defined.

**Type identification:** The optional `kb_type` frontmatter field identifies the document subtype. It is set by the template at creation time and changeable by the user at any time via their editor.

**CLI access:**

```
mdam kb list                    # list all KB documents
mdam kb list --type howto       # filter by KB type
mdam kb create --template howto # create new KB doc from template
```

### 3.6 Scratch Pad

A single persistent scratch pad document for ephemeral content that doesn't belong anywhere yet.

**Use cases:** Temporary code snippets, configuration values copied from an email before pasting into another application, draft LLM prompts, quick notes during a call. It's a persistent clipboard.

**Behavior:** The scratch pad is a managed document (has frontmatter, lives in the managed tree) treated as a singleton. It is always accessible via a dedicated keybinding in the TUI or `mdam scratch` from the CLI.

**What the scratch pad is not:** It is not a staging area with smart content extraction. The engine does not parse, analyze, or selectively promote scratch pad content. The scratch pad is opened in `$EDITOR` like any other document.

**Promotion flow (manual):** When the user decides something on the scratch pad is worth keeping as its own document, they trigger a "new document" action (keybinding in TUI, `mdam new` in CLI), the engine creates a new file via the template system, and the user transfers content between the two files in their editor. The engine's role is making new document creation fast and frictionless — one action to scaffold a properly named, frontmattered file in the right directory. The user handles the content.

**CLI access:**

```
mdam scratch                    # open scratch pad in $EDITOR
mdam new                        # create new document via template picker
mdam new --template howto --title "Setup Nginx" # headless creation
```

### 3.7 Template System

Templates scaffold every document type. A template is itself a markdown file with frontmatter, stored in a dedicated templates directory within the managed tree.

**Variable interpolation:** Templates contain placeholder variables resolved at document creation time:

| Variable         | Resolves to                                   |
|------------------|-----------------------------------------------|
| `{{date}}`       | Current date, ISO 8601                        |
| `{{date_short}}` | Current date, YYYY-MM-DD                      |
| `{{title}}`      | Document title (prompted or passed via flag)   |
| `{{author}}`     | Configured author name from config.yml        |
| `{{tags}}`       | Default tags for this document type           |
| `{{type}}`       | Document type identifier                      |

Custom variables can be defined in templates. The engine prompts for any unresolved variable at creation time (TUI) or requires them as flags (CLI).

**Built-in templates:** Journal entry, KB document (generic), HowTo, meeting notes, global TODO structure.

**User-defined templates:** Drop a markdown file in the templates directory. The engine discovers it on next scan — no registration step. The template filename (minus extension) becomes the template name.

**Template picker:** In the TUI, document creation presents a list of available templates. The user selects one, fills in variables, and the engine creates the file in the appropriate directory. In headless mode, the template is specified via `--template` flag.

**Implementation:** Template discovery is a directory scan. Interpolation is string replacement. Both are Phase 1 engine features, fully testable without UI.

**CLI access:**

```
mdam template list              # list available templates
mdam template show <name>       # display a template's content
```

### 3.8 Search

Active, user-initiated fuzzy search across the managed document tree.

**Search targets:** Frontmatter fields (tags, title, type, category), filenames, timestamps (date ranges), and optionally document body content (full-text).

**Ranking:** Frontmatter matches rank above body content matches. Exact tag matches rank above fuzzy title matches. Recency optionally boosts ranking (configurable).

**Implementation:** Search reads frontmatter from all managed files (fast — the engine already parses frontmatter on scan). Full-text search reads file content, which is slower but acceptable for hundreds to low-thousands of files. Fuzzy matching uses string similarity from the standard library where possible.

**CLI access:**

```
mdam search "nginx setup"           # fuzzy search across all documents
mdam search --tag devops            # exact tag match
mdam search --type kb               # filter by document type
mdam search --modified-after 2026-03-01  # date range filter
```

### 3.9 Export & Share

Strip frontmatter from a document and place a clean markdown copy in a configured export directory.

**Flow:** Select a document, trigger export. The engine reads the file, removes everything between the first and second `---` delimiters (inclusive), trims leading whitespace, and writes the clean markdown to the export directory with the same filename.

**Clipboard variant:** Copy the stripped content to the system clipboard instead of writing a file.

**CLI access:**

```
mdam export <filename>                   # export to configured directory
mdam export <filename> --to ~/Desktop    # export to specific directory
mdam export <filename> --clipboard       # copy to clipboard
```

### 3.10 Git Integration

Git provides version control and multi-device synchronization. The engine provides status awareness; the user handles actions via their preferred git workflow.

#### 3.10.1 Status Awareness

The engine runs `git status --porcelain` and `git rev-list --left-right --count HEAD...@{upstream}` on startup and after every editor return. This data is lightweight to gather and provides:

- **Per-file status:** Modified, untracked, staged — shown as indicators in the TUI file list.
- **Repository status:** Current branch name, commits ahead/behind remote, uncommitted change count — shown in the TUI status bar.
- **Stash count:** Number of stash entries, if any.

This mirrors the same data model as the user's existing `ws_status` fish function: visibility without automation.

#### 3.10.2 Lazygit Handoff

A dedicated keybinding (`ctrl+g`) suspends the TUI and opens `lazygit` in the managed tree's root directory. Same suspension/resume pattern as the `$EDITOR` handoff. On lazygit exit, the TUI resumes and re-scans.

This keeps git operations in a tool purpose-built for them. The TUI does not attempt to replicate lazygit's functionality.

#### 3.10.3 Optional Auto-Commit

For engine-initiated file mutations (TODO sweep, journal creation, import), an optional auto-commit creates a commit with a descriptive message (e.g., `"mdam: journal 2026-03-14"`, `"mdam: todo sweep, 3 tasks promoted"`). This is off by default and controlled via `git.auto_commit` in config.yml. Auto-commit never pushes — pushing is always a user action.

#### 3.10.4 Headless Git Status

The CLI exposes the same status data for use in shell scripts and existing workflows:

```
mdam status                     # show git status summary for the managed tree
mdam status --porcelain         # machine-readable output for scripting
```

---

## 4. TUI Interface

### 4.1 Navigation & Modality

The TUI uses vim-style keybindings for all navigation and actions. It does not handle text editing — that is exclusively `$EDITOR`'s domain.

**TUI modes:**

| Mode    | Purpose                                                     | Activation |
|---------|-------------------------------------------------------------|------------|
| Normal  | Navigate lists, panels, trigger actions                     | Default    |
| Command | Colon-prefixed commands (`:sync`, `:import`, `:export`)     | `:`        |
| Search  | Fuzzy find across the managed tree                          | `/`        |

There is no insert or visual mode. The TUI is a navigation and command interface, not an editor.

**Editor handoff (critical path):**

1. User presses `Enter` (or equivalent keybinding) on a document.
2. TUI dispatches a `tea.Cmd` using `tea.ExecProcess`, BubbleTea's built-in mechanism for shelling out to external processes.
3. BubbleTea suspends and fully relinquishes control of stdin/stdout.
4. `$EDITOR` opens the selected file.
5. User edits, saves, quits the editor.
6. BubbleTea resumes. The engine re-scans the directory tree and refreshes git status. The TUI repaints with updated data.

The same pattern is used for the lazygit handoff.

**Core keybindings (finalized during Phase 3–4, configurable in config.yml):**

| Key        | Action                                 |
|------------|----------------------------------------|
| `j` / `k`  | Navigate up / down                    |
| `h` / `l`  | Navigate panels / collapse-expand     |
| `Enter`    | Open in `$EDITOR`                      |
| `/`        | Search mode                            |
| `:`        | Command mode                           |
| `q`        | Quit                                   |
| `gg` / `G` | Top / bottom of list                  |
| `Tab`      | Switch panel focus                     |
| `ctrl+g`   | Open lazygit                           |
| `s`        | Open scratch pad in `$EDITOR`          |
| `?`        | Show keybinding help                   |

Additional keybindings for document operations (create, delete, move, export, TODO operations) will be defined and documented in `KEYBINDINGS.md` during development, reviewed and rationalized at the Phase 3–4 transition.

### 4.2 Ambient Findability

Passive discoverability built into the TUI layout so documents surface through browsing and recognition, not just recall and search.

**Mechanisms:**

- **Recent documents:** Last N opened/edited documents, visible in a sidebar or one keybinding away.
- **Pinned/favorites:** User-pinned documents for quick access, stored in config or a dotfile in the managed tree.
- **Category browsing:** Visual navigation of the folder tree with document preview on selection.
- **Tag browser:** Browse all tags across the managed tree, see document counts per tag, drill into a tag to list its documents.
- **Today's context:** A dashboard or landing view showing today's journal entry, open TODO count and top items, recently modified documents, and git status.
- **Smart lists:** Auto-generated filtered views — "modified this week," "untagged documents," "unsorted inbox items," "stale drafts."

### 4.3 TUI Design

**Inspirations:** lazygit (panel layout, contextual actions), atac (polish), k9s (information density), yazi (speed and preview).

**Design principles:**

- Panel-based layout with contextual sidebars.
- Live markdown preview of the selected document via `glamour`, rendered in a read-only viewport.
- Contextual help: available keybindings shown in the footer bar or accessible via `?`.
- Status bar displaying current mode, document counts, git branch, git status summary, and active filters.
- Responsive to terminal size with graceful degradation in small terminals.
- Color theming via config.yml using `lipgloss`. Ships with a default dark theme; supports community palettes (Nord, TokyoNight, Catppuccin, Dracula, etc.).
- Nerd Font icons for file type indicators, git status markers, TODO status, and navigation hints.

**Design timeline:** The TUI design (wireframes, color palette, icon selection, panel layout) is not addressed until the start of Phase 5. See section 6 for rationale.

---

## 5. Technology Stack

### Go Standard Library (Engine & Testing)

| Package      | Purpose                                                  |
|--------------|----------------------------------------------------------|
| `testing`    | Table-driven tests for all business logic                |
| `os`         | File read/write (`ReadFile`, `WriteFile`, `MkdirAll`)    |
| `os/exec`    | Editor handoff, lazygit handoff, git operations          |
| `bufio`      | Line-by-line parsing (TODO extraction, frontmatter)      |
| `strings`    | Pattern matching, frontmatter stripping, kebab-case validation |
| `regexp`     | Markdown syntax identification (`- [ ]`, `- [x]`)       |
| `filepath`   | Path manipulation, directory walking                     |
| `time`       | Timestamp generation, date comparison for sweep logic    |
| `encoding`   | YAML frontmatter parsing (evaluate stdlib feasibility first) |
| `flag`       | Fallback CLI flag parsing if Cobra is deemed too heavy   |
| `io`         | Stream handling for clipboard operations                 |
| `net/http`   | REST API calls if needed (no third-party HTTP client)    |

### Third-Party Libraries (Justified)

| Library                    | Purpose                                          | Justification                                    |
|----------------------------|--------------------------------------------------|--------------------------------------------------|
| `charmbracelet/bubbletea`  | MVU event loop, TUI framework                    | No stdlib equivalent for terminal UIs            |
| `charmbracelet/bubbles`    | UI components (list, viewport, textinput)         | Built for BubbleTea, avoids reinventing widgets  |
| `charmbracelet/lipgloss`   | Styling, layout, theming                          | Terminal CSS — no stdlib equivalent              |
| `charmbracelet/glamour`    | Markdown preview rendering in terminal            | Markdown-to-ANSI rendering is non-trivial        |
| `yuin/goldmark`            | YAML frontmatter extraction from markdown          | Robust, spec-compliant markdown parsing          |
| `spf13/cobra`              | CLI subcommands (`mdam ui`, `mdam sync`, etc.)    | Standard Go CLI framework, pairs with Viper      |
| `spf13/viper`              | Configuration file management (config.yml)         | YAML config with env var overlay, well-tested    |

### Deferred (Not v1)

| Library                    | Purpose                                          |
|----------------------------|--------------------------------------------------|
| `google/go-github`         | GitHub API integration (if/when needed)           |
| `golang.org/x/crypto/ssh`  | SSH automation (if/when needed)                   |

---

## 6. Execution Plan

### Phase 1: Headless Engine

**Goal:** Prove all core business logic with full test coverage, no UI dependency.

**Scope:**

- Configuration loading and validation (Viper, config.yml)
- Directory scanning and recursive file discovery
- Frontmatter parsing and validation (required fields, YAML structure)
- Filename validation (kebab-case, POSIX-safe)
- Import pipeline (validate, auto-fix, report issues)
- TODO extraction, sweep logic, and archive
- Template discovery and variable interpolation
- Export (frontmatter stripping)
- Git status detection (`git status --porcelain`, ahead/behind, branch name)
- CLI subcommand scaffolding (Cobra) for all engine operations

**Deliverable:** A fully tested library of pure functions, accessible via `mdam <subcommand>` CLI. Runnable via `go test` with comprehensive table-driven test suites covering edge cases.

**Testing emphasis:** Empty files, missing frontmatter, malformed YAML, files with no TODO section, TODO items with multi-line descriptions, duplicate filenames on import, non-UTF8 content, symlinks, empty directories.

### Phase 2: TUI Skeleton

**Goal:** Establish the BubbleTea framework and event loop with correct architecture.

**Scope:**

- Initialize BubbleTea with MVU model structure
- Render a basic interactive list with dummy data
- Wire core navigation keybindings (j/k, gg/G, Tab for panel switching)
- Implement command mode input (`:` prefix, basic command parsing)
- Implement search mode input (`/` prefix)
- No styling, no real data, no `$EDITOR` integration

**Deliverable:** A running TUI that responds to keybindings and displays placeholder content. Architecture is correct even though content is fake.

### Phase 3: Integration

**Goal:** Connect the tested engine to the TUI, making it a functional tool.

**Scope:**

- Hook directory scanning into BubbleTea's `Init()` function
- Display actual files from the managed tree
- Implement the template picker flow in the TUI
- Wire TODO views (global list, filtered by status/category)
- Wire search (fuzzy find over frontmatter and filenames, results in a selectable list)
- Wire import flow (validation results displayed, auto-fix/manual choice)
- Display git status in a status bar (branch, ahead/behind, change count)
- Display per-file git status indicators in file lists

**Deliverable:** A functional (but unstyled) tool that shows real data and supports all core workflows except document editing.

**Keybinding review:** At the transition from Phase 3 to Phase 4, conduct a full review of all keybindings. Document in `KEYBINDINGS.md`. Check for conflicts, ensure consistency (e.g., all "create" actions use `c` prefix, all "go to" actions use `g` prefix), and finalize the mapping.

### Phase 4: Editor Handoff

**Goal:** Seamless `$EDITOR` and lazygit suspension/resume — the single most failure-prone interaction in the application.

**Scope:**

- Implement `tea.ExecProcess` for `$EDITOR` spawning on document selection
- Implement `tea.ExecProcess` for lazygit spawning
- Verify clean stdin/stdout relinquishment and reacquisition
- Re-scan directory and refresh git status on process exit
- Test with multiple editors: Neovim, Vim, nano, VS Code (terminal mode)
- Handle edge cases: editor crash, user kills process via signal, terminal resize during external process, editor that forks to background

**Deliverable:** Rock-solid process suspension and resume. The user can open any document in their editor and return to an updated TUI without artifacts, crashes, or stale data.

### Phase 5: Polish

**Goal:** Production-quality UX and visual design.

**Design process (start of Phase 5):**

1. **Use the ugly tool for at least a week.** This reveals what information you actually need on screen, what actions you reach for most, and what layout feels natural. Design from experience, not speculation.
2. **Pick a color palette.** Start with one (TokyoNight, Nord, etc.) as the default. Design for it. Theming support comes after the default looks right.
3. **Draw ASCII wireframes.** The design medium is the terminal. Wireframes map directly to BubbleTea components and lipgloss layouts. Example:

```
┌─ Files (30%) ─────────┬─ Preview (70%) ──────────────────────┐
│ > 2026-03-14.md    [M] │                                      │
│   2026-03-13.md        │  # Daily Journal                     │
│   2026-03-12.md        │                                      │
│   setup-nginx.md   [?] │  ## TODOs                            │
│   deploy-runbook.md    │  - [ ] Review PR #42                 │
│                        │  - [x] Update DNS records            │
├─ Status ───────────────┤                                      │
│  main · 2↑ · 1 mod   │                                      │
├─ TODOs ────────────────┤                                      │
│ 3 open · 1 in-progress│                                      │
└────────────────────────┴──────────────────────────────────────┘
 NORMAL │ journal/ │ 12 docs │ /search  :cmd  ?help  ^g:lazygit
```

4. **Map wireframe panels to components.** Left pane = `bubbles/list`. Right pane = `bubbles/viewport` with `glamour` rendering. Status bar = custom lipgloss-styled component. Each panel has a clear BubbleTea model and update function.
5. **Iterate in the terminal.** Run it, adjust proportions, tweak colors, repeat. The feedback loop is faster than any external mockup tool.

**Scope:**

- `lipgloss` styling and theming implementation
- Panel layout (file list, preview, status bar, TODO summary)
- `glamour` markdown preview in viewport
- Nerd Font icons for file types, git status, TODO states
- Git status display (status bar + per-file indicators)
- Lazygit handoff keybinding
- Export feature in TUI
- Ambient findability views (recent, pinned, smart lists, tag browser, today's context)
- Fuzzy search refinement and result ranking
- Terminal resize handling
- Error states, user feedback, and loading indicators
- Color theme loading from config.yml

**Deliverable:** A polished, visually coherent TUI that feels like a peer to lazygit, yazi, and k9s.

---

## 7. Configuration

Single YAML file at `~/.config/mdam/config.yml`.

```yaml
# ── Core ─────────────────────────────────────────────────────
editor: nvim                          # $EDITOR override (falls back to $EDITOR env var)
author: "Your Name"

# ── Directories ──────────────────────────────────────────────
base_dir: ~/notes                     # root of the managed tree
export_dir: ~/Downloads               # where exported (stripped) docs go
import:
  inbox_dir: ~/notes/.inbox           # drop files here for import
  auto_fix: false                     # auto-fix validation issues without prompting

# ── Theme ────────────────────────────────────────────────────
theme: tokyonight                     # nord, dracula, catppuccin, gruvbox
nerd_fonts: false                     # true if terminal font has Nerd Font glyphs

# ── Git ──────────────────────────────────────────────────────
git:
  enabled: true
  auto_commit: false                  # auto-commit on engine mutations (sweep, journal create, import)
  lazygit: true                       # enable lazygit handoff via ctrl+g

# ── TODO ─────────────────────────────────────────────────────
todo:
  default_category: personal
  archive_after_days: 30              # archive completed tasks older than N days

# ── Journal ──────────────────────────────────────────────────
journal:
  auto_create: true                   # create today's entry on TUI startup if missing
  sweep_on_create: true               # run TODO sweep when creating a new journal entry
```

No plugin system. No extensibility beyond configuration and user-defined templates.

---

## 8. CLI Reference

All core engine features are accessible as headless subcommands. The TUI is launched via `mdam` with no subcommand (or explicitly via `mdam ui`).

```
mdam                                    # launch TUI (default)
mdam ui                                 # launch TUI (explicit)

mdam journal create [date]              # create journal entry
mdam journal list [--month YYYY-MM]     # list journal entries

mdam todo list [--status S] [--category C] [--all]
mdam todo sweep                         # run TODO sweep manually
mdam todo archive [--older-than N]      # archive old completed tasks

mdam kb list [--type T]                 # list KB documents
mdam kb create --template T --title "T" # create KB doc headlessly

mdam search "query" [--tag T] [--type T] [--modified-after D]

mdam scratch                            # open scratch pad in $EDITOR
mdam new [--template T] [--title "T"]   # create document from template

mdam import <path> [--auto-fix] [--dry-run]

mdam export <filename> [--to DIR] [--clipboard]

mdam status [--porcelain]               # git status summary for managed tree

mdam template list                      # list available templates
mdam template show <name>               # display template content

mdam config                             # show current configuration
mdam config --edit                      # open config.yml in $EDITOR
```

---

## 9. Future Considerations (Not v1)

**AI / Agent integration:** An `.ai/` directory containing application context, skill definitions, or an interface specification that allows LLM agents (Claude Code, Cursor, etc.) to query or operate on the managed tree. The v1 engine is designed to be agent-friendly by convention — consistent frontmatter, predictable paths, machine-readable metadata — so this layer can be added without engine changes. Potential use cases: "summarize my last week of journal entries," "find the KB doc about Nginx setup," "what TODOs have I been carrying for more than a week?"

**Multi-device conflict resolution:** With git as the sync mechanism, merge conflicts are possible. The engine could detect conflict markers in files and surface them in the TUI. Deferred until this becomes a real problem in practice.

**Structured TODO format:** If the flat `- [ ] text @category !priority (date)` format proves insufficient for complex task management, a more structured format (YAML task blocks, or a dedicated frontmatter section for tasks) could replace it. The sweep logic would need to adapt accordingly.

**File watchers:** Instead of re-scanning the directory on every operation, use OS-level file watchers (`fsnotify`) to detect changes. Likely unnecessary given Go's I/O speed on the expected file counts (hundreds to low-thousands), but available as an optimization if the tree grows large.

---

## 10. Document Lifecycle

```
External file ──→ [Inbox] ──→ Validation ──→ Auto-fix / Manual fix ──→ Managed tree
                                                                            │
Template picker ──→ Variables ──→ [Template Engine] ──→ New document ───────┘
                                                                            │
                                                        Managed tree ───────┘
                                                            │
                                        ┌───────────────────┼──────────────────┐
                                        │                   │                  │
                                     Journal            Knowledge           Scratch
                                        │               Base                 Pad
                                        │                   │                  │
                                   TODO sweep          Long-term           Promote
                                   ↕ Global            reference           (manual)
                                    TODO                                       │
                                        │                                      ↓
                                     Archive                            New document
                                                                        via template

── Export (strip frontmatter) ──→ Export directory / clipboard
── Git (status awareness) ──→ TUI display ──→ Lazygit handoff ──→ Remote
── $EDITOR handoff ──→ Any document ──→ Resume + re-scan
```
