package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/AegirAexx/mdam/internal/search"
)

// renderDashboard renders the "today's context" dashboard view.
// It shows: today's journal entry link, open TODO count + top tasks,
// recently modified docs, and pinned docs.
func (m Model) renderDashboard() string {
	var b strings.Builder

	contentHeight := m.height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	var lines []string

	// Header
	today := time.Now().Format("2006-01-02")
	lines = append(lines, m.theme.DashboardHeader.Render("  Today — "+today))
	lines = append(lines, "")

	// Journal entry for today
	lines = append(lines, m.theme.DashboardHeader.Render("  Journal"))
	journalEntry := m.todayJournal()
	if journalEntry != "" {
		lines = append(lines, m.theme.DashboardItem.Render("    "+truncate(journalEntry, m.width-6)))
	} else {
		lines = append(lines, m.theme.PreviewMeta.Render("    (no entry for today)"))
	}
	lines = append(lines, "")

	// TODO summary
	lines = append(lines, m.theme.DashboardHeader.Render(fmt.Sprintf("  TODOs (%d open)", len(m.todos))))
	shown := 5
	if len(m.todos) < shown {
		shown = len(m.todos)
	}
	for _, task := range m.todos[:shown] {
		lines = append(lines, m.theme.TodoOpen.Render("    · "+truncate(task.Text, m.width-6)))
	}
	if len(m.todos) == 0 {
		lines = append(lines, m.theme.PreviewMeta.Render("    (no open tasks)"))
	}
	lines = append(lines, "")

	// Pinned documents
	var pinned []search.Result
	for _, d := range m.docs {
		if m.pinnedPaths[d.Path] {
			pinned = append(pinned, d)
		}
	}
	lines = append(lines, m.theme.DashboardHeader.Render(fmt.Sprintf("  Pinned (%d)", len(pinned))))
	if len(pinned) == 0 {
		lines = append(lines, m.theme.PreviewMeta.Render("    (pin docs with p)"))
	}
	for _, d := range pinned {
		if len(lines) >= contentHeight-5 {
			break
		}
		name := truncate(d.Frontmatter.Title, m.width-6)
		lines = append(lines, m.theme.FilePinned.Render("    "+m.icons.Pinned+" "+name))
	}
	lines = append(lines, "")

	// Recently modified (up to 5)
	recent := recentDocs(m.docs, 5)
	lines = append(lines, m.theme.DashboardHeader.Render("  Recent"))
	for _, d := range recent {
		if len(lines) >= contentHeight-1 {
			break
		}
		ts := d.Frontmatter.Modified.Format("01-02")
		name := truncate(d.Frontmatter.Title, m.width-12)
		lines = append(lines, m.theme.DashboardItem.Render("    "+ts+"  "+name))
	}
	if len(recent) == 0 {
		lines = append(lines, m.theme.PreviewMeta.Render("    (no documents)"))
	}

	// Pad to content height
	for len(lines) < contentHeight {
		lines = append(lines, "")
	}

	for _, l := range lines[:contentHeight] {
		b.WriteString(l + "\n")
	}

	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())
	return b.String()
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
