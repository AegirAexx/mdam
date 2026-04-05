# Frontmatter Contract

Every managed document requires YAML frontmatter. Field order is enforced for consistency:

```yaml
---
type: kb              # Required. One of: journal | kb | todo | scratch | unsorted — FIRST field
title: My Document    # Required. Human-readable title
tags: [devops, nginx] # Required. List of strings (may be empty [])
created: 2026-03-14   # Required. YYYY-MM-DD (emitted as YAML !!timestamp, unquoted)
modified: 2026-03-14  # Required. YYYY-MM-DD (emitted as YAML !!timestamp, unquoted)
---
```

Additional fields are passed through without validation and may appear after `modified`.

## Date format

New documents always emit `YYYY-MM-DD` as an unquoted YAML `!!timestamp`.
The parser also accepts full ISO 8601 (`2026-03-15T12:32:26Z`) for backwards compatibility.

## Template variables

Templates in `{base_dir}/.templates/` support variable interpolation at document creation time.

### Built-in variables

| Variable | Resolves to |
|---|---|
| `{{date}}` | Full ISO 8601 timestamp (`2026-04-05T10:30:00Z`) |
| `{{date_short}}` | Date only (`2026-04-05`) |
| `{{date:FORMAT}}` | Custom format using Go's [time layout](https://pkg.go.dev/time#pkg-constants) syntax |
| `{{title}}` | Document title (prompted at creation) |
| `{{author}}` | Author from config |
| `{{tags}}` | Tags (prompted at creation) |
| `{{type}}` | Document type |

### Date formatting examples

The `{{date:FORMAT}}` variable accepts any Go time layout string:

| Template variable | Output |
|---|---|
| `{{date:2006-01-02}}` | `2026-04-05` |
| `{{date:Monday - January 02 2006}}` | `Saturday - April 05 2026` |
| `{{date:Jan 2, 2006}}` | `Apr 5, 2026` |
| `{{date:15:04}}` | `10:30` |
| `{{date:2006-01-02T15:04:05Z07:00}}` | `2026-04-05T10:30:00+00:00` |

Go time layouts use the reference time `Mon Jan 2 15:04:05 MST 2006` — each component is a fixed value that tells the formatter which field to substitute.

### Precedence

Caller-supplied variables (e.g. `{{title}}`) are resolved first, then `{{date:FORMAT}}` patterns, then the fixed built-ins `{{date}}` and `{{date_short}}`. Unresolved variables remain as-is in the output.

## Document types and destinations

| `type` | Destination directory |
|---|---|
| `journal` | `{base_dir}/journal/YYYY-MM-DD.md` |
| `kb` | `{base_dir}/kb/{kebab-title}.md` |
| `scratch` | `{base_dir}/scratch/scratch.md` |
| `todo` | `{base_dir}/todo/todo.md` |
| `unsorted` | `{base_dir}/{kebab-title}.md` |
