package tui

import (
	"strings"
	"testing"

	"github.com/AegirAexx/mdam/internal/document"
	"github.com/AegirAexx/mdam/internal/search"
)

func makeKBDoc(docType, title, path string) search.Result {
	return search.Result{
		Path: path,
		Frontmatter: document.Frontmatter{
			Type:  docType,
			Title: title,
		},
	}
}

func TestKBSubtype(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"kb", "KB"},
		{"KB", "KB"},
		{"kb_summary", "Summary"},
		{"KB_SUMMARY", "Summary"},
		{"kb_ancient-history", "Ancient History"},
		{"kb_domain", "Domain"},
		{"KB_DOMAIN", "Domain"},
		{"kb_foo_bar", "Foo Bar"},
	}
	for _, tt := range tests {
		got := kbSubtype(tt.input)
		if got != tt.want {
			t.Errorf("kbSubtype(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFilterKBDocsPrefixMatch(t *testing.T) {
	docs := []search.Result{
		makeKBDoc("kb", "Base", "/kb/base.md"),
		makeKBDoc("kb_summary", "Summary", "/kb/summary.md"),
		makeKBDoc("KB_DOMAIN", "Domain", "/kb/domain.md"),
		makeKBDoc("journal", "Journal", "/journal/2026-04-03.md"),
		makeKBDoc("scratch", "Scratch", "/scratch/scratch.md"),
	}
	got := filterKBDocs(docs)
	if len(got) != 3 {
		t.Errorf("filterKBDocs: got %d docs, want 3", len(got))
	}
	for _, d := range got {
		if !strings.HasPrefix(strings.ToLower(d.Frontmatter.Type), "kb") {
			t.Errorf("filterKBDocs included non-KB type %q", d.Frontmatter.Type)
		}
	}
}

func TestBuildKBRowsGrouping(t *testing.T) {
	docs := []search.Result{
		makeKBDoc("kb_summary", "Alpha", "/kb/alpha.md"),
		makeKBDoc("kb_summary", "Beta", "/kb/beta.md"),
		makeKBDoc("kb", "Base Entry", "/kb/base.md"),
	}
	expanded := map[string]bool{"Summary": true}
	rows := buildKBRows(docs, expanded)

	// Expect alphabetical: "KB" folder, then "Summary" folder + 2 files.
	if len(rows) < 4 {
		t.Fatalf("expected at least 4 rows, got %d: %+v", len(rows), rows)
	}
	if !rows[0].isFolder || rows[0].subtype != "KB" {
		t.Errorf("rows[0] should be KB folder, got %+v", rows[0])
	}
	if !rows[1].isFolder || rows[1].subtype != "Summary" {
		t.Errorf("rows[1] should be Summary folder, got %+v", rows[1])
	}
	if rows[2].isFolder || rows[2].title != "Alpha" {
		t.Errorf("rows[2] should be file Alpha, got %+v", rows[2])
	}
	if rows[3].isFolder || rows[3].title != "Beta" {
		t.Errorf("rows[3] should be file Beta, got %+v", rows[3])
	}
}

func TestBuildKBRowsCollapsed(t *testing.T) {
	docs := []search.Result{
		makeKBDoc("kb_summary", "Alpha", "/kb/alpha.md"),
		makeKBDoc("kb", "Base", "/kb/base.md"),
	}
	expanded := map[string]bool{} // nothing expanded
	rows := buildKBRows(docs, expanded)

	if len(rows) != 2 {
		t.Fatalf("expected 2 folder rows (collapsed), got %d", len(rows))
	}
	for _, r := range rows {
		if !r.isFolder {
			t.Errorf("expected folder row, got file row: %+v", r)
		}
	}
}

func TestKBFilePanel(t *testing.T) {
	docs := []search.Result{
		makeKBDoc("kb_summary", "My Summary", "/kb/summary.md"),
	}
	m := newTestModel()
	m.width = 80
	m.height = 24
	m.docs = docs
	m.activeView = ViewKB
	m.activePanel = PanelFiles
	m.kbExpanded = map[string]bool{"Summary": true}
	m.kbCursor = 1 // file row

	width := m.width / 3
	rendered := stripANSI(m.renderKBFilePanel(width, m.contentHeight()))

	if !strings.Contains(rendered, "Summary") {
		t.Errorf("panel missing Summary folder: %q", rendered)
	}
	if !strings.Contains(rendered, "My Summary") {
		t.Errorf("panel missing file title: %q", rendered)
	}
}

// TestKBExpandRelocatesCursor verifies that expanding a folder below a large
// expanded folder relocates kbCursor to the newly-expanded folder, preventing
// an index-out-of-bounds panic on the next collapse.
func TestKBExpandRelocatesCursor(t *testing.T) {
	// Build a doc set where "Summary" has many children and "Vps" has few.
	var docs []search.Result
	for i := 0; i < 14; i++ {
		docs = append(docs, makeKBDoc("kb_summary", string(rune('A'+i)), "/kb/s"+string(rune('a'+i))+".md"))
	}
	docs = append(docs, makeKBDoc("kb_vps", "Server1", "/kb/v1.md"))
	docs = append(docs, makeKBDoc("kb_vps", "Server2", "/kb/v2.md"))

	m := newTestModel()
	m.width = 80
	m.height = 24
	m.docs = docs
	m.activeView = ViewKB
	m.activePanel = PanelFiles
	m.kbExpanded = map[string]bool{"Summary": true}

	// With Summary expanded: row 0 = Summary folder, rows 1-14 = files,
	// row 15 = Vps folder. Place cursor on Vps.
	rows := buildKBRows(m.docs, m.kbExpanded)
	vpsFolderIdx := -1
	for i, r := range rows {
		if r.isFolder && r.subtype == "Vps" {
			vpsFolderIdx = i
			break
		}
	}
	if vpsFolderIdx < 0 {
		t.Fatal("Vps folder not found in rows")
	}
	m.kbCursor = vpsFolderIdx

	// Expand Vps (collapses Summary, rebuilds list, relocates cursor).
	m = sendKey(m, "l")

	newRows := buildKBRows(m.docs, m.kbExpanded)
	if m.kbCursor >= len(newRows) {
		t.Fatalf("kbCursor %d out of range after expand (len %d)", m.kbCursor, len(newRows))
	}
	if !newRows[m.kbCursor].isFolder || newRows[m.kbCursor].subtype != "Vps" {
		t.Errorf("expected cursor on Vps folder, got %+v", newRows[m.kbCursor])
	}

	// Collapse via h — must not panic.
	m = sendKey(m, "h")

	finalRows := buildKBRows(m.docs, m.kbExpanded)
	if m.kbCursor >= len(finalRows) {
		t.Fatalf("kbCursor %d out of range after collapse (len %d)", m.kbCursor, len(finalRows))
	}
}

// TestKBCollapseClampsStaleCursor verifies that pressing h with a stale
// (out-of-bounds) kbCursor clamps it instead of panicking.
func TestKBCollapseClampsStaleCursor(t *testing.T) {
	docs := []search.Result{
		makeKBDoc("kb_summary", "Alpha", "/kb/alpha.md"),
		makeKBDoc("kb", "Base", "/kb/base.md"),
	}

	m := newTestModel()
	m.width = 80
	m.height = 24
	m.docs = docs
	m.activeView = ViewKB
	m.activePanel = PanelFiles
	m.kbExpanded = map[string]bool{}
	m.kbCursor = 20 // deliberately stale / out of bounds

	// Must not panic.
	m = sendKey(m, "h")

	rows := buildKBRows(m.docs, m.kbExpanded)
	if m.kbCursor < 0 || m.kbCursor >= len(rows) {
		t.Fatalf("kbCursor %d out of range after clamp (len %d)", m.kbCursor, len(rows))
	}
}
