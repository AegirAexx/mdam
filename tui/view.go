package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/AegirAexx/mdam/internal/git"
)

// contentHeight returns the number of terminal rows available for pane content,
// accounting for the tab bar (1 line), the separator (1 line), and the status bar (1 line).
func (m Model) contentHeight() int {
	h := m.height - 3
	if h < 1 {
		h = 1
	}
	return h
}

// renderTabBar renders a one-line tab bar showing all 4 panes.
// The active pane tab is rendered with inverted colors.
func (m Model) renderTabBar() string {
	type tabDef struct {
		view  View
		label string
	}
	tabs := []tabDef{
		{ViewDashboard, "1: Dashboard"},
		{ViewJournal, "2: Journal"},
		{ViewKB, "3: KB"},
		{ViewTags, "4: Tag Browser"},
	}

	var parts []string
	for _, t := range tabs {
		if m.activeView == t.view {
			parts = append(parts, lipgloss.NewStyle().Reverse(true).Padding(0, 1).Render(t.label))
		} else {
			parts = append(parts, m.theme.Subtle.Padding(0, 1).Render(t.label))
		}
	}
	line := strings.Join(parts, " ")
	return lipgloss.NewStyle().Width(m.width).Render(line)
}

// renderReadMode renders the full-screen glamour read overlay.
// Layout: document title header (1 line) | viewport | status bar (1 line).
func (m Model) renderReadMode() string {
	var b strings.Builder
	title := m.readDocTitle
	if title == "" {
		title = "Document"
	}
	b.WriteString(m.theme.Accent.Width(m.width).Render(" " + title))
	b.WriteString("\n")
	b.WriteString(m.readViewport.View())
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())
	return b.String()
}

