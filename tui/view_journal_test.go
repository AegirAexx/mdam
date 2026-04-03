package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/AegirAexx/mdam/internal/document"
	"github.com/AegirAexx/mdam/internal/search"
)

// journalDocs builds a slice of journal search.Result for testing.
func makeJournalDocs(dates []string) []search.Result {
	var out []search.Result
	for _, d := range dates {
		t, _ := time.Parse("2006-01-02", d)
		out = append(out, search.Result{
			Path: "/base/journal/" + d + ".md",
			Frontmatter: document.Frontmatter{
				Type:    "journal",
				Title:   d,
				Created: t,
			},
		})
	}
	return out
}

func TestBuildJournalRowsGrouping(t *testing.T) {
	docs := makeJournalDocs([]string{
		"2026-04-03",
		"2026-04-01",
		"2026-03-15",
	})
	expanded := map[string]bool{"2026-04": true}
	rows := buildJournalRows(docs, expanded)

	// Expect: folder "2026-04" (expanded), file 2026-04-03, file 2026-04-01,
	//         folder "2026-03" (collapsed).
	if len(rows) < 4 {
		t.Fatalf("expected at least 4 rows, got %d", len(rows))
	}
	if !rows[0].isFolder || rows[0].month != "2026-04" {
		t.Errorf("rows[0] should be folder 2026-04, got %+v", rows[0])
	}
	if rows[1].isFolder || rows[1].label != "2026-04-03" {
		t.Errorf("rows[1] should be file 2026-04-03, got %+v", rows[1])
	}
	if rows[2].isFolder || rows[2].label != "2026-04-01" {
		t.Errorf("rows[2] should be file 2026-04-01, got %+v", rows[2])
	}
	if !rows[3].isFolder || rows[3].month != "2026-03" {
		t.Errorf("rows[3] should be folder 2026-03, got %+v", rows[3])
	}
}

func TestBuildJournalRowsCollapsed(t *testing.T) {
	docs := makeJournalDocs([]string{"2026-04-03", "2026-03-15"})
	expanded := map[string]bool{} // nothing expanded
	rows := buildJournalRows(docs, expanded)

	// Only 2 folder rows expected.
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (folders only), got %d", len(rows))
	}
	for _, r := range rows {
		if !r.isFolder {
			t.Errorf("expected folder row, got file row %+v", r)
		}
	}
}

func TestBuildJournalRowsSortOrder(t *testing.T) {
	docs := makeJournalDocs([]string{
		"2026-01-01",
		"2026-04-03",
		"2026-03-15",
	})
	expanded := map[string]bool{}
	rows := buildJournalRows(docs, expanded)

	// Months descending: 2026-04, 2026-03, 2026-01.
	if len(rows) != 3 {
		t.Fatalf("expected 3 folder rows, got %d", len(rows))
	}
	months := []string{rows[0].month, rows[1].month, rows[2].month}
	expected := []string{"2026-04", "2026-03", "2026-01"}
	for i, m := range expected {
		if months[i] != m {
			t.Errorf("rows[%d].month = %q, want %q", i, months[i], m)
		}
	}
}

func TestJournalAutoExpand(t *testing.T) {
	// Use docs from the current month so initJournalView finds something.
	cur := currentMonthKey()
	today := cur + "-15"
	yesterday := cur + "-14"
	docs := makeJournalDocs([]string{today, yesterday})

	m := newTestModel()
	m.docs = docs
	m = initJournalView(m)

	if !m.journalExpanded[cur] {
		t.Errorf("current month %q not expanded after initJournalView", cur)
	}
	// Cursor should be on a file row (not folder), pointing at the most recent.
	rows := buildJournalRows(m.docs, m.journalExpanded)
	if m.journalCursor >= len(rows) {
		t.Fatalf("journalCursor %d out of range (%d rows)", m.journalCursor, len(rows))
	}
	if rows[m.journalCursor].isFolder {
		t.Errorf("journalCursor points to folder row, expected file row")
	}
	if rows[m.journalCursor].label != today {
		t.Errorf("journalCursor label = %q, want most recent %q", rows[m.journalCursor].label, today)
	}
}

func TestJournalToggleExpand(t *testing.T) {
	docs := makeJournalDocs([]string{"2026-04-03", "2026-03-15"})
	m := newTestModel()
	m.docs = docs
	m.activeView = ViewJournal
	m.activePanel = PanelFiles
	m.journalExpanded = make(map[string]bool)
	m.journalCursor = 0 // on "2026-04" folder

	// Press l to expand 2026-04.
	m = sendKey(m, "l")
	if !m.journalExpanded["2026-04"] {
		t.Error("2026-04 should be expanded after l")
	}

	// Press h on same folder row to collapse.
	m = sendKey(m, "h")
	if m.journalExpanded["2026-04"] {
		t.Error("2026-04 should be collapsed after h")
	}
}

func TestJournalEnterOnFolderToggles(t *testing.T) {
	docs := makeJournalDocs([]string{"2026-04-03", "2026-03-15"})
	m := newTestModel()
	m.docs = docs
	m.activeView = ViewJournal
	m.activePanel = PanelFiles
	m.journalExpanded = make(map[string]bool)
	m.journalCursor = 0 // on "2026-04" folder

	// Enter expands.
	m = sendKey(m, "enter")
	if !m.journalExpanded["2026-04"] {
		t.Error("2026-04 should be expanded after enter")
	}

	// Enter again collapses.
	m = sendKey(m, "enter")
	if m.journalExpanded["2026-04"] {
		t.Error("2026-04 should be collapsed after second enter")
	}
}

func TestJournalCursorBounds(t *testing.T) {
	docs := makeJournalDocs([]string{"2026-04-03"})
	m := newTestModel()
	m.docs = docs
	m.activeView = ViewJournal
	m.activePanel = PanelFiles
	m.journalExpanded = make(map[string]bool)
	m.journalCursor = 0 // only 1 folder row

	// j at bottom stays.
	m = sendKey(m, "j")
	if m.journalCursor != 0 {
		t.Errorf("journalCursor = %d after j at bottom, want 0", m.journalCursor)
	}

	// k at top stays.
	m = sendKey(m, "k")
	if m.journalCursor != 0 {
		t.Errorf("journalCursor = %d after k at top, want 0", m.journalCursor)
	}
}

func TestJournalFilePanel(t *testing.T) {
	docs := makeJournalDocs([]string{"2026-04-03", "2026-04-01"})
	m := newTestModel()
	m.width = 80
	m.height = 24
	m.docs = docs
	m.activeView = ViewJournal
	m.activePanel = PanelFiles
	m.journalExpanded = map[string]bool{"2026-04": true}
	m.journalCursor = 1 // first file row

	width := m.width / 3
	rendered := stripANSI(m.renderJournalFilePanel(width, m.contentHeight()))

	if !strings.Contains(rendered, "2026-04") {
		t.Errorf("panel missing month label: %q", rendered)
	}
	if !strings.Contains(rendered, "2026-04-03") {
		t.Errorf("panel missing entry 2026-04-03: %q", rendered)
	}
}
