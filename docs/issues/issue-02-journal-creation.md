# Issue #2: Fix Journal Creation Flow

**Type:** Bug
**Priority:** High
**Labels:** `bug`, `journal`, `ux`, `phase-6`

## Problem

Creating a new journal entry via the TUI (`n` → select "journal") has three bugs:

1. **Unnecessary prompts:** The user is asked to fill in `date_short` and `date` fields manually. These should auto-resolve to today's date.
2. **Wrong filename:** The created file is named `journal.md` instead of `2026-03-15.md`.
3. **Wrong location:** The file is placed in the base directory root instead of the `journal/` subdirectory.

### Screenshots

The template picker prompts for `date_short` and `date` as if they are user-defined variables. These are built-in variables that the template engine should resolve automatically.

## Root Cause

The journal template contains `{{date_short}}` and `{{date}}` placeholders. The `template.Render` function resolves built-in variables, but the TUI's `ModeTemplateVars` flow in `model.go` checks for unresolved variables **before** `Render` is called, so it treats the date variables as user inputs.

Additionally, the `cmdCreateDoc` function does not route journal-type documents to the `journal/` subdirectory. It writes to `base_dir` directly.

## Expected Behavior

### Journal entry from TUI (`n` → "journal")

1. User presses `n`, selects "journal" from the template list.
2. **No prompts.** All variables are auto-resolved: `{{date}}` → ISO 8601 now, `{{date_short}}` → `YYYY-MM-DD` today.
3. File is created at `{base_dir}/journal/2026-03-15.md`.
4. If the journal directory doesn't exist, create it.
5. If today's entry already exists, show status message: "today's journal already exists" and open it in `$EDITOR`.

### Journal entry from CLI (`mdam journal create`)

Same behavior — this already works more correctly than the TUI path because it goes through `journal.Create()` directly rather than the template picker.

## Requirements

1. **Fix variable resolution order:** In `updateTemplateVars` or before entering `ModeTemplateVars`, call `template.Render` first to resolve built-in variables (`date`, `date_short`, `title`, `author`, `type`, `tags`). Only prompt for variables that remain unresolved after rendering.
2. **Fix file routing:** `cmdCreateDoc` must route by document type:
   - `type: journal` → `{base_dir}/journal/YYYY-MM-DD.md`
   - `type: kb` → `{base_dir}/kb/{kebab-title}.md`
   - `type: scratch` → `{base_dir}/scratch/scratch.md`
   - `type: todo` → `{base_dir}/todo/todo.md`
   - default → `{base_dir}/{kebab-title}.md`
3. **Journal shortcut:** Selecting "journal" from the template picker should bypass variable input entirely and behave identically to `mdam journal create`.

## Acceptance Criteria

- [ ] Pressing `n` → "journal" creates `journal/YYYY-MM-DD.md` with no prompts
- [ ] Frontmatter has correct date, title, type, and tags
- [ ] If today's journal exists, it opens in `$EDITOR` instead of creating a duplicate
- [ ] Journal directory is created if missing
- [ ] Other template types still prompt for unresolved custom variables
- [ ] `go test ./...` passes
