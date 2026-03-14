package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscover(t *testing.T) {
	dir := t.TempDir()

	// Write a couple of template files.
	files := map[string]string{
		"journal.md": "---\ntitle: Journal\n---\n",
		"kb.md":      "---\ntitle: KB\n---\n",
		"README.txt": "not a template",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	templates, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(templates) != 2 {
		t.Errorf("Discover() returned %d templates, want 2", len(templates))
	}
	names := map[string]bool{}
	for _, tmpl := range templates {
		names[tmpl.Name] = true
	}
	if !names["journal"] || !names["kb"] {
		t.Errorf("Discover() names = %v, want journal and kb", names)
	}
}

func TestDiscoverNonExistentDir(t *testing.T) {
	templates, err := Discover("/does/not/exist")
	if err != nil {
		t.Fatalf("Discover() on missing dir error = %v", err)
	}
	if len(templates) != 0 {
		t.Errorf("Discover() returned %d templates, want 0", len(templates))
	}
}

func TestFind(t *testing.T) {
	dir := t.TempDir()
	content := "---\ntitle: Test\n---\nBody"
	if err := os.WriteFile(filepath.Join(dir, "test.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	tmpl, err := Find(dir, "test")
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if tmpl.Name != "test" {
		t.Errorf("Name = %q, want test", tmpl.Name)
	}
	if tmpl.Content != content {
		t.Errorf("Content mismatch")
	}

	_, err = Find(dir, "nonexistent")
	if err == nil {
		t.Error("Find() nonexistent should return error")
	}
}

func TestRender(t *testing.T) {
	tmpl := Template{
		Name:    "test",
		Content: "Hello {{title}}, date is {{date_short}}, author: {{author}}",
	}

	rendered, err := Render(tmpl, map[string]string{
		"title":  "My Doc",
		"author": "Alice",
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.Contains(rendered, "My Doc") {
		t.Errorf("Render() missing title: %q", rendered)
	}
	if !strings.Contains(rendered, "Alice") {
		t.Errorf("Render() missing author: %q", rendered)
	}
	// date_short should be resolved automatically (YYYY-MM-DD format).
	if strings.Contains(rendered, "{{date_short}}") {
		t.Errorf("Render() did not resolve {{date_short}}")
	}
	if strings.Contains(rendered, "{{date}}") {
		t.Errorf("Render() did not resolve {{date}}")
	}
}

func TestRenderUnresolvedVars(t *testing.T) {
	tmpl := Template{
		Name:    "test",
		Content: "Hello {{title}} and {{custom_var}}",
	}
	rendered, err := Render(tmpl, map[string]string{
		"title": "Doc",
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	// {{custom_var}} should remain unresolved.
	if !strings.Contains(rendered, "{{custom_var}}") {
		t.Errorf("Render() should leave unresolved custom_var in output")
	}
}

func TestUnresolvedVars(t *testing.T) {
	tests := []struct {
		content string
		want    []string
	}{
		{"no vars here", nil},
		{"{{title}} and {{author}}", []string{"{{title}}", "{{author}}"}},
		{"already {{resolved}}", []string{"{{resolved}}"}},
		{"", nil},
	}
	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			got := UnresolvedVars(tt.content)
			if len(got) != len(tt.want) {
				t.Errorf("UnresolvedVars(%q) = %v, want %v", tt.content, got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("UnresolvedVars()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestWriteBuiltins(t *testing.T) {
	dir := t.TempDir()
	if err := WriteBuiltins(dir); err != nil {
		t.Fatalf("WriteBuiltins() error = %v", err)
	}

	expected := []string{"journal.md", "kb.md", "howto.md", "meeting.md", "scratch.md"}
	for _, name := range expected {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("WriteBuiltins() missing file %s", name)
		}
	}

	// Running again should not error (skip existing files).
	if err := WriteBuiltins(dir); err != nil {
		t.Fatalf("WriteBuiltins() second run error = %v", err)
	}
}

func TestBuiltinTemplatesContent(t *testing.T) {
	builtins := BuiltinTemplates()
	for name, content := range builtins {
		if !strings.HasPrefix(content, "---") {
			t.Errorf("builtin template %q does not start with frontmatter delimiter", name)
		}
	}
}
