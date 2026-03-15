# Issue #3: Fix Template Picker — Reduce to Document Types, Not Templates

**Type:** Bug / UX
**Priority:** High
**Labels:** `bug`, `ux`, `templates`, `phase-6`

## Problem

The "New Document" template picker shows 5 options:

```
journal, kb, howto, meeting, scratch
```

This list is wrong. It exposes raw template filenames instead of document types. The user should be choosing **what kind of document** they want, not which template file to use.

## Expected Behavior

The template picker should show only top-level document types:

```
journal
kb
```

### Rationale for each removed item

- **`howto`** — This is a subtype of `kb`. When the user selects "kb", they could be offered a secondary choice of KB subtypes (howto, runbook, meeting notes, reference, etc.) if multiple KB templates exist. For now, a single "kb" option using the default KB template is sufficient.
- **`meeting`** — Same as howto. It's a KB subtype. If a user wants meeting notes as a dedicated template, they can add a template file and configure it. It should not be a default top-level option.
- **`scratch`** — The scratch pad is a singleton document. It cannot be "created" — it just exists. Pressing `s` opens it (creating it if missing). It should never appear in the "new document" flow.

## Requirements

1. **Filter the template list** in `ModeTemplatePicker` to show only templates whose `type` field maps to a user-creatable document type: `journal` and `kb`.
2. **Remove scratch from template picker entirely.** The scratch pad is managed by `cmdEnsureAndOpenScratch`, not the template system.
3. **Future consideration (not this issue):** If multiple KB templates exist (howto, meeting, runbook), show a secondary picker after selecting "kb". For now, just use the first KB template found.

## Acceptance Criteria

- [ ] Template picker shows exactly 2 options: `journal` and `kb`
- [ ] Scratch pad is not listed in the template picker
- [ ] Howto and meeting templates still exist on disk but are not shown as top-level options
- [ ] `go test ./...` passes
