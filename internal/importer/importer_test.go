package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AegirAexx/mdam/internal/document"
)

func validDoc(title, typ string) string {
	return "---\ntitle: " + title + "\ntags: []\ncreated: 2026-03-14T10:00:00Z\nmodified: 2026-03-14T10:00:00Z\ntype: " + typ + "\n---\n\nBody.\n"
}

func TestImportFileValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "my-note.md")
	if err := os.WriteFile(path, []byte(validDoc("My Note", "kb")), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ImportFile(path, "", Options{})
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if len(result.Errors) != 0 {
		t.Errorf("ImportFile() errors = %v, want none", result.Errors)
	}
}

func TestImportFileInvalidFilenameNoAutoFix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "My Note.md")
	if err := os.WriteFile(path, []byte(validDoc("My Note", "kb")), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ImportFile(path, "", Options{AutoFix: false})
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if len(result.Errors) == 0 {
		t.Error("ImportFile() should report invalid filename error")
	}
}

func TestImportFileInvalidFilenameAutoFix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "My Note.md")
	if err := os.WriteFile(path, []byte(validDoc("My Note", "kb")), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ImportFile(path, "", Options{AutoFix: true})
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if !result.Fixed {
		t.Error("ImportFile() Fixed = false, want true")
	}
	// Renamed file should exist.
	if _, err := os.Stat(filepath.Join(dir, "my-note.md")); err != nil {
		t.Errorf("renamed file not found: %v", err)
	}
}

func TestImportFileMissingFrontmatterNoAutoFix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bare-note.md")
	if err := os.WriteFile(path, []byte("# No frontmatter\n\nJust a body.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ImportFile(path, "", Options{AutoFix: false})
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if len(result.Errors) == 0 {
		t.Error("ImportFile() should report missing frontmatter")
	}
}

func TestImportFileMissingFrontmatterAutoFix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bare-note.md")
	if err := os.WriteFile(path, []byte("# Title\n\nBody text.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ImportFile(path, "", Options{AutoFix: true})
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if !result.Fixed {
		t.Error("ImportFile() Fixed = false for scaffolded frontmatter")
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "---") {
		t.Error("auto-fixed file should have frontmatter delimiters")
	}
	if !strings.Contains(string(data), "type: unsorted") {
		t.Error("scaffolded frontmatter should have type: unsorted")
	}
}

func TestImportFileDryRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "My Note.md")
	if err := os.WriteFile(path, []byte(validDoc("My Note", "kb")), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ImportFile(path, "", Options{AutoFix: true, DryRun: true})
	if err != nil {
		t.Fatalf("ImportFile() dry-run error = %v", err)
	}

	// File should not have been renamed.
	if _, err := os.Stat(path); err != nil {
		t.Error("dry-run should not rename file")
	}
}

func TestImportFileDuplicate(t *testing.T) {
	dir := t.TempDir()
	baseDir := t.TempDir()

	existing := filepath.Join(baseDir, "my-note.md")
	if err := os.WriteFile(existing, []byte(validDoc("My Note", "kb")), 0o644); err != nil {
		t.Fatal(err)
	}

	incoming := filepath.Join(dir, "my-note.md")
	if err := os.WriteFile(incoming, []byte(validDoc("My Note v2", "kb")), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ImportFile(incoming, baseDir, Options{})
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if !result.Skipped {
		t.Error("ImportFile() Skipped = false for duplicate")
	}
	if len(result.Errors) == 0 {
		t.Error("ImportFile() should report duplicate error")
	}
}

func TestImportFileIsDirectory(t *testing.T) {
	_, err := ImportFile(t.TempDir(), "", Options{})
	if err == nil {
		t.Error("ImportFile() on directory: expected error, got nil")
	}
}

func TestImportDir(t *testing.T) {
	inboxDir := t.TempDir()
	baseDir := t.TempDir()

	files := map[string]string{
		"note-one.md": validDoc("Note One", "kb"),
		"note-two.md": validDoc("Note Two", "kb"),
		"readme.txt":  "not a markdown file",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(inboxDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	results, err := ImportDir(inboxDir, baseDir, Options{})
	if err != nil {
		t.Fatalf("ImportDir() error = %v", err)
	}
	if len(results) != 2 {
		t.Errorf("ImportDir() returned %d results, want 2", len(results))
	}
}

func TestScaffoldFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "my-note.md")
	if err := os.WriteFile(path, []byte("body"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	fm := scaffoldFrontmatter("my-note", info.ModTime())
	if fm.Title != "My note" {
		t.Errorf("scaffoldFrontmatter() Title = %q, want %q", fm.Title, "My note")
	}
	if fm.Type != "unsorted" {
		t.Errorf("scaffoldFrontmatter() Type = %q, want unsorted", fm.Type)
	}
	if fm.Tags == nil {
		t.Error("Tags is nil")
	}
	if fm.Created.IsZero() {
		t.Error("Created is zero")
	}
	if fm.Modified.IsZero() {
		t.Error("Modified is zero")
	}
}

func TestFixFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.md")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)

	fm := fixFrontmatter(document.Frontmatter{}, "my-note", info.ModTime())
	if fm.Title == "" {
		t.Error("fixFrontmatter() Title is empty")
	}
	if fm.Type == "" {
		t.Error("fixFrontmatter() Type is empty")
	}
}

func TestStripExistingFrontmatter(t *testing.T) {
	tests := []struct {
		content string
		want    string
	}{
		{"---\ntitle: T\n---\n\nBody\n", "Body\n"},
		{"# No frontmatter\n", "# No frontmatter\n"},
		{"---\ntitle: T\n---\n", ""},
	}
	for _, tt := range tests {
		got := stripExistingFrontmatter(tt.content)
		if got != tt.want {
			t.Errorf("stripExistingFrontmatter(%q) = %q, want %q", tt.content, got, tt.want)
		}
	}
}