// View renders the current model state to a string.
func (m Model) View() string {
	if m.mode == ModeRead {
		return m.renderReadMode()
	}
	if m.showHelp {
		return m.viewHelp()
	}
	if m.mode == ModeTemplatePicker {
		return m.viewTemplatePicker()
	}
	if m.mode == ModeTemplateVars {
		return m.viewTemplateVars()
	}
	if m.activeView == ViewDashboard {
		return m.renderDashboard()
	}
	if m.activeView == ViewJournal {
		return m.renderTabBar() + "\n" + m.renderJournalView()
	}
	if m.activeView == ViewKB {
		return m.renderTabBar() + "\n" + m.renderKBView()
	}
	if m.activeView == ViewTags {
		return m.renderTagBrowser()
	}

	var b strings.Builder

	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

	contentHeight := m.contentHeight()

	// Split content area: left panel (~33%) | right panel (~67%).
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 1 // 1 for separator

	left := m.renderFilePanel(leftWidth, contentHeight)
	right := m.renderPreviewPanel(rightWidth, contentHeight)

	leftLines := splitLines(left, contentHeight)
	rightLines := splitLines(right, contentHeight)

	divider := m.theme.Divider.Render("│")
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
		b.WriteString(divider)
		b.WriteString(r)
		b.WriteString("\n")
	}

	b.WriteString(m.theme.Divider.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// renderFilePanel renders the files list panel with real document data.
func (m Model) renderFilePanel(width, height int) string {
	focused := m.activePanel == PanelFiles
	title := styledPanelHeader("Files", focused, width, m.theme, m.icons)

	var lines []string
	lines = append(lines, title)

	docs := m.visibleDocs()

	if m.loading {
		frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
		lines = append(lines, "  "+frame+" scanning…")
		return strings.Join(lines, "\n")
	}

	if m.errorMsg != "" {
		lines = append(lines, m.theme.GitModified.Render(truncate("  "+m.errorMsg, width-1)))
		return strings.Join(lines, "\n")
	}

	if len(docs) == 0 {
		lines = append(lines, "  (no documents)")
		return strings.Join(lines, "\n")
	}

	visibleRows := height - len(lines) - 1
	if visibleRows < 1 {
		visibleRows = 1
	}
	start := 0
	if m.fileCursor >= visibleRows {
		start = m.fileCursor - visibleRows + 1
	}

	for i := start; i < len(docs) && len(lines) < height-1; i++ {
		doc := docs[i]
		selected := i == m.fileCursor

		// Build the name and markers.
		name := filepath.Base(doc.Path)
		pinMarker := ""
		if m.pinnedPaths[doc.Path] {
			pinMarker = " [*]"
		}
		gitMarker := ""
		if g := m.gitMarkerStyled(doc.Path); g != "" {
			gitMarker = " " + g
		}

		nameWidth := width - 2 - len(pinMarker) - lipgloss.Width(gitMarker)
		if nameWidth < 1 {
			nameWidth = 1
		}

		itemText := " " + truncate(name, nameWidth) + pinMarker + gitMarker

		var line string
		if selected && focused {
			// Full-row inverted highlight (§2 focus indicator).
			line = lipgloss.NewStyle().Reverse(true).Width(width).Render(itemText)
		} else if m.pinnedPaths[doc.Path] {
			line = m.theme.FilePinned.Render(itemText)
		} else {
			line = m.theme.FileNormal.Render(itemText)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderPreviewPanel renders the preview/detail panel.
func (m Model) renderPreviewPanel(width, height int) string {
	focused := m.activePanel == PanelPreview

	viewportAvail := height - 1 // minus panel header
	if viewportAvail < 2 {
		viewportAvail = 2
	}

	// For tree views (Journal/KB), use the pane-specific selected path.
	if m.activeView == ViewJournal || m.activeView == ViewKB {
		selectedPath := m.selectedDoc()
		panelTitle := "Preview"
		if selectedPath != "" {
			for _, d := range m.docs {
				if d.Path == selectedPath {
					if d.Frontmatter.Title != "" {
						panelTitle = d.Frontmatter.Title
					}
					break
				}
			}
		}
		title := styledPanelHeader(panelTitle, focused, width, m.theme, m.icons)
		var lines []string
		lines = append(lines, title)
		if selectedPath == "" {
			lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
				m.theme.Muted.Render("Select a document to preview."),
			))
			return strings.Join(lines, "\n")
		}
		if m.preview.Width > 0 && m.preview.TotalLineCount() > 0 {
			vpLines := strings.Split(m.preview.View(), "\n")
			for i, l := range vpLines {
				if i >= viewportAvail {
					break
				}
				lines = append(lines, l)
			}
		}
		return strings.Join(lines, "\n")
	}

	// ViewAll / ViewRecent: use fileCursor into visibleDocs.
	docs := m.visibleDocs()
	panelTitle := "Preview"
	if m.fileCursor < len(docs) {
		if t := docs[m.fileCursor].Frontmatter.Title; t != "" {
			panelTitle = t
		}
	}
	title := styledPanelHeader(panelTitle, focused, width, m.theme, m.icons)

	var lines []string
	lines = append(lines, title)

	if m.fileCursor >= len(docs) {
		lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
			m.theme.Muted.Render("Select a document to preview."),
		))
		return strings.Join(lines, "\n")
	}

	if m.fileCursor < len(docs) {
		r := docs[m.fileCursor]
		fm := r.Frontmatter

		if m.preview.Width > 0 && m.preview.TotalLineCount() > 0 {
			// Glamour-rendered viewport content.
			vpLines := strings.Split(m.preview.View(), "\n")
			for i, l := range vpLines {
				if i >= viewportAvail {
					break
				}
				lines = append(lines, l)
			}
		} else {
			// Fallback: show styled frontmatter metadata.
			lines = append(lines, "")
			lines = append(lines, m.theme.PreviewTitle.Render(truncate("  "+fm.Title, width-1)))
			lines = append(lines, "")
			lines = append(lines, m.theme.PreviewKey.Render("  type: ")+m.theme.PreviewMeta.Render(fm.Type))
			if len(fm.Tags) > 0 {
				lines = append(lines, m.theme.PreviewKey.Render("  tags: ")+
					m.theme.PreviewMeta.Render(truncate(strings.Join(fm.Tags, ", "), width-10)))
			}
			if !fm.Modified.IsZero() {
				lines = append(lines, m.theme.PreviewKey.Render("  modified: ")+
					m.theme.PreviewMeta.Render(fm.Modified.Format("2006-01-02")))
			}
			lines = append(lines, "")
			lines = append(lines, m.theme.PreviewMeta.Render(truncate("  "+filepath.Base(r.Path), width-1)))
		}
	}

	return strings.Join(lines, "\n")
}

