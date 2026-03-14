# Phase 4 Report — Editor Handoff

**Completed:** 2026-03-14
**Goal:** Implement the `tea.ExecProcess` suspend/resume pattern for `$EDITOR` and lazygit. The TUI relinquishes stdin/stdout to an external process, waits for it to exit, and re-scans on return. This is the single most failure-prone interaction in the application.

---

## Summary

Phase 4 is complete. Pressing `Enter` on any document suspends the TUI and opens it in `$EDITOR`. Pressing `s` opens the scratch pad (creating it with valid frontmatter if it does not exist). Pressing `ctrl+g` opens lazygit rooted at `cfg.BaseDir`. When any external process exits, the TUI resumes and immediately fires a full re-scan of documents, git status, and TODOs.

**10 new tests. 64 total TUI tests passing. All 11 packages pass. `go vet` clean.**

---

## What Was Built

### Modified Files

| File | Changes |
|---|---|
| `tui/messages.go` | Added `editorReturnMsg`, `scratchReadyMsg` |
| `tui/commands.go` | Added `resolveEditor`, `cmdOpenEditor`, `cmdOpenLazygit`, `cmdEnsureAndOpenScratch`; added `os/exec` import |
| `tui/model.go` | Wired `Enter`, `s`, `ctrl+g`; added `editorReturnMsg` and `scratchReadyMsg` handlers |
| `tui/model_test.go` | Added `ctrl+g` to `sendKey` helper; added 10 Phase 4 tests |

---

## Architecture

### The ExecProcess Pattern

`tea.ExecProcess` is BubbleTea's built-in mechanism for external process handoff. It fully suspends the TUI event loop — relinquishing stdin, stdout, and the terminal's raw mode — runs the given `*exec.Cmd` attached to the real terminal, and sends a callback message when the process exits.

```go
func cmdOpenEditor(path, editor string) tea.Cmd {
    c := exec.Command(editor, path)
    return tea.ExecProcess(c, func(err error) tea.Msg {
        return editorReturnMsg{err: err}
    })
}
```

The same pattern is used for lazygit:

```go
func cmdOpenLazygit(dir string) tea.Cmd {
    c := exec.Command("lazygit", "-p", dir)
    return tea.ExecProcess(c, func(err error) tea.Msg {
        return editorReturnMsg{err: err}
    })
}
```

Both share `editorReturnMsg` — the resume handler does not need to know which process exited.

### Resume and Re-scan

On any `editorReturnMsg`, the model sets `loading = true` and batches three async commands:

```
editorReturnMsg → tea.Batch(
    cmdLoadDocs(cfg.BaseDir)      → docsLoadedMsg
    cmdLoadGitStatus(cfg.BaseDir) → gitStatusMsg
    cmdLoadTodos(cfg.TodoPath())  → todosLoadedMsg
)
```

The file panel immediately shows `scanning…` and repaints with fresh data as each command returns. Documents edited outside the TUI (including by the editor just closed) appear correctly on resume without any manual `R` rescan.

### New Message Types

| Message | Sent when |
|---|---|
| `editorReturnMsg` | `$EDITOR` or lazygit process exits |
| `scratchReadyMsg` | Scratch pad file confirmed to exist on disk |

### resolveEditor

```go
func resolveEditor(cfgEditor string) string {
    if cfgEditor != "" {
        return cfgEditor
    }
    return os.Getenv("EDITOR")
}
```

Precedence: `config.yml editor:` → `$EDITOR` env var → empty string. Empty string is an error condition: the TUI shows `"no editor configured ($EDITOR not set)"` and does not attempt to open anything. No hardcoded fallback binary.

---

## Features Implemented

### Enter — Open in $EDITOR

Pressing `Enter` on a selected document in the file panel opens it in the configured editor:

```
Enter
  → resolveEditor(cfg.Editor)
  → cmdOpenEditor(selectedDoc, editor)     [tea.ExecProcess]
  → user edits, saves, quits
  → editorReturnMsg{}
  → loading = true, batch rescan
```

Error cases handled:
- No document under cursor → `"no document selected"` in status bar
- No editor configured → `"no editor configured ($EDITOR not set)"`

### s — Scratch Pad

Pressing `s` opens the scratch pad. Because the scratch file may not exist on first use, creation is handled asynchronously before the editor is opened:

```
s
  → cmdEnsureAndOpenScratch(cfg)           [goroutine]
      if scratch.md missing:
          create with frontmatter (type: scratch)
      → scratchReadyMsg{path}
  → cmdOpenEditor(path, editor)            [tea.ExecProcess]
  → user edits
  → editorReturnMsg{}
  → batch rescan
```

