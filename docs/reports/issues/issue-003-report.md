# Issues #3, #2, #5 Fix Report — Template Picker, Journal Creation, README

**Branch:** `fix/issue-03-02-05`
**Report index:** 003 (third issue fix session)
**Date:** 2026-03-15
**Commits:** `b508637`, `d95e088`

---

## Problem Statements

Three issues addressed in one session on the same branch:

**Issue #3 (High, bug/ux)** — The template picker (`n` key) exposed all 5 built-in template filenames (journal, kb, howto, meeting, scratch) directly to the user. howto and meeting are KB subtypes and should not appear at the top level; scratch is a singleton managed via `s`. The user should see exactly two choices: *Journal* and *KB*.

**Issue #2 (High, bug)** — Two bugs in the new-document flow:
1. `updateTemplatePicker` called `tmpl.UnresolvedVars` on the raw template content — before `Render()` resolved built-ins. Built-in vars (`{{date_short}}`, `{{date}}`) survived as unresolved and were presented to the user as prompts they had to answer.
2. Selecting the journal template went through `ModeTemplateVars` and then `cmdCreateDoc`, which used `vars["title"]` (empty) → `t.Name` → `"journal.md"` as the filename in `cfg.BaseDir`. Journal entries should be named `YYYY-MM-DD.md` and placed in `cfg.JournalDir()`, as `journal.Create()` already handles correctly.

**Issue #5 (Medium, docs)** — README was structured for developers who already knew the codebase. An end user had no clear installation path, no quick start, and had to dig through a dense feature matrix and project structure table to find anything actionable.

---

## Changes Made

### `internal/template/template.go`

Added `TemplateType(content string) string`. Scans the first 30 lines of a template's content for a `type:` frontmatter field and returns the trimmed value. Returns `""` if not found.

```go
func TemplateType(content string) string {
    for _, line := range strings.SplitN(content, "\n", 30) {
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "type:") {
            return strings.TrimSpace(strings.TrimPrefix(trimmed, "type:"))
        }
    }
    return ""
}
```

Used by Issue #2's journal shortcut to detect the selected template's document type without re-parsing YAML.

---

### `tui/model.go`

**New field `pickerTemplates []tmpl.Template`** — holds the filtered subset shown in the picker overlay. `m.templates` continues to hold the full discovered set and is used for other lookups (future-proofing for tag-based KB subtypes).

**"n" key handler** — after populating `m.templates`, builds `m.pickerTemplates` by keeping only entries whose `Name` is `"journal"` or `"kb"`:

```go
m.pickerTemplates = nil
for _, t := range m.templates {
    if t.Name == "journal" || t.Name == "kb" {
        m.pickerTemplates = append(m.pickerTemplates, t)
    }
}
```

Filtering by name (rather than `TemplateType()`) is intentional: with 5 built-ins, howto and meeting both have `type: kb`, so type-based filtering would produce 4 entries, not 2. Name-based filtering cleanly selects the two canonical top-level types. `TemplateType` remains the right tool for detecting journal in the enter handler (below).

**`updateTemplatePicker`** — three changes:
1. Navigation bounds (`j`/`k`) and entry selection (`enter`) use `m.pickerTemplates` instead of `m.templates`.
2. **Journal shortcut:** after selecting a template, `TemplateType(m.pendingTmpl.Content)` is checked. If `"journal"`, mode is set to `ModeNormal` and `cmdJournalCreate(m.cfg)` is returned immediately — bypassing `ModeTemplateVars` entirely.
3. **Render-first for all other types:** `tmpl.Render(m.pendingTmpl, map[string]string{})` is called before `UnresolvedVars`, so built-in vars are substituted before the prompt list is built. Only genuinely user-supplied vars (like `{{title}}` in the KB template) survive.

---

### `tui/view.go`

**`viewTemplatePicker`** — iterates `m.pickerTemplates` instead of `m.templates`. Header text changed from "Select Template" to "Select Type" to reflect that the user is choosing a document type, not a template filename.

---

### `tui/commands.go`

Added `cmdJournalCreate(cfg config.Config) tea.Cmd`. Delegates to `journal.Create(cfg.JournalDir(), cfg.TemplatesDir(), time.Now())`. On success returns `scratchReadyMsg{path: path}` — this reuses the existing handler which opens the file in `$EDITOR` and triggers a rescan on return via `editorReturnMsg`. On error returns `fileCreatedMsg{err: ...}`.

The `scratchReadyMsg` reuse pattern is intentional: the journal flow (create if absent → open in editor) is identical to the scratch pad flow. No new message type needed.

Required adding `"github.com/AegirAexx/mdam/internal/journal"` to imports.

---

### `README.md` + new docs files (Issue #5)

README fully rewritten with end-user framing:

1. **Header + philosophy** — what it is and the design philosophy (two sentences)
2. **Installation** — prerequisites, build, install to PATH, first-run note
3. **Quick Start** — TUI wireframe, condensed keybindings table, 5 CLI examples
4. **Configuration** — minimal `config.yml` example with most-used keys
5. **Features** — user-benefit list, not implementation details
6. **Links table** → four new sub-docs + existing KEYBINDINGS, HANDOFF, spec

Four new documentation files extracted from the old README:

| File | Contents extracted |
|---|---|
| `docs/FRONTMATTER.md` | Frontmatter contract, field table, type/destination matrix |
| `docs/TODO-FORMAT.md` | Task syntax, field reference, sweep and archive behaviour |
| `docs/CLI.md` | Full subcommand reference |
| `docs/DEVELOPMENT.md` | Project structure, package map, code style rules, test rules, dep table |

No content was deleted — only reorganised. The README went from ~290 lines to ~115 lines.

---

## Tests Added / Updated

### `internal/template/template_test.go`

`TestTemplateType` — 10 table-driven cases covering: journal/kb/scratch, absent `type:` field, empty content, leading/trailing spaces, and all relevant built-in template variables (`journalTemplate`, `kbTemplate`, `howtoTemplate`, `scratchTemplate`).

### `tui/model_test.go`

`TestTemplatePickerNavigation` — updated to set `m.pickerTemplates` instead of `m.templates` (navigation bounds now read from `pickerTemplates`).

Five new tests:

| Test | What it asserts |
|---|---|
| `TestNKeyFiltersPickerToJournalAndKB` | With 5 built-ins loaded, `n` key sets `len(pickerTemplates) == 2` containing only "journal" and "kb" |
| `TestPickerViewShowsTypeHeader` | `viewTemplatePicker` output contains "Select Type" |
| `TestJournalTemplateSelectionBypassesVarMode` | Selecting journal enters `ModeNormal` (not `ModeTemplateVars`) and returns a non-nil cmd |
| `TestKBTemplateSelectionEntersVarMode` | Selecting kb (which has `{{title}}`) enters `ModeTemplateVars` |
| `TestKBTemplateNoBuiltinVarsPrompted` | `varNames` after kb selection does not contain `"date_short"` or `"date"` |

---

## Verification

```
go test ./...   → all 12 packages pass
go vet ./...    → clean
```

All internal links in the new README resolve to existing files.