// renderStatusBar renders the bottom status line with mode, git info, and hints.
func (m Model) renderStatusBar() string {
	// Mode indicator (styled by current mode).
	var modeStyle lipgloss.Style
	switch m.mode {
	case ModeCommand:
		modeStyle = m.theme.StatusCommand
	case ModeSearch:
		modeStyle = m.theme.StatusSearch
	case ModeTemplatePicker, ModeTemplateVars:
		modeStyle = m.theme.StatusNewDoc
	case ModeDeleteConfirm:
		modeStyle = m.theme.StatusCommand // reuse yellow for danger
	default:
		modeStyle = m.theme.StatusNormal
	}
	modeStr := modeStyle.Render(m.mode.String())

	left := modeStr + " "

	// Git branch and sync info.
	if m.gitStatus.Branch != "" {
		gitStr := m.gitStatus.Branch
		if m.gitStatus.Ahead > 0 {
			gitStr += fmt.Sprintf(" ↑%d", m.gitStatus.Ahead)
		}
		if m.gitStatus.Behind > 0 {
			gitStr += fmt.Sprintf(" ↓%d", m.gitStatus.Behind)
		}
		left += m.theme.StatusGit.Render("│ "+gitStr+" ")
	}

	// Doc count / loading state.
	if m.loading {
		frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
		left += m.theme.StatusInfo.Render("│ " + frame + " scanning… ")
	} else {
		j, k, s := docCounts(m.docs)
		left += m.theme.StatusInfo.Render(fmt.Sprintf("│ %d journal · %d kb · %d scratch ", j, k, s))
	}

	// File path of highlighted document (centre section).
	if !m.loading {
		if rel := highlightedRelPath(m); rel != "" {
			left += m.theme.StatusInfo.Render("│ " + truncate(rel, 40) + " ")
		}
	}

	// Right side: mode input or status/hint.
	var right string
	switch m.mode {
	case ModeCommand:
		right = m.theme.StatusMsg.Render(":") + m.cmdInput.View()
	case ModeSearch:
		right = m.theme.StatusMsg.Render("/") + m.searchInput.View()
	case ModeDeleteConfirm:
		right = m.theme.Warning.Render(fmt.Sprintf("Delete %q? (y/n)", m.deleteConfirmTitle))
	default:
		if m.statusMsg != "" {
			right = m.theme.StatusMsg.Render(m.statusMsg)
		} else {
			right = m.theme.Muted.Render("/  :  o:read  ?  q")
		}
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	gap := m.width - leftWidth - rightWidth - 1
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right + " "
}

// viewTemplatePicker renders the template selection overlay.
func (m Model) viewTemplatePicker() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.theme.DashboardHeader.Render("  New Document — Select Type"))
	b.WriteString("\n\n")
	for i, t := range m.pickerTemplates {
		if i == m.pickerCursor {
			b.WriteString(lipgloss.NewStyle().Reverse(true).Render(fmt.Sprintf("  %-14s", t.Name)))
		} else {
			b.WriteString(m.theme.FileNormal.Render(fmt.Sprintf("  %-14s", t.Name)))
		}
		b.WriteString("\n")
	}
	if len(m.pickerTemplates) == 0 {
		b.WriteString(m.theme.PreviewMeta.Render("  (no templates found)"))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(m.theme.StatusHint.Render("  j/k to navigate, Enter to select, Esc to cancel"))
	b.WriteString("\n")
	return b.String()
}

