package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/AegirAexx/mdam/internal/search"
)

// kbRow represents a single rendered line in the KB tree: either a subtype
// folder header or a file entry.
type kbRow struct {
	isFolder bool
	subtype  string // normalised subtype key (e.g. "Summary", "KB")
	icon     string // "▶" or "▼" (folder rows only)
	label    string // display label: subtype name for folders, doc title for files
	path     string // absolute path (file rows only)
	title    string // doc title (file rows only)
	count    int    // number of files in this subtype folder (folder rows only)
}

// kbSubtype derives the display subtype label from a KB document's type field.
//
//	"kb"            → "KB"
//	"kb_summary"    → "Summary"
//	"kb_ancient-history" → "Ancient History"
func kbSubtype(docType string) string {
	lower := strings.ToLower(strings.TrimSpace(docType))
	suffix := strings.TrimPrefix(lower, "kb")
	suffix = strings.TrimPrefix(suffix, "_")
	if suffix == "" {
		return "KB"
	}
	// Replace hyphens and underscores with spaces, then title-case each word.
	suffix = strings.ReplaceAll(suffix, "-", " ")
	suffix = strings.ReplaceAll(suffix, "_", " ")
	words := strings.Fields(suffix)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// filterKBDocs returns all documents whose type begins with "kb" (case-insensitive).
func filterKBDocs(docs []search.Result) []search.Result {
	var out []search.Result
	for _, d := range docs {
		if strings.HasPrefix(strings.ToLower(d.Frontmatter.Type), "kb") {
			out = append(out, d)
		}
	}
	return out
}

// buildKBRows constructs the flat visible row list from KB docs and expanded state.
// Subtype folders are sorted alphabetically; file rows within each open folder
// are sorted alphabetically by title.
func buildKBRows(docs []search.Result, expanded map[string]bool) []kbRow {
	type group struct {
		subtype string
		files   []kbRow
	}
	bySubtype := map[string]*group{}

	for _, d := range filterKBDocs(docs) {
		sub := kbSubtype(d.Frontmatter.Type)
		if _, ok := bySubtype[sub]; !ok {
			bySubtype[sub] = &group{subtype: sub}
		}
		bySubtype[sub].files = append(bySubtype[sub].files, kbRow{
			isFolder: false,
			subtype:  sub,
			label:    d.Frontmatter.Title,
			path:     d.Path,
			title:    d.Frontmatter.Title,
		})
	}

	// Sort subtypes alphabetically.
	keys := make([]string, 0, len(bySubtype))
	for k := range bySubtype {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var rows []kbRow
	for _, k := range keys {
		g := bySubtype[k]
		// Sort files alphabetically by title.
		sort.Slice(g.files, func(i, j int) bool {
			return g.files[i].title < g.files[j].title
		})
		isExpanded := expanded[k]
		icon := "▶"
		if isExpanded {
			icon = "▼"
		}
		rows = append(rows, kbRow{
			isFolder: true,
			subtype:  k,
			icon:     icon,
			label:    k, // bare subtype label (§9.1 TUI-UX)
			count:    len(g.files),
		})
		if isExpanded {
			rows = append(rows, g.files...)
		}
	}
	return rows
}

// renderKBView renders the full KB pane (tab bar prepended by View()).
func (m Model) renderKBView() string {
	var b strings.Builder

	contentHeight := m.contentHeight()

	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 1

	left := m.renderKBFilePanel(leftWidth, contentHeight)
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

// renderKBFilePanel renders the KB subtype tree into the left panel.
func (m Model) renderKBFilePanel(width, height int) string {
	focused := m.activePanel == PanelFiles
	header := styledPanelHeader("KB", focused, width, m.theme, m.icons)
	var lines []string
	lines = append(lines, header)

	rows := buildKBRows(m.docs, m.kbExpanded)
	if len(rows) == 0 {
		lines = append(lines, lipgloss.NewStyle().PaddingTop(1).PaddingLeft(1).Render(
			m.theme.Muted.Render("No knowledge base documents."),
		))
		return strings.Join(lines, "\n")
	}

	visibleRows := height - 2
	start := 0
	if m.kbCursor >= visibleRows {
		start = m.kbCursor - visibleRows + 1
	}

	for i := start; i < len(rows) && len(lines) < height-1; i++ {
		row := rows[i]
		var line string
		if row.isFolder {
			// Render: icon + label in Primary, [count] in Subtle outside reverse block.
			countStr := m.theme.Subtle.Render(fmt.Sprintf(" [%d]", row.count))
			countWidth := lipgloss.Width(countStr)
			nameText := " " + row.icon + " " + truncate(row.label, width-4-countWidth)
			if i == m.kbCursor && focused {
				namePart := lipgloss.NewStyle().Reverse(true).Width(width - countWidth).Render(nameText)
				line = namePart + countStr
			} else {
				line = m.theme.FileNormal.Render(nameText) + countStr
			}
		} else {
			pinMarker := ""
			if m.pinnedPaths[row.path] {
				pinMarker = " [*]"
			}
			pinW := lipgloss.Width(pinMarker)
			itemText := "  " + truncate(row.label, width-3-pinW) + pinMarker
			if i == m.kbCursor && focused {
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

// kbSelectedPath returns the absolute path of the KB file row under the cursor,
// or "" if the cursor is on a folder row or there are no rows.
func (m Model) kbSelectedPath() string {
	rows := buildKBRows(m.docs, m.kbExpanded)
	if m.kbCursor < 0 || m.kbCursor >= len(rows) {
		return ""
	}
	r := rows[m.kbCursor]
	if r.isFolder {
		return ""
	}
	return r.path
}
