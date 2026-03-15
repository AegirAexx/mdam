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

## Document types and destinations

| `type` | Destination directory |
|---|---|
| `journal` | `{base_dir}/journal/YYYY-MM-DD.md` |
| `kb` | `{base_dir}/kb/{kebab-title}.md` |
| `scratch` | `{base_dir}/scratch/scratch.md` |
| `todo` | `{base_dir}/todo/todo.md` |
| `unsorted` | `{base_dir}/{kebab-title}.md` |
