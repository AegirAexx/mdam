package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/AegirAexx/mdam/internal/search"
)

// tagEntry holds a tag name and the number of documents that carry it.
type tagEntry struct {
	Name  string
	Count int
}

// buildTagIndex aggregates all tags from docs, returning entries sorted by
// count descending, then name ascending for equal counts.
func buildTagIndex(docs []search.Result) []tagEntry {
	counts := make(map[string]int)
	for _, d := range docs {
		for _, tag := range d.Frontmatter.Tags {
			counts[tag]++
		}
	}
	entries := make([]tagEntry, 0, len(counts))
	for name, count := range counts {
		entries = append(entries, tagEntry{Name: name, Count: count})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries
}

// renderTagBrowser renders the tag browser view (replaces normal panel layout).
func (m Model) renderTagBrowser() string {
	var b strings.Builder

	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

	contentHeight := m.contentHeight()

	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 1

	// Left: tag list
	left := m.renderTagPanel(leftWidth, contentHeight)
	// Right: preview of docs with selected tag
	right := m.renderTagDocPanel(rightWidth, contentHeight)

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

// renderTagPanel renders the list of tags in the left panel.
func (m Model) renderTagPanel(width, height int) string {
	tagFocused := m.activePanel == PanelFiles
	header := styledPanelHeader("Tags", tagFocused, width, m.theme, m.icons)
	var lines []string
	lines = append(lines, header)

	// Filter input row.
	if m.tagFilterActive {
		lines = append(lines, " "+m.tagFilterInput.View())
	} else if v := m.tagFilterInput.Value(); v != "" {
		lines = append(lines, m.theme.Muted.PaddingLeft(1).Render(
			fmt.Sprintf("filter: %s  (Esc to clear)", v)))
	} else {
		lines = append(lines, m.theme.Muted.PaddingLeft(1).Render("Enter to filter..."))
	}

	entries := m.visibleTagEntries()

	if len(entries) == 0 {
		if m.tagFilterInput.Value() != "" {
			lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
				m.theme.Muted.Render("No matching tags."),
			))
		} else {
			lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
				m.theme.Muted.Render("No tags found."),
			))
		}
		return strings.Join(lines, "\n")
	}

	visibleRows := height - 3 // header + filter input + bottom margin
	start := 0
	if m.tagCursor >= visibleRows {
		start = m.tagCursor - visibleRows + 1
	}

	for i := start; i < len(entries) && len(lines) < height-1; i++ {
		te := entries[i]
		tagName := m.icons.Tag + te.Name
		countStr := fmt.Sprintf(" (%d)", te.Count)
		itemText := " " + truncate(tagName, width-len(countStr)-2) + countStr
		var line string
		if i == m.tagCursor && tagFocused {
			line = lipgloss.NewStyle().Reverse(true).Width(width).Render(itemText)
		} else {
			line = m.theme.FileNormal.Render(itemText)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// renderTagDocPanel renders docs that have the currently selected tag.
func (m Model) renderTagDocPanel(width, height int) string {
	docFocused := m.activePanel == PanelPreview
	header := styledPanelHeader("Documents", docFocused, width, m.theme, m.icons)
	var lines []string
	lines = append(lines, header)

	if m.tagCursor >= len(m.tagEntries) {
		lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
			m.theme.Muted.Render("Select a tag to see documents."),
		))
		return strings.Join(lines, "\n")
	}

	tagged := m.taggedDocs()

	if len(tagged) == 0 {
		lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
			m.theme.Muted.Render("Select a tag to see documents."),
		))
		return strings.Join(lines, "\n")
	}

	for i, d := range tagged {
		if len(lines) >= height-1 {
			break
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
		itemText := " " + truncate(d.Frontmatter.Title, width-2-markerW) + pinMarker + gitMarker
		var line string
		if i == m.tagDocCursor && docFocused {
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
