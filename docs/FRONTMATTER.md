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

### How Go date layouts work

Go does not use `%d`-style format codes. Instead, it uses a **reference time** where each component has a specific magic value:

```
Mon Jan 2 15:04:05 MST 2006
```

| Value | Meaning |
|---|---|
| `2006` | Year |
| `01` | Month (zero-padded number) |
| `1` | Month (no padding) |
| `January` | Month (full name) |
| `Jan` | Month (short name) |
| `02` | Day of month (zero-padded) |
| `2` | Day of month (no padding) |
| `Monday` | Weekday (full name) |
| `Mon` | Weekday (short name) |
| `15` | Hour (24h) |
| `3` | Hour (12h) |
| `04` | Minute |
| `05` | Second |

To format a date, write the layout using these exact values. For example, `Monday - January 02 2006` produces `Sunday - April 06 2026`. Using `01` where you mean the day will output the month number instead.

### Frontmatter date fields

The `created` and `modified` fields in frontmatter must use `{{date_short}}` in templates. This produces the `YYYY-MM-DD` format that mdam expects. Do not use `{{date}}` or `{{date:FORMAT}}` for these fields — they will produce formats the frontmatter parser does not recognise as dates.

```yaml
created: {{date_short}}    # correct — produces 2026-04-06
modified: {{date_short}}   # correct — produces 2026-04-06
created: {{date}}          # wrong — produces full ISO 8601 timestamp
```

### Precedence

Caller-supplied variables (e.g. `{{title}}`) are resolved first, then `{{date:FORMAT}}` patterns, then the fixed built-ins `{{date}}` and `{{date_short}}`. Unresolved variables remain as-is in the output.

## Document types and destinations

| `type` | Destination directory |
|---|---|
| `journal` | `{base_dir}/journal/YYYY-MM-DD.md` |
| `kb` | `{base_dir}/kb/{kebab-title}.md` |
| `scratch` | `{base_dir}/scratch.md` |
| `todo` | `{base_dir}/todo.md` |
| `unsorted` | `{base_dir}/{kebab-title}.md` |
