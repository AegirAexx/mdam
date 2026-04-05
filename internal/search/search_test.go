package search

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// makeDoc writes a test document to dir and returns its path.
func makeDoc(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const docTemplate = `---
title: %s
tags: [%s]
created: 2026-03-14T10:00:00Z
modified: 2026-03-14T10:00:00Z
type: %s
---

%s
`

func TestSearch(t *testing.T) {
	dir := t.TempDir()

	import_fmt := func(title, tags, docType, body string) string {
		return "---\ntitle: " + title + "\ntags:\n  - " + tags + "\ncreated: 2026-03-14T10:00:00Z\nmodified: 2026-03-14T10:00:00Z\ntype: " + docType + "\n---\n\n" + body + "\n"
	}

	makeDoc(t, dir, "setup-nginx.md", import_fmt("Setup Nginx", "devops", "kb", "How to configure nginx."))
	makeDoc(t, dir, "deploy-runbook.md", import_fmt("Deploy Runbook", "devops", "kb", "Deployment steps."))
	makeDoc(t, dir, "2026-03-14.md", import_fmt("Journal 2026-03-14", "journal", "journal", "Today's notes."))

	// Search for "nginx" should find setup-nginx.md first.
	results, err := Search(dir, "nginx", Filters{})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search() returned no results")
	}
	if results[0].Frontmatter.Title != "Setup Nginx" {
		t.Errorf("Search() top result = %q, want Setup Nginx", results[0].Frontmatter.Title)
	}
}

func TestSearchTagFilter(t *testing.T) {
	dir := t.TempDir()

	doc := func(title, tag, typ string) string {
		return "---\ntitle: " + title + "\ntags:\n  - " + tag + "\ncreated: 2026-03-14T10:00:00Z\nmodified: 2026-03-14T10:00:00Z\ntype: " + typ + "\n---\n"
	}
	makeDoc(t, dir, "a.md", doc("A Doc", "devops", "kb"))
	makeDoc(t, dir, "b.md", doc("B Doc", "personal", "kb"))

	results, err := Search(dir, "", Filters{Tag: "devops"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 || results[0].Frontmatter.Title != "A Doc" {
		t.Errorf("Search() with tag filter = %v, want [A Doc]", results)
	}
}

func TestSearchTypeFilter(t *testing.T) {
	dir := t.TempDir()
	doc := func(title, typ string) string {
		return "---\ntitle: " + title + "\ntags: []\ncreated: 2026-03-14T10:00:00Z\nmodified: 2026-03-14T10:00:00Z\ntype: " + typ + "\n---\n"
	}
	makeDoc(t, dir, "journal.md", doc("Journal", "journal"))
	makeDoc(t, dir, "kb.md", doc("KB", "kb"))

	results, err := Search(dir, "", Filters{Type: "kb"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 || results[0].Frontmatter.Type != "kb" {
		t.Errorf("Search() type filter results = %v", results)
	}
}

func TestSearchModifiedAfterFilter(t *testing.T) {
	dir := t.TempDir()
	makeDoc(t, dir, "old.md", "---\ntitle: Old\ntags: []\ncreated: 2026-01-01T00:00:00Z\nmodified: 2026-01-01T00:00:00Z\ntype: kb\n---\n")
	makeDoc(t, dir, "new.md", "---\ntitle: New\ntags: []\ncreated: 2026-03-14T00:00:00Z\nmodified: 2026-03-14T00:00:00Z\ntype: kb\n---\n")

	cutoff := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	results, err := Search(dir, "", Filters{ModifiedAfter: cutoff})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 || results[0].Frontmatter.Title != "New" {
		t.Errorf("Search() modified-after filter = %v, want [New]", results)
	}
}

func TestSearchEmptyDir(t *testing.T) {
	dir := t.TempDir()
	results, err := Search(dir, "anything", Filters{})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() on empty dir = %v, want empty", results)
	}
}

func TestSearchSkipsHiddenDirs(t *testing.T) {
	dir := t.TempDir()
	hiddenDir := filepath.Join(dir, ".hidden")
	if err := os.MkdirAll(hiddenDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Doc in hidden dir should not appear.
	makeDoc(t, hiddenDir, "secret.md", "---\ntitle: Secret\ntags: []\ncreated: 2026-03-14T10:00:00Z\nmodified: 2026-03-14T10:00:00Z\ntype: kb\n---\n")

	results, err := Search(dir, "Secret", Filters{})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() should skip hidden dirs, got %v", results)
	}
}

func TestFuzzyContains(t *testing.T) {
	tests := []struct {
		s    string
		sub  string
		want bool
	}{
		{"nginx setup guide", "nginx", true},
		{"nginx setup guide", "ngnx", true},   // fuzzy: n-g-n-x in order
		{"nginx setup guide", "xyz", false},
		{"", "a", false},
		{"abc", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.s+"/"+tt.sub, func(t *testing.T) {
			got := fuzzyContains(tt.s, tt.sub)
			if got != tt.want {
				t.Errorf("fuzzyContains(%q, %q) = %v, want %v", tt.s, tt.sub, got, tt.want)
			}
		})
	}
}

func TestListAll(t *testing.T) {
	dir := t.TempDir()
	doc := func(title string) string {
		return "---\ntitle: " + title + "\ntags: []\ncreated: 2026-03-14T10:00:00Z\nmodified: 2026-03-14T10:00:00Z\ntype: kb\n---\n"
	}
	makeDoc(t, dir, "a.md", doc("A"))
	makeDoc(t, dir, "b.md", doc("B"))
	makeDoc(t, dir, "c.md", doc("C"))

	results, skipped, err := ListAll(dir)
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if skipped != 0 {
		t.Errorf("ListAll() skipped = %d, want 0", skipped)
	}
	if len(results) != 3 {
		t.Errorf("ListAll() = %d results, want 3", len(results))
	}
}

func TestSearchWithBody(t *testing.T) {
	dir := t.TempDir()
	makeDoc(t, dir, "note.md", "---\ntitle: Note\ntags: []\ncreated: 2026-03-14T10:00:00Z\nmodified: 2026-03-14T10:00:00Z\ntype: kb\n---\n\nThe quick brown fox jumps.\n")

	results, err := SearchWithBody(dir, "quick brown", Filters{})
	if err != nil {
		t.Fatalf("SearchWithBody() error = %v", err)
	}
	if len(results) == 0 {
		t.Fatal("SearchWithBody() returned no results")
	}
	if results[0].Snippet == "" {
		t.Error("SearchWithBody() should set Snippet for body match")
	}
}
