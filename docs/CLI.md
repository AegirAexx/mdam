# CLI Reference

Run `mdam --help` or `mdam <subcommand> --help` for flags and options.

## Default (TUI)

```
mdam                                      Launch interactive TUI
```

## Journal

```
mdam journal create [date]                Create journal entry (today if no date)
mdam journal list [--month YYYY-MM]       List journal entries
```

## TODO

```
mdam todo list [--status S] [--category C] [--all]
mdam todo sweep                           Sweep yesterday's journal into todo.md
mdam todo archive [--older-than N]        Archive completed tasks older than N days
```

## Search

```
mdam search "query" [--tag T] [--type T] [--modified-after D]
```

## Import / Export

```
mdam import <path> [--auto-fix] [--dry-run]
mdam export <file> [--to DIR]
```

## Git status

```
mdam status [--porcelain]                 Git status summary for the managed tree
```

## Templates

```
mdam template list                        List available templates
mdam template show <name>                 Display template content
```

## Config

```
mdam config                               Show current configuration
mdam config --edit                        Open config.yml in $EDITOR
```
