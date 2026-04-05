package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/AegirAexx/mdam/internal/search"
	"github.com/AegirAexx/mdam/internal/todo"
)

// dashItem represents a single row in the dashboard left column.
type dashItem struct {
	isHeader      bool
	isBlank       bool   // non-navigable separator line between sections
	isPlaceholder bool   // non-navigable muted empty-state row
	label         string // section header label, display text, or placeholder text
	doc           search.Result // zero value for headers/blanks/placeholders
}

// buildDashItems returns the flat navigable item list for the dashboard left column.
// Sections: Journal (last 5 days), Pinned, Recent (last 20). Docs are deduped by path.
// Blank rows separate sections; placeholder rows appear when a section is empty.
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
	items = append(items, dashItem{isBlank: true})
	journalDocs := filterByType(m.docs, "journal")
	recent := recentDocs(journalDocs, 5)
	if len(recent) == 0 {
		items = append(items, dashItem{isPlaceholder: true, label: "No recent journal entries."})
	} else {
		for _, d := range recent {
			addDoc(d)
		}
	}
	items = append(items, dashItem{isBlank: true})

	// Pinned.
	items = append(items, dashItem{isHeader: true, label: "Pinned [*]"})
	items = append(items, dashItem{isBlank: true})
	var pinned []search.Result
	for _, d := range m.docs {
		if m.pinnedPaths[d.Path] {
			pinned = append(pinned, d)
		}
	}
	if len(pinned) == 0 {
		items = append(items, dashItem{isPlaceholder: true, label: "No pinned documents."})
	} else {
		for _, d := range pinned {
			addDoc(d)
		}
	}
	items = append(items, dashItem{isBlank: true})

	// Recent (last 20, excluding already-shown).
	items = append(items, dashItem{isHeader: true, label: "Recent"})
	items = append(items, dashItem{isBlank: true})
	allRecent := recentDocs(m.docs, 20)
	recentAdded := 0
	for _, d := range allRecent {
		if addDoc(d) {
			recentAdded++
		}
	}
	if recentAdded == 0 {
		items = append(items, dashItem{isPlaceholder: true, label: "No recent documents."})
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

	visibleRows := height - 2
	start := 0
	if m.dashCursor >= visibleRows {
		start = m.dashCursor - visibleRows + 1
	}

	for i := start; i < len(items) && len(lines) < height-1; i++ {
		it := items[i]
		var line string
		switch {
		case it.isBlank:
			line = ""
		case it.isHeader:
			line = m.theme.Accent.Render(" " + truncate(it.label, width-2))
		case it.isPlaceholder:
			line = lipgloss.NewStyle().PaddingLeft(1).Render(
				m.theme.Muted.Render(truncate(it.label, width-2)),
			)
		default:
			itemText := " " + truncate(it.label, width-2)
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

// renderDashRight renders the static TODO right column with priority grouping.
func (m Model) renderDashRight(width, height int) string {
	header := styledPanelHeader("TODOs", m.dashRight, width, m.theme, m.icons)
	var lines []string
	lines = append(lines, header)

	if len(m.todos) == 0 {
		lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
			m.theme.Muted.Render("No open tasks."),
		))
		return strings.Join(lines, "\n")
	}

	type priorityGroup struct {
		label string
		prio  string
	}
	groups := []priorityGroup{
		{"!high", "high"},
		{"!medium", "medium"},
		{"!low", "low"},
		{"", ""},
	}

	for _, g := range groups {
		var tasks []todo.Task
		for _, t := range m.todos {
			if g.prio == "" {
				// unprioritised: tasks with no priority set
				if t.Priority == "" || strings.EqualFold(t.Priority, "none") {
					tasks = append(tasks, t)
				}
			} else if strings.EqualFold(t.Priority, g.prio) {
				tasks = append(tasks, t)
			}
		}
		if len(tasks) == 0 {
			continue
		}
		if len(lines) >= height-1 {
			break
		}
		if g.label != "" {
			lines = append(lines, m.theme.Subtle.Render(" "+g.label))
		}
		for _, task := range tasks {
			if len(lines) >= height-1 {
				break
			}
			lines = append(lines, m.theme.FileNormal.Render(
				fmt.Sprintf("  %s", truncate(task.Text, width-3)),
			))
		}
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
