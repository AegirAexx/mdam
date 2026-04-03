package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/AegirAexx/mdam/internal/search"
)

// dashItem represents a single row in the dashboard left column.
type dashItem struct {
	isHeader bool
	label    string        // section header label or display text
	doc      search.Result // zero value for headers
}

// buildDashItems returns the flat navigable item list for the dashboard left column.
// Sections: Journal (last 5 days), Pinned, Recent (last 20). Docs are deduped by path.
func buildDashItems(m Model) []dashItem {
	var items []dashItem
	seen := map[string]bool{}

	addDoc := func(d search.Result) bool {
		if seen[d.Path] {
			return false
		}
		seen[d.Path] = true
		name := d.Frontmatter.Title
		if name == "" {
			name = d.Path
		}
		items = append(items, dashItem{label: name, doc: d})
		return true
	}

	// Journal: last 5 days.
	items = append(items, dashItem{isHeader: true, label: "Journal"})
	journalDocs := filterByType(m.docs, "journal")
	recent := recentDocs(journalDocs, 5)
	for _, d := range recent {
		addDoc(d)
	}

	// Pinned.
	var pinned []search.Result
	for _, d := range m.docs {
		if m.pinnedPaths[d.Path] {
			pinned = append(pinned, d)
		}
	}
	items = append(items, dashItem{isHeader: true, label: "Pinned"})
	for _, d := range pinned {
		addDoc(d)
	}

	// Recent (last 20, excluding already-shown).
	items = append(items, dashItem{isHeader: true, label: "Recent"})
	allRecent := recentDocs(m.docs, 20)
	for _, d := range allRecent {
		addDoc(d)
	}

	return items
}

// renderDashboard renders the two-column dashboard pane.
func (m Model) renderDashboard() string {
	var b strings.Builder

	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

	contentHeight := m.contentHeight()
	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth - 1

	left := m.renderDashLeft(leftWidth, contentHeight)
	right := m.renderDashRight(rightWidth, contentHeight)

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

// renderDashLeft renders the navigable left column.
func (m Model) renderDashLeft(width, height int) string {
	header := styledPanelHeader("Overview", !m.dashRight, width, m.theme, m.icons)
	var lines []string
	lines = append(lines, header)

	items := buildDashItems(m)
	today := time.Now().Format("2006-01-02")

	visibleRows := height - 2
	start := 0
	if m.dashCursor >= visibleRows {
		start = m.dashCursor - visibleRows + 1
	}

	for i := start; i < len(items) && len(lines) < height-1; i++ {
		it := items[i]
		var line string
		if it.isHeader {
			line = m.theme.DashboardHeader.Render(" " + truncate(it.label, width-2))
		} else {
			var prefix string
			if strings.Contains(it.doc.Path, today) {
				prefix = m.icons.Pinned + " "
			}
			itemText := " " + prefix + truncate(it.label, width-2-len(prefix))
			if i == m.dashCursor && !m.dashRight {
				line = lipgloss.NewStyle().Reverse(true).Width(width).Render(itemText)
			} else {
				line = m.theme.FileNormal.Render(itemText)
			}
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// renderDashRight renders the static TODO right column.
func (m Model) renderDashRight(width, height int) string {
	header := styledPanelHeader("TODOs", m.dashRight, width, m.theme, m.icons)
	var lines []string
	lines = append(lines, header)

	if len(m.todos) == 0 {
		lines = append(lines, m.theme.PreviewMeta.Render("  (no open tasks)"))
		return strings.Join(lines, "\n")
	}

	for _, task := range m.todos {
		if len(lines) >= height-1 {
			break
		}
		lines = append(lines, m.theme.TodoOpen.Render(
			fmt.Sprintf("  · %s", truncate(task.Text, width-4)),
		))
	}
	return strings.Join(lines, "\n")
}

// todayJournal returns the path of today's journal entry if it exists in docs.
func (m Model) todayJournal() string {
	today := time.Now().Format("2006-01-02")
	for _, d := range m.docs {
		if strings.Contains(d.Path, today+".md") {
			return d.Frontmatter.Title
		}
	}
	return ""
}
