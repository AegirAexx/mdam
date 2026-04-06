package tui

import (
	"testing"
	"time"

	"github.com/AegirAexx/mdam/internal/document"
	"github.com/AegirAexx/mdam/internal/search"
)

var searchTestDocs = []search.Result{
	{
		Path: "/notes/2026-03-14.md",
		Frontmatter: document.Frontmatter{
			Title:    "Journal 2026-03-14",
			Type:     "journal",
			Tags:     []string{"daily", "project"},
			Modified: time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC),
		},
	},
	{
		Path: "/notes/setup-nginx.md",
		Frontmatter: document.Frontmatter{
			Title:    "Setup Nginx",
			Type:     "kb",
			Tags:     []string{"devops"},
			Modified: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		},
	},
	{
		Path: "/notes/kb-go.md",
		Frontmatter: document.Frontmatter{
			Title:    "Go Notes",
			Type:     "kb_language",
			Tags:     []string{"go", "project"},
			Modified: time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
		},
	},
}

func TestCategorizeResults(t *testing.T) {
	journal, kb, tags := categorizeResults(searchTestDocs, "project")

	if len(journal) != 1 {
		t.Errorf("journal count = %d, want 1", len(journal))
	}
	if len(kb) != 2 {
		t.Errorf("kb count = %d, want 2 (kb + kb_language)", len(kb))
	}
	if len(tags) != 2 {
		t.Errorf("tags count = %d, want 2 (docs with tag matching 'project')", len(tags))
	}
}

func TestCategorizeResultsEmpty(t *testing.T) {
	journal, kb, tags := categorizeResults(nil, "anything")
	if len(journal) != 0 || len(kb) != 0 || len(tags) != 0 {
		t.Errorf("empty results: journal=%d, kb=%d, tags=%d, want all 0",
			len(journal), len(kb), len(tags))
	}
}

func TestCategorizeResultsOverlap(t *testing.T) {
	// A journal doc with a tag matching the query appears in both Journal and Tags.
	journal, _, tags := categorizeResults(searchTestDocs, "daily")
	if len(journal) != 1 {
		t.Errorf("journal count = %d, want 1", len(journal))
	}
	if len(tags) != 1 {
		t.Errorf("tags count = %d, want 1 (daily tag matches 'daily')", len(tags))
	}
}

func TestSearchCategoryDocs(t *testing.T) {
	m := newTestModel()
	m.searchJournalDocs = searchTestDocs[:1]
	m.searchKBDocs = searchTestDocs[1:3]
	m.searchTagDocs = searchTestDocs[2:3]

	tests := []struct {
		cursor int
		want   int
	}{
		{0, 1}, // journal
		{1, 2}, // kb
		{2, 1}, // tags
	}
	for _, tt := range tests {
		m.searchCatCursor = tt.cursor
		docs := m.searchCategoryDocs()
		if len(docs) != tt.want {
			t.Errorf("searchCatCursor=%d: got %d docs, want %d", tt.cursor, len(docs), tt.want)
		}
	}
}

func TestRenderSearchPaneEmptyState(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m.activeView = ViewSearch
	view := stripANSI(m.renderSearchPane())
	if view == "" {
		t.Error("renderSearchPane with no results returned empty string")
	}
}
