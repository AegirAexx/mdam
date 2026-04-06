package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/AegirAexx/mdam/internal/search"
)

// journalRow represents a single rendered line in the journal tree:
// either a month-folder header or a file entry.
type journalRow struct {
	isFolder bool
	month    string    // "2026-04" — used as key in expanded map
	icon     string    // "▶" or "▼" (folder rows only)
	label    string    // display label: YYYY-MM for folders, YYYY-MM-DD for files
	path     string    // absolute path (file rows only)
	date     time.Time
	count    int // number of files in this month (folder rows only)
}

// buildJournalRows constructs the flat visible list of rows from the given docs
// and expanded-month state. Month folders appear in descending order; files
// within each open month appear in descending date order.
func buildJournalRows(docs []search.Result, expanded map[string]bool) []journalRow {
	// Group by month key.
	type monthGroup struct {
		key   string
		files []journalRow
	}
	byMonth := map[string]*monthGroup{}

	for _, d := range docs {
		if !strings.EqualFold(d.Frontmatter.Type, "journal") {
			continue
		}
		base := filepath.Base(d.Path)
		// Parse the date from the filename (YYYY-MM-DD.md).
		name := strings.TrimSuffix(base, filepath.Ext(base))
		t, err := time.Parse("2006-01-02", name)
		if err != nil {
			// Try the doc's created date as fallback.
			t = d.Frontmatter.Created
		}
		key := t.Format("2006-01")
		if _, ok := byMonth[key]; !ok {
			byMonth[key] = &monthGroup{key: key}
		}
		byMonth[key].files = append(byMonth[key].files, journalRow{
			isFolder: false,
			month:    key,
			label:    name,
			path:     d.Path,
			date:     t,
		})
	}

	// Sort months descending.
	keys := make([]string, 0, len(byMonth))
	for k := range byMonth {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	var rows []journalRow
	for _, k := range keys {
		g := byMonth[k]
		// Sort files descending by date within each month.
		sort.Slice(g.files, func(i, j int) bool {
			return g.files[i].date.After(g.files[j].date)
		})
		isExpanded := expanded[k]
		icon := "▶"
		if isExpanded {
			icon = "▼"
		}
		rows = append(rows, journalRow{
			isFolder: true,
			month:    k,
			icon:     icon,
			label:    k, // bare YYYY-MM key (§9.1 TUI-UX)
			count:    len(g.files),
		})
		if isExpanded {
			rows = append(rows, g.files...)
		}
	}
	return rows
}

// currentMonthKey returns the current month as "YYYY-MM".
func currentMonthKey() string {
	return time.Now().Format("2006-01")
}

// initJournalView sets up journalExpanded and journalCursor for entering the
// Journal pane. The current month is auto-expanded; the cursor lands on the
// most recent file entry in that month.
func initJournalView(m Model) Model {
	if m.journalExpanded == nil {
		m.journalExpanded = make(map[string]bool)
	}
	// Collapse all, then expand current month.
	for k := range m.journalExpanded {
		delete(m.journalExpanded, k)
	}
	cur := currentMonthKey()
	m.journalExpanded[cur] = true

	rows := buildJournalRows(m.docs, m.journalExpanded)
	// Position cursor on the first file row (most recent entry, since sorted desc).
	m.journalCursor = 0
	for i, r := range rows {
		if !r.isFolder {
			m.journalCursor = i
			return m
		}
	}
	// Current month has no entries. Expand the most recent past-month folder
	// that has files so the cursor lands on a document rather than a folder.
	for _, r := range rows {
		if r.isFolder && !m.journalExpanded[r.month] {
			m.journalExpanded[r.month] = true
			rows = buildJournalRows(m.docs, m.journalExpanded)
			for i, row := range rows {
				if !row.isFolder {
					m.journalCursor = i
					break
				}
			}
			break
		}
	}
	return m
}

// renderJournalView renders the full journal pane (tab bar already handled by View()).
func (m Model) renderJournalView() string {
	var b strings.Builder

	contentHeight := m.contentHeight()

	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 1

	left := m.renderJournalFilePanel(leftWidth, contentHeight)
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
		b.WriteString(m.theme.Divider.Render("│"))
		b.WriteString(r)
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())
	return b.String()
}

// renderJournalFilePanel renders the tree of journal months and entries.
func (m Model) renderJournalFilePanel(width, height int) string {
	focused := m.activePanel == PanelFiles
	header := styledPanelHeader("Journal", focused, width, m.theme, m.icons)
	var lines []string
	lines = append(lines, header)

	if m.journalExpanded == nil {
		lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
			m.theme.Muted.Render("No journal entries."),
		))
		return strings.Join(lines, "\n")
	}

	rows := buildJournalRows(m.docs, m.journalExpanded)
	if len(rows) == 0 {
		lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
			m.theme.Muted.Render("No journal entries."),
		))
		return strings.Join(lines, "\n")
	}

	visibleRows := height - 2
	start := 0
	if m.journalCursor >= visibleRows {
		start = m.journalCursor - visibleRows + 1
	}

	for i := start; i < len(rows) && len(lines) < height-1; i++ {
		row := rows[i]
		var line string
		if row.isFolder {
			// Render: icon + label in Primary, [count] in Subtle outside reverse block.
			countStr := m.theme.Subtle.Render(fmt.Sprintf(" [%d]", row.count))
			countWidth := lipgloss.Width(countStr)
			nameText := " " + row.icon + " " + truncate(row.label, width-4-countWidth)
			if i == m.journalCursor && focused {
				namePart := lipgloss.NewStyle().Reverse(true).Width(width - countWidth).Render(nameText)
				line = namePart + countStr
			} else {
				line = m.theme.FileNormal.Render(nameText) + countStr
			}
		} else {
			pinMarker := ""
			if m.pinnedPaths[row.path] {
				pinMarker = " " + m.icons.Pinned
			}
			gitMarker := ""
			if g := m.gitMarkerStyled(row.path); g != "" {
				gitMarker = " " + g
			}
			markerW := lipgloss.Width(pinMarker) + lipgloss.Width(gitMarker)
			itemText := "  " + truncate(row.label, width-3-markerW) + pinMarker + gitMarker
			if i == m.journalCursor && focused {
				line = lipgloss.NewStyle().Reverse(true).Width(width).Render(itemText)
			} else if m.pinnedPaths[row.path] {
				line = m.theme.FilePinned.Render(itemText)
			} else {
				line = m.theme.FileNormal.Render(itemText)
			}
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// journalSelectedPath returns the absolute path of the journal file row under
// the cursor, or "" if the cursor is on a folder row or there are no rows.
func (m Model) journalSelectedPath() string {
	if m.journalExpanded == nil {
		return ""
	}
	rows := buildJournalRows(m.docs, m.journalExpanded)
	if m.journalCursor < 0 || m.journalCursor >= len(rows) {
		return ""
	}
	r := rows[m.journalCursor]
	if r.isFolder {
		return ""
	}
	return r.path
}
