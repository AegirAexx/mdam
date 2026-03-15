# Issue #1 Fix Report — First-Run Initialization & Scaffolding

**Issue:** [#1 First-Run Initialization & Scaffolding](../issues/issue-01-initialization.md)
**Branch:** `fix/issue-01`
**Status:** Complete

---

## Problem

MadaM assumed the config file, base directory, and subfolder structure all existed before launch. A fresh `./mdam` with no prior setup produced undefined behavior: missing directories caused crashes and no guidance was given to the user.

Three additional bugs were discovered during investigation — path helper methods in `internal/config/config.go` returned wrong paths that were inconsistent with the spec and `CLAUDE.md`.

---

## Changes Made

### 1. `internal/config/config.go` — Path bug fixes + default changes

Four path methods were returning wrong paths. Fixed to match the spec:

| Method | Before | After |
|---|---|---|
| `TemplatesDir()` | `{base}/templates` | `{base}/.templates` |
| `TodoPath()` | `{base}/todo.md` | `{base}/todo/todo.md` |
| `ScratchPath()` | `{base}/scratch.md` | `{base}/scratch/scratch.md` |
| `ArchivePath()` | `{base}/archive.md` | `{base}/todo/archive.md` |

Two new helpers added to support the corrected structure:
- `TodoDir() string` — returns `{base}/todo`
- `ScratchDir() string` — returns `{base}/scratch`

Default values updated:
- `base_dir` default: `~/notes` → `""` (empty — triggers first-run detection)
- `import.inbox_dir` default: `~/notes/.inbox` → `""` (derived at runtime)

### 2. `internal/config/config_test.go` — Test updates

`TestConfigDirs` updated to assert the corrected paths. Assertions added for the two new helpers (`TodoDir`, `ScratchDir`). New test `TestLoadDefaultsBaseDir` verifies that the `base_dir` default is now an empty string.

### 3. `internal/setup/setup.go` — New package

New `setup` package with pure functions, no global state:

| Function | Description |
|---|---|
| `IsFirstRun(cfgPath, cfg)` | Returns true if config is missing or `base_dir` is empty/non-existent |
| `WriteDefaultConfig(path)` | Creates parent dirs and writes a commented `config.yml`; no-op if file exists |
| `ValidateConfig(cfg)` | Returns human-readable warnings for invalid config values (e.g. unknown theme) |
| `PromptBaseDir(r, w)` | Prompts for base directory, expands `~`, defaults to `~/notes` |
| `ScaffoldDirs(baseDir)` | Creates all 6 subdirs (`journal/`, `kb/`, `todo/`, `scratch/`, `.inbox/`, `.templates/`); idempotent |
| `SeedTemplates(dir)` | Delegates to `template.WriteBuiltins` to copy built-in templates; skips existing files |
| `EnsureScratch(path)` | Creates `scratch/scratch.md` with valid frontmatter; no-op if file exists |
| `Run(cfgPath, cfg, r, w)` | Orchestrates the full first-run flow; returns updated config |

Three internal helpers support the above: `promptYN`, `updateConfigBaseDir` (line-by-line YAML rewrite, no new deps), and `expandHome`.

The `defaultConfigYAML` constant contains a fully commented config template that is written on first run.

### 4. `internal/setup/setup_test.go` — New test file

Table-driven tests covering all exported functions:

| Test | Cases |
|---|---|
| `TestIsFirstRun` | missing file + empty BaseDir → true; file exists + valid BaseDir → false; non-existent BaseDir → true |
| `TestWriteDefaultConfig` | creates file with parent dirs; second call is no-op |
| `TestUpdateConfigBaseDir` | patches `base_dir` line; preserves other fields |
| `TestValidateConfig` | valid theme → 0 warnings; unknown theme → 1 warning; empty theme → 0 warnings |
| `TestPromptBaseDir` | empty input → `~/notes` expanded; tilde path → expanded; absolute path → unchanged |
| `TestScaffoldDirs` | all 6 dirs created; second call no error |
| `TestSeedTemplates` | `.md` files written; second call unchanged |
| `TestEnsureScratch` | file created with `type: scratch`; second call does not overwrite |
| `TestRun` | integration: path input + `y` → config written + dirs scaffolded + scratch created |

### 5. `internal/cli/root.go` — First-run wired into startup

`initConfig` replaced with a version that:
1. Resolves the config path (flag or default).
2. Loads config, falling back to empty defaults on error.
3. Calls `setup.IsFirstRun` — if true, runs `setup.Run` to guide the user.
4. On subsequent runs, calls `setup.ValidateConfig` and prints any warnings to stderr.

### 6. `tui/commands.go` — Hardened scratch creation

`cmdEnsureAndOpenScratch` now calls `os.MkdirAll(filepath.Dir(path), 0o755)` before `os.WriteFile`. This handles the case where the user skips the guided setup: the `scratch/` subdirectory is created on demand rather than failing silently.

---

## Acceptance Criteria — Verification

| Criterion | Result |
|---|---|
| `mdam` with no config produces a valid commented `config.yml` | `WriteDefaultConfig` + `Run` |
| Empty `base_dir` prompts for a path | `PromptBaseDir` in `Run` |
| Folder scaffolding creates all expected subdirectories | `ScaffoldDirs` creates 6 dirs |
| Built-in templates copied to `.templates/` | `SeedTemplates` → `template.WriteBuiltins` |
| Scratch pad created with valid frontmatter | `EnsureScratch` |
| All initialization is idempotent | Every function is a no-op on second call |
| Existing files and directories are never overwritten | `os.Stat` guards throughout |
| `go test ./...` passes | All 12 packages pass, `go vet` clean |
