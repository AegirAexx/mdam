# TODO Task Format

Tasks live in `{base_dir}/todo/todo.md`. Each task is a single line:

```
- [ ] Task text @category !priority (YYYY-MM-DD) ✓YYYY-MM-DD
- [x] Completed task @work (2026-03-12) ✓2026-03-13
```

## Fields

| Field | Syntax | Description |
|---|---|---|
| Status | `- [ ]` / `- [x]` | Open or completed |
| Text | free text | The task description |
| Category | `@word` | Optional grouping label |
| Priority | `!high` / `!medium` / `!low` | Optional priority flag |
| Created date | `(YYYY-MM-DD)` | Optional creation date |
| Completed date | `✓YYYY-MM-DD` | Set automatically on completion |

## Example tasks

```
- [ ] Review PR #42 @work !high (2026-03-14)
- [ ] Buy groceries @personal
- [x] Update DNS records @work (2026-03-12) ✓2026-03-13
```

## TODO sweep

The sweep command (`mdam todo sweep` or `:todo sweep` in TUI) reads unchecked tasks
from yesterday's journal entry and appends them to `todo/todo.md`.

## Archive

`mdam todo archive` (or `:todo archive` in TUI) moves completed tasks older than
`todo.archive_after_days` (default 30) to `{base_dir}/todo/archive.md`.
