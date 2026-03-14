package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AegirAexx/mdam/internal/git"
)

// View renders the current model state to a string.
// Phase 2/3: plain text layout, no lipgloss styling.
// Phase 5 will apply lipgloss styling and Nerd Font icons.
func (m Model) View() string {
	if m.showHelp {
		return m.viewHelp()
	}
	if m.mode == ModeTemplatePicker {
		return m.viewTemplatePicker()
	}
	if m.mode == ModeTemplateVars {
		return m.viewTemplateVars()
	}

	var b strings.Builder

	// Available height for content (minus status bar and separator line).
	contentHeight := m.height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Split content area: left panel (~33%) | right panel (~67%).
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 1 // 1 for separator

	left := m.renderFilePanel(leftWidth, contentHeight)
	right := m.renderPreviewPanel(rightWidth, contentHeight)

	leftLines := splitLines(left, contentHeight)
	rightLines := splitLines(right, contentHeight)

	for i := 0; i < contentHeight; i++ {
		l := ""
		if i < len(leftLines) {
			l = leftLines[i]
		}
		r := ""
		if i < len(rightLines) {
			r = rightLines[i]
		}
		b.WriteString(padRight(l, leftWidth))
		b.WriteString("│")
		b.WriteString(r)
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// renderFilePanel renders the files list panel with real document data.
func (m Model) renderFilePanel(width, height int) string {
	focused := m.activePanel == PanelFiles
	title := panelHeader("Files", focused, width)

	var lines []string
	lines = append(lines, title)

	docs := m.visibleDocs()

	if m.loading {
		lines = append(lines, "  scanning…")
		return strings.Join(lines, "\n")
	}

	if m.errorMsg != "" {
		lines = append(lines, truncate("  "+m.errorMsg, width-1))
		return strings.Join(lines, "\n")
	}

	if len(docs) == 0 {
		lines = append(lines, "  (no documents)")
		return strings.Join(lines, "\n")
	}

	visibleRows := height - 2
	start := 0
	if m.fileCursor >= visibleRows {
		start = m.fileCursor - visibleRows + 1
	}

	for i := start; i < len(docs) && len(lines) < height-1; i++ {
		cursor := "  "
		if i == m.fileCursor && focused {
			cursor = "> "
		} else if i == m.fileCursor {
			cursor = "- "
		}
		gitMarker := m.gitMarkerFor(docs[i].Path)
		name := filepath.Base(docs[i].Path)
		// Reserve space: 2 prefix + name + space + marker
		nameWidth := width - 2 - len(gitMarker) - 1
		if nameWidth < 1 {
			nameWidth = 1
		}
		line := cursor + truncate(name, nameWidth) + " " + gitMarker
		lines = append(lines, strings.TrimRight(line, " "))
	}

	return strings.Join(lines, "\n")
}

// renderPreviewPanel renders the preview/detail panel with frontmatter info.
func (m Model) renderPreviewPanel(width, height int) string {
	focused := m.activePanel == PanelPreview
	title := panelHeader("Preview", focused, width)

	var lines []string
	lines = append(lines, title)

	docs := m.visibleDocs()
	if m.fileCursor < len(docs) {
		r := docs[m.fileCursor]
		fm := r.Frontmatter
		lines = append(lines, "")
		lines = append(lines, truncate("  "+fm.Title, width-1))
		lines = append(lines, "")
		lines = append(lines, truncate(fmt.Sprintf("  type: %s", fm.Type), width-1))
		if len(fm.Tags) > 0 {
			lines = append(lines, truncate(fmt.Sprintf("  tags: %s", strings.Join(fm.Tags, ", ")), width-1))
		}
		if !fm.Modified.IsZero() {
			lines = append(lines, truncate(fmt.Sprintf("  modified: %s", fm.Modified.Format("2006-01-02")), width-1))
		}
		lines = append(lines, "")
		lines = append(lines, truncate("  "+filepath.Base(r.Path), width-1))
	}

	// TODO mini-panel at bottom of right side.
	todoStart := height - len(m.todos) - 3
	if todoStart < len(lines)+1 {
		todoStart = len(lines) + 1
	}
	for len(lines) < todoStart {
		lines = append(lines, "")
	}

	todoFocused := m.activePanel == PanelTodo
	lines = append(lines, panelHeader("TODOs", todoFocused, width))
	for i, task := range m.todos {
		cursor := "  "
		if i == m.todoCursor && todoFocused {
			cursor = "> "
		}
		if len(lines) < height-1 {
			lines = append(lines, cursor+truncate(task.Raw, width-3))
		}
	}
	if len(m.todos) == 0 && len(lines) < height-1 {
		lines = append(lines, "  (no open tasks)")
	}

	return strings.Join(lines, "\n")
}

// renderStatusBar renders the bottom status line with git info and doc count.
func (m Model) renderStatusBar() string {
	left := fmt.Sprintf(" %s ", m.mode.String())

	// Append git branch and sync info if available.
	if m.gitStatus.Branch != "" {
		left += fmt.Sprintf("│ %s ", m.gitStatus.Branch)
		if m.gitStatus.Ahead > 0 {
			left += fmt.Sprintf("↑%d ", m.gitStatus.Ahead)
		}
		if m.gitStatus.Behind > 0 {
			left += fmt.Sprintf("↓%d ", m.gitStatus.Behind)
		}
	}

	// Doc count.
	if !m.loading {
		left += fmt.Sprintf("│ %d docs ", len(m.docs))
	} else {
		left += "│ scanning… "
	}

	var right string
	switch m.mode {
	case ModeCommand:
		right = fmt.Sprintf(":%s", m.cmdInput.View())
	case ModeSearch:
		right = fmt.Sprintf("/%s", m.searchInput.View())
	default:
		if m.statusMsg != "" {
			right = m.statusMsg
		} else {
			right = "/search  :cmd  ?help  q:quit"
		}
	}

	gap := m.width - len(left) - len(right) - 1
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right + " "
}

// viewTemplatePicker renders the template selection overlay.
func (m Model) viewTemplatePicker() string {
	var b strings.Builder
	b.WriteString("\n  New Document — Select Template\n\n")
	for i, t := range m.templates {
		cursor := "  "
		if i == m.pickerCursor {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("  %s%-12s\n", cursor, t.Name))
	}
	if len(m.templates) == 0 {
		b.WriteString("  (no templates found)\n")
	}
	b.WriteString("\n  j/k to navigate, Enter to select, Esc to cancel\n")
	return b.String()
}

// viewTemplateVars renders the variable input form for a selected template.
func (m Model) viewTemplateVars() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n  New %s — Enter Details\n\n", m.pendingTmpl.Name))
	for i, name := range m.varNames {
		if i < m.varCursor {
			b.WriteString(fmt.Sprintf("  %-12s %s\n", name+":", m.varValues[i]))
		} else if i == m.varCursor {
			b.WriteString(fmt.Sprintf("  %-12s %s\n", name+":", m.varInput.View()))
		} else {
			b.WriteString(fmt.Sprintf("  %-12s\n", name+":"))
		}
	}
	b.WriteString("\n  Enter to confirm, Esc to go back\n")
	return b.String()
}

// viewHelp renders the help overlay.
func (m Model) viewHelp() string {
	var b strings.Builder
	b.WriteString("\n  Keybindings\n")
	b.WriteString("  ───────────\n\n")
	b.WriteString("  Navigation\n")
	b.WriteString("    j / k       move down / up\n")
	b.WriteString("    h / l       prev / next panel\n")
	b.WriteString("    gg / G      top / bottom\n")
	b.WriteString("    Tab         cycle panel focus\n\n")
	b.WriteString("  Modes\n")
	b.WriteString("    /           search\n")
	b.WriteString("    :           command\n")
	b.WriteString("    Esc         cancel / return to normal\n\n")
	b.WriteString("  Views\n")
	b.WriteString("    1           all documents\n")
	b.WriteString("    2           journal\n")
	b.WriteString("    3           knowledge base\n")
	b.WriteString("    4           todos\n")
	b.WriteString("    5           recent\n\n")
	b.WriteString("  Actions\n")
	b.WriteString("    Enter       open in $EDITOR (Phase 4)\n")
	b.WriteString("    g           open lazygit (Phase 4)\n")
	b.WriteString("    n           new document\n")
	b.WriteString("    s           scratch pad (Phase 4)\n")
	b.WriteString("    e           export\n")
	b.WriteString("    R           rescan\n")
	b.WriteString("    q           quit\n\n")
	b.WriteString("  Commands (:)\n")
	b.WriteString("    :todo sweep     run TODO sweep\n")
	b.WriteString("    :todo archive   archive old tasks\n")
	b.WriteString("    :q / :quit      quit\n\n")
	b.WriteString("  Press ? to close help\n")
	return b.String()
}

// gitMarkerFor returns a git status marker string for the given file path.
func (m Model) gitMarkerFor(path string) string {
	fs, ok := m.gitFileMap[path]
	if !ok {
		return ""
	}
	switch {
	case fs.IsUntracked():
		return "[?]"
	case fs.IsStaged():
		return "[A]"
	case fs.IsModified():
		return "[M]"
	default:
		return ""
	}
}

// gitMarkerForStatus returns a marker directly from a FileStatus (for testing).
func gitMarkerForStatus(fs git.FileStatus) string {
	switch {
	case fs.IsUntracked():
		return "[?]"
	case fs.IsStaged():
		return "[A]"
	case fs.IsModified():
		return "[M]"
	default:
		return ""
	}
}

// --- Layout helpers ---

// panelHeader renders a panel title line with focus indicator.
func panelHeader(title string, focused bool, width int) string {
	prefix := "─ "
	if focused {
		prefix = "▶ "
	}
	label := prefix + title + " "
	rest := width - len(label)
	if rest < 0 {
		rest = 0
	}
	return label + strings.Repeat("─", rest)
}

// padRight pads s with spaces to length n, or truncates if longer.
func padRight(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}

// truncate shortens s to maxLen, appending "…" if truncated.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return s[:maxLen-1] + "…"
}

// splitLines splits a multiline string into a slice, padding to count lines.
func splitLines(s string, count int) []string {
	lines := strings.Split(s, "\n")
	for len(lines) < count {
		lines = append(lines, "")
	}
	return lines
}