The generated frontmatter is minimal but valid:

```yaml
---
title: Scratch Pad
tags: []
created: 2026-03-14T09:00:00Z
modified: 2026-03-14T09:00:00Z
type: scratch
---
```

### ctrl+g — Lazygit

Pressing `ctrl+g` opens lazygit rooted at `cfg.BaseDir`, subject to the `git.lazygit` config flag:

```
ctrl+g
  → cfg.Git.Lazygit check
  → cmdOpenLazygit(cfg.BaseDir)            [tea.ExecProcess]
  → user commits, pushes, etc.
  → editorReturnMsg{}
  → batch rescan (picks up updated git status)
```

If `git.lazygit = false` in config, the key shows `"lazygit disabled (git.lazygit = false)"` and does nothing.

### Keybinding Decision: ctrl+g vs g

The spec lists both `g` (lazygit) and `gg` (jump to top). These conflict: the existing chord detection holds state after the first `g` press, making it impossible to reliably distinguish a single `g` from the start of `gg` without a timer.

Phase 4 resolves this by using `ctrl+g` for lazygit, keeping the `gg` chord fully intact. This is a pragmatic decision noted for review at the Phase 5 keybinding pass.

---

## Update Dispatch — Phase 4 Additions

```
Update(msg)
├── editorReturnMsg   → set loading=true, batch rescan (docs + git + todos)
├── scratchReadyMsg   → resolveEditor → cmdOpenEditor
└── tea.KeyMsg (ModeNormal)
    ├── Enter         → resolveEditor → cmdOpenEditor(selectedDoc)
    ├── s             → cmdEnsureAndOpenScratch(cfg)
    └── ctrl+g        → cmdOpenLazygit(cfg.BaseDir)  [if cfg.Git.Lazygit]
```

---

## Test Coverage

10 new test functions in `tui/model_test.go`.

| Test | What it verifies |
|---|---|
| `TestResolveEditor` | Config takes precedence over `$EDITOR`; empty when neither set |
| `TestEditorReturnTriggersRescan` | `editorReturnMsg` sets `loading=true` and returns a batch command |
| `TestEditorReturnErrorSetsStatus` | `editorReturnMsg` with error sets `"editor error: …"` in status bar |
| `TestEnterNoDocSelected` | `Enter` with empty file list sets `"no document selected"` |
| `TestEnterNoEditorConfigured` | `Enter` with no editor set shows `"no editor configured"` |
| `TestEnterWithDocReturnsCmd` | `Enter` with a selected doc and editor returns a non-nil `tea.Cmd` |
| `TestScratchReadyMsgOpensEditor` | `scratchReadyMsg` with editor configured returns `cmdOpenEditor` |
| `TestScratchReadyMsgNoEditor` | `scratchReadyMsg` with no editor shows error, returns nil cmd |
| `TestLazygitDisabled` | `ctrl+g` with `Git.Lazygit=false` shows `"disabled"` in status bar |
| `TestLazygitEnabledReturnsCmd` | `ctrl+g` with `Git.Lazygit=true` returns a non-nil `tea.Cmd` |

`sendKey` helper extended with `"ctrl+g"` → `tea.KeyMsg{Type: tea.KeyCtrlG}`.

---

## Phase 4 Checklist (from Phase 3 report)

- [x] `Enter` — editor handoff via `tea.ExecProcess($EDITOR, selectedDoc)`
- [x] `s` — scratch pad open in `$EDITOR` (auto-create with frontmatter if missing)
- [x] `ctrl+g` — lazygit handoff via `tea.ExecProcess("lazygit", cfg.BaseDir)`
- [x] Re-scan on editor/lazygit exit (docs + git status + todos)
- [ ] `d` — delete selected document with confirmation (deferred to Phase 5)

---

## Phase 5 Handoff

Phase 5 is polish: lipgloss styling, glamour preview, Nerd Font icons, color theming, ambient findability views, and terminal resize handling. The engine and TUI wiring are complete.

Outstanding items from earlier phases to address in Phase 5:

- **`d` delete** — requires a confirmation overlay; deferred because it needs UI work
- **`gg` vs `g` keybinding** — finalize lazygit key after real usage informs the decision
- **Keybinding review** — full pass over `KEYBINDINGS.md` as called for in the spec
- **Glamour preview** — replace the frontmatter text preview with a rendered markdown viewport
- **Ambient findability** — recent docs sidebar, tag browser, today's context dashboard
- **Color theming** — load palette from `cfg.Theme`, apply via lipgloss
