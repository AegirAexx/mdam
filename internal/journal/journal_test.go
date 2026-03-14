package journal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEntryPath(t *testing.T) {
	date := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	got := EntryPath("/journals", date)
	want := "/journals/2026-03-14.md"
	if got != want {
		t.Errorf("EntryPath() = %q, want %q", got, want)
	}
}

func TestTodayPath(t *testing.T) {
	path := TodayPath("/journals")
	today := time.Now().Format(DateFormat)
	if !strings.Contains(path, today) {
		t.Errorf("TodayPath() = %q, does not contain today %q", path, today)
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2026-03-14", false},
		{"2026-01-01", false},
		{"not-a-date", true},
		{"2026/03/14", true},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)

	if Exists(dir, date) {
		t.Error("Exists() = true before creation, want false")
	}

	path := EntryPath(dir, date)
	if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !Exists(dir, date) {
		t.Error("Exists() = false after creation, want true")
	}
}

func TestCreate(t *testing.T) {
	dir := t.TempDir()
	journalDir := filepath.Join(dir, "journal")
	tmplDir := filepath.Join(dir, "templates") // no templates dir — uses built-in

	date := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	path, err := Create(journalDir, tmplDir, date)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("Create() file not found at %s", path)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "2026-03-14") {
		t.Errorf("Create() content does not contain date: %q", string(content))
	}
}

func TestCreateIdempotent(t *testing.T) {
	dir := t.TempDir()
	journalDir := filepath.Join(dir, "journal")
	tmplDir := filepath.Join(dir, "templates")

	date := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)

	path1, err := Create(journalDir, tmplDir, date)
	if err != nil {
		t.Fatalf("Create() first call error = %v", err)
	}

	// Write custom content.
	if err := os.WriteFile(path1, []byte("custom content"), 0o644); err != nil {
		t.Fatal(err)
	}

	path2, err := Create(journalDir, tmplDir, date)
	if err != nil {
		t.Fatalf("Create() second call error = %v", err)
	}

	if path1 != path2 {
		t.Errorf("Create() paths differ: %q vs %q", path1, path2)
	}

	// File should not have been overwritten.
	content, _ := os.ReadFile(path2)
	if string(content) != "custom content" {
		t.Errorf("Create() overwrote existing file")
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()

	dates := []string{"2026-03-12", "2026-03-14", "2026-03-13"}
	for _, d := range dates {
		if err := os.WriteFile(filepath.Join(dir, d+".md"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Non-journal file should be ignored.
	if err := os.WriteFile(filepath.Join(dir, "notes.md"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	paths, err := List(dir)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(paths) != 3 {
		t.Errorf("List() returned %d entries, want 3", len(paths))
	}
	// Should be newest-first.
	if !strings.Contains(paths[0], "2026-03-14") {
		t.Errorf("List() first entry = %q, want 2026-03-14", paths[0])
	}
}

func TestListEmptyDir(t *testing.T) {
	paths, err := List("/does/not/exist")
	if err != nil {
		t.Fatalf("List() on missing dir error = %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("List() = %v, want empty", paths)
	}
}

func TestListByMonth(t *testing.T) {
	dir := t.TempDir()
	files := []string{"2026-03-01", "2026-03-14", "2026-04-01", "2026-02-28"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f+".md"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	paths, err := ListByMonth(dir, "2026-03")
	if err != nil {
		t.Fatalf("ListByMonth() error = %v", err)
	}
	if len(paths) != 2 {
		t.Errorf("ListByMonth() returned %d entries, want 2", len(paths))
	}
}

func TestScaffoldFrontmatter(t *testing.T) {
	date := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	fm := ScaffoldFrontmatter(date)
	if fm.Type != "journal" {
		t.Errorf("Type = %q, want journal", fm.Type)
	}
	if !strings.Contains(fm.Title, "2026-03-14") {
		t.Errorf("Title = %q, does not contain date", fm.Title)
	}
	if fm.Tags == nil {
		t.Error("Tags is nil, want empty slice")
	}
}
