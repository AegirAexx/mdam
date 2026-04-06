package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderSearchPane renders the search pane (View 5) with two columns.
func (m Model) renderSearchPane() string {
	var b strings.Builder

	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

	contentHeight := m.contentHeight()

	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 1

	left := m.renderSearchLeftPanel(leftWidth, contentHeight)
	right := m.renderSearchDocPanel(rightWidth, contentHeight)

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
		b.WriteString(m.theme.Divider.Render("│"))
		b.WriteString(r)
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())
	return b.String()
}

// renderSearchLeftPanel renders the search input and category list.
func (m Model) renderSearchLeftPanel(width, height int) string {
	focused := m.activePanel == PanelFiles
	header := styledPanelHeader("Search", focused, width, m.theme, m.icons)
	var lines []string
	lines = append(lines, header)

	// Search input row.
	if m.searchInputFocused {
		lines = append(lines, " "+m.searchPaneInput.View())
	} else if m.searchPaneQuery != "" {
		lines = append(lines, m.theme.Muted.PaddingLeft(1).Render(
			fmt.Sprintf("filter: %s  (Esc to clear)", m.searchPaneQuery)))
	} else {
		lines = append(lines, m.theme.Muted.PaddingLeft(1).Render("Enter to search..."))
	}
	lines = append(lines, "") // blank separator

	// Category entries.
	hasResults := len(m.searchJournalDocs) > 0 || len(m.searchKBDocs) > 0 || len(m.searchTagDocs) > 0

	if !hasResults && m.searchPaneQuery == "" {
		lines = append(lines, lipgloss.NewStyle().PaddingLeft(1).Render(
			m.theme.Muted.Render("Type a query and press Enter.")))
		return strings.Join(lines, "\n")
	}
	if !hasResults && m.searchPaneQuery != "" {
		lines = append(lines, lipgloss.NewStyle().PaddingLeft(1).Render(
			m.theme.Muted.Render(fmt.Sprintf("No results for %q.", m.searchPaneQuery))))
		return strings.Join(lines, "\n")
	}

	type catDef struct {
		label string
		count int
	}
	cats := []catDef{
		{"Journal", len(m.searchJournalDocs)},
		{"KB", len(m.searchKBDocs)},
		{"Tags", len(m.searchTagDocs)},
	}

	for i, c := range cats {
		itemText := fmt.Sprintf(" %s (%d)", c.label, c.count)
		var line string
		if i == m.searchCatCursor && focused && !m.searchInputFocused {
			line = lipgloss.NewStyle().Reverse(true).Width(width).Render(itemText)
		} else if c.count == 0 {
			line = m.theme.Muted.Render(itemText)
		} else {
			line = m.theme.FileNormal.Render(itemText)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderSearchDocPanel renders the documents for the selected search category.
func (m Model) renderSearchDocPanel(width, height int) string {
	docFocused := m.activePanel == PanelPreview
	header := styledPanelHeader("Documents", docFocused, width, m.theme, m.icons)
	var lines []string
	lines = append(lines, header)

	docs := m.searchCategoryDocs()

	if len(docs) == 0 {
		placeholder := "Select a category to see documents."
		if m.searchPaneQuery != "" {
			placeholder = "No documents in this category."
		}
		lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
			m.theme.Muted.Render(placeholder)))
		return strings.Join(lines, "\n")
	}

	visibleRows := height - 2
	start := 0
	if m.searchDocCursor >= visibleRows {
		start = m.searchDocCursor - visibleRows + 1
	}

	for i := start; i < len(docs) && len(lines) < height-1; i++ {
		d := docs[i]
		name := d.Frontmatter.Title
		if name == "" {
			name = filepath.Base(d.Path)
		}
		pinMarker := ""
		if m.pinnedPaths[d.Path] {
			pinMarker = " " + m.icons.Pinned
		}
		gitMarker := ""
		if g := m.gitMarkerStyled(d.Path); g != "" {
			gitMarker = " " + g
		}
		markerW := lipgloss.Width(pinMarker) + lipgloss.Width(gitMarker)
		itemText := " " + truncate(name, width-2-markerW) + pinMarker + gitMarker
		var line string
		if i == m.searchDocCursor && docFocused {
			line = lipgloss.NewStyle().Reverse(true).Width(width).Render(itemText)
		} else if m.pinnedPaths[d.Path] {
			line = m.theme.FilePinned.Render(itemText)
		} else {
			line = m.theme.FileNormal.Render(itemText)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
