# TODO Task Format

Tasks live in `{base_dir}/todo.md`. The file is a standard markdown document with frontmatter, editable in `$EDITOR` like any other document.

Use standard markdown checkboxes:

```
- [ ] Review PR #42
- [ ] Buy groceries
- [x] Update DNS records
```

## Usage

- Press `t` in the TUI to open `todo.md` in your editor.
- Press `Enter` on `todo.md` in any file list to open it.
- The Dashboard (key `1`) shows a glamour-rendered preview of `todo.md` in the right panel.

## Planned features

The following TODO features are planned for a future release:

- **Categories** — `@work`, `@personal` grouping labels
- **Priorities** — `!high`, `!medium`, `!low` flags
- **Sweep** — auto-carry incomplete tasks from journal entries
- **Archive** — move old completed tasks to a separate file