// viewTemplateVars renders the variable input form for a selected template.
func (m Model) viewTemplateVars() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.theme.DashboardHeader.Render(fmt.Sprintf("  New %s — Enter Details", m.pendingTmpl.Name)))
	b.WriteString("\n\n")
	for i, name := range m.varNames {
		if i < m.varCursor {
			b.WriteString(m.theme.PreviewKey.Render(fmt.Sprintf("  %-12s", name+":")))
			b.WriteString(m.theme.FileNormal.Render(m.varValues[i]))
		} else if i == m.varCursor {
			b.WriteString(m.theme.PreviewKey.Render(fmt.Sprintf("  %-12s", name+":")))
			b.WriteString(m.varInput.View())
		} else {
			b.WriteString(m.theme.PreviewMeta.Render(fmt.Sprintf("  %-12s", name+":")))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(m.theme.StatusHint.Render("  Enter to confirm, Esc to go back"))
	b.WriteString("\n")
	return b.String()
}

// viewHelp renders the help overlay as a centered bordered box.
func (m Model) viewHelp() string {
	h := m.theme.Accent
	k := m.theme.Subtle
	n := m.theme.FileNormal

	var b strings.Builder
	b.WriteString(h.Render("Keybindings") + "\n\n")

	b.WriteString(h.Render("Navigation") + "\n")
	b.WriteString(k.Render("  j / k       ") + n.Render("move down / up") + "\n")
	b.WriteString(k.Render("  h / l       ") + n.Render("prev / next panel") + "\n")
	b.WriteString(k.Render("  gg / G      ") + n.Render("top / bottom") + "\n\n")

	b.WriteString(h.Render("Modes") + "\n")
	b.WriteString(k.Render("  /           ") + n.Render("search") + "\n")
	b.WriteString(k.Render("  :           ") + n.Render("command") + "\n")
	b.WriteString(k.Render("  Esc         ") + n.Render("cancel / return to normal") + "\n\n")

	b.WriteString(h.Render("Panes") + "\n")
	b.WriteString(k.Render("  1           ") + n.Render("dashboard") + "\n")
	b.WriteString(k.Render("  2           ") + n.Render("journal") + "\n")
	b.WriteString(k.Render("  3           ") + n.Render("knowledge base") + "\n")
	b.WriteString(k.Render("  4           ") + n.Render("tag browser") + "\n")
	b.WriteString(k.Render("  Tab         ") + n.Render("next pane") + "\n")
	b.WriteString(k.Render("  Shift+Tab   ") + n.Render("prev pane") + "\n\n")

	b.WriteString(h.Render("Actions") + "\n")
	b.WriteString(k.Render("  o           ") + n.Render("read document (glamour)") + "\n")
	b.WriteString(k.Render("  Enter       ") + n.Render("open in $EDITOR") + "\n")
	b.WriteString(k.Render("  n           ") + n.Render("new document") + "\n")
	b.WriteString(k.Render("  s           ") + n.Render("scratch pad") + "\n")
	b.WriteString(k.Render("  e           ") + n.Render("export") + "\n")
	b.WriteString(k.Render("  p           ") + n.Render("pin / unpin") + "\n")
	b.WriteString(k.Render("  d           ") + n.Render("delete (with confirmation)") + "\n")
	b.WriteString(k.Render("  R           ") + n.Render("rescan") + "\n")
	b.WriteString(k.Render("  q           ") + n.Render("quit") + "\n\n")

	b.WriteString(h.Render("Commands (:)") + "\n")
	b.WriteString(k.Render("  :todo sweep     ") + n.Render("run TODO sweep") + "\n")
	b.WriteString(k.Render("  :todo archive   ") + n.Render("archive old tasks") + "\n")
	b.WriteString(k.Render("  :q / :quit      ") + n.Render("quit") + "\n\n")

	b.WriteString(m.theme.Muted.Render("Press ? or Esc to close"))

	accentFg := m.theme.Accent.GetForeground()
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentFg).
		Padding(0, 1)
	boxed := boxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxed)
}

// gitMarkerStyled returns a styled git status marker string for the given file path.
func (m Model) gitMarkerStyled(path string) string {
	fs, ok := m.gitFileMap[path]
	if !ok {
		return ""
	}
	switch {
	case fs.IsUntracked():
		return m.theme.GitUntracked.Render(m.icons.GitUntracked)
	case fs.IsStaged():
		return m.theme.GitStaged.Render(m.icons.GitStaged)
	case fs.IsModified():
		return m.theme.GitModified.Render(m.icons.GitModified)
	default:
		return ""
	}
}

// gitMarkerFor returns a plain git status marker string for the given file path.
// Kept for backward compatibility with tests.
func (m Model) gitMarkerFor(path string) string {
	fs, ok := m.gitFileMap[path]
	if !ok {
		return ""
	}
	return gitMarkerForStatus(fs)
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

// styledPanelHeader renders a panel title with lipgloss styling and focus indicator.
func styledPanelHeader(title string, focused bool, width int, th Theme, icons Icons) string {
	_ = icons // reserved for future icon use in headers
	var prefix string
	var style lipgloss.Style
	if focused {
		prefix = "▶ "
		style = th.HeaderFocused
	} else {
		prefix = "─ "
		style = th.HeaderUnfocused
	}
	label := prefix + title + " "
	rest := width - lipgloss.Width(label)
	if rest < 0 {
		rest = 0
	}
	return style.Render(label + strings.Repeat("─", rest))
}

// panelHeader renders a plain panel title line with focus indicator.
// Kept for backward compatibility with tests.
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

// padRight pads s with spaces to visual width n, using ANSI-aware measurement.
func padRight(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

// truncate shortens s to maxLen runes, appending "…" if truncated.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}

// splitLines splits a multiline string into a slice, padding to count lines.
func splitLines(s string, count int) []string {
	lines := strings.Split(s, "\n")
	for len(lines) < count {
		lines = append(lines, "")
	}
	return lines
}
