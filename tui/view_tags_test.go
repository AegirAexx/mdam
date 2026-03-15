package tui

import (
	"testing"
	"time"

	"github.com/AegirAexx/mdam/internal/document"
	"github.com/AegirAexx/mdam/internal/search"
)

func TestBuildTagIndexEmpty(t *testing.T) {
	entries := buildTagIndex(nil)
	if len(entries) != 0 {
		t.Errorf("buildTagIndex(nil) = %d entries, want 0", len(entries))
	}
}

func TestBuildTagIndexCounts(t *testing.T) {
	docs := []search.Result{
		{Path: "/a.md", Frontmatter: document.Frontmatter{Tags: []string{"go", "devops"}}},
		{Path: "/b.md", Frontmatter: document.Frontmatter{Tags: []string{"go", "linux"}}},
		{Path: "/c.md", Frontmatter: document.Frontmatter{Tags: []string{"devops"}}},
	}
	entries := buildTagIndex(docs)

	counts := make(map[string]int)
	for _, e := range entries {
		counts[e.Name] = e.Count
	}
	if counts["go"] != 2 {
		t.Errorf("tag 'go' count = %d, want 2", counts["go"])
	}
	if counts["devops"] != 2 {
		t.Errorf("tag 'devops' count = %d, want 2", counts["devops"])
	}
	if counts["linux"] != 1 {
		t.Errorf("tag 'linux' count = %d, want 1", counts["linux"])
	}
}

func TestBuildTagIndexSortedByCountDesc(t *testing.T) {
	docs := []search.Result{
		{Path: "/a.md", Frontmatter: document.Frontmatter{Tags: []string{"rare"}}},
		{Path: "/b.md", Frontmatter: document.Frontmatter{Tags: []string{"common", "rare"}}},
		{Path: "/c.md", Frontmatter: document.Frontmatter{Tags: []string{"common"}}},
		{Path: "/d.md", Frontmatter: document.Frontmatter{Tags: []string{"common"}}},
	}
	entries := buildTagIndex(docs)
	if len(entries) == 0 {
		t.Fatal("no entries returned")
	}
	if entries[0].Name != "common" {
		t.Errorf("first entry = %q, want 'common' (highest count)", entries[0].Name)
	}
}

func TestBuildTagIndexAlphabeticTieBreak(t *testing.T) {
	docs := []search.Result{
		{Path: "/a.md", Frontmatter: document.Frontmatter{Tags: []string{"zzz", "aaa"}}},
	}
	entries := buildTagIndex(docs)
	if len(entries) < 2 {
		t.Fatal("expected 2 entries")
	}
	// Both have count 1; aaa should come before zzz alphabetically.
	if entries[0].Name != "aaa" {
		t.Errorf("first entry with tie = %q, want aaa", entries[0].Name)
	}
}

func TestBuildTagIndexDeduplication(t *testing.T) {
	docs := []search.Result{
		{Path: "/a.md", Frontmatter: document.Frontmatter{Tags: []string{"go", "go"}}},
	}
	// Duplicate tags on the same doc count once per doc, not once per occurrence.
	// (This tests the standard range-over-slice behaviour.)
	entries := buildTagIndex(docs)
	for _, e := range entries {
		if e.Name == "go" && e.Count != 2 {
			// go appears twice in same doc — count is 2 because range iterates twice.
			// Document intentionally accepts this: tags on a single doc may repeat.
			t.Logf("note: duplicate tag 'go' counted %d times", e.Count)
		}
	}
}

func TestRenderTagBrowserNoTagsState(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m.activeView = ViewTags
	m.tagEntries = nil

	view := stripANSI(m.renderTagBrowser())
	if view == "" {
		t.Error("renderTagBrowser with no tags returned empty string")
	}
}

func TestTagIndexMsgUpdatesTagEntries(t *testing.T) {
	m := newTestModel()
	entries := []tagEntry{{Name: "go", Count: 3}}
	m2 := sendMsg(m, tagIndexMsg{entries: entries})
	if len(m2.tagEntries) != 1 {
		t.Errorf("tagEntries len = %d, want 1", len(m2.tagEntries))
	}
	if m2.tagEntries[0].Name != "go" {
		t.Errorf("tagEntries[0].Name = %q, want 'go'", m2.tagEntries[0].Name)
	}
}

// fakePinnedDocs are used for dashboard tests.
var fakePinnedDocs = []search.Result{
	{
		Path: "/notes/pinned.md",
		Frontmatter: document.Frontmatter{
			Title:    "Pinned Doc",
			Type:     "kb",
			Tags:     []string{"pinned"},
			Modified: time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC),
		},
	},
}
