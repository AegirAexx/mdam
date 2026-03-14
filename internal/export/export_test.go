package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStrip(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name: "standard frontmatter",
			content: `---
title: My Doc
type: kb
---

# Heading

Body text.
`,
			want:    "# Heading\n\nBody text.\n",
			wantErr: false,
		},
		{
			name:    "no frontmatter",
			content: "# Just markdown\n",
			wantErr: true,
		},
		{
			name: "no closing delimiter",
			content: `---
title: oops
# body
`,
			wantErr: true,
		},
		{
			name: "empty body",
			content: `---
title: Empty
---
`,
			want:    "",
			wantErr: false,
		},
		{
			name: "body with leading newlines trimmed",
			content: `---
title: Test
---


First paragraph.
`,
			want:    "First paragraph.\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Strip(tt.content)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Strip() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("Strip() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToFile(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	content := `---
title: Export Test
tags: []
type: kb
---

# Export Test

Some content here.
`
	srcPath := filepath.Join(srcDir, "export-test.md")
	if err := os.WriteFile(srcPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	destPath, err := ToFile(srcPath, destDir)
	if err != nil {
		t.Fatalf("ToFile() error = %v", err)
	}

	if filepath.Base(destPath) != "export-test.md" {
		t.Errorf("ToFile() destPath = %q, want export-test.md filename", destPath)
	}

	result, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("reading exported file: %v", err)
	}

	body := string(result)
	if strings.Contains(body, "---") {
		t.Errorf("ToFile() result still contains frontmatter delimiter")
	}
	if !strings.Contains(body, "# Export Test") {
		t.Errorf("ToFile() result missing body: %q", body)
	}
}

func TestToFileSourceNotFound(t *testing.T) {
	_, err := ToFile("/does/not/exist.md", t.TempDir())
	if err == nil {
		t.Error("ToFile() on missing source should return error")
	}
}

func TestToFileCreatesDestDir(t *testing.T) {
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "note.md")
	content := "---\ntitle: T\n---\nBody\n"
	if err := os.WriteFile(srcPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(t.TempDir(), "new", "nested", "dir")
	_, err := ToFile(srcPath, destDir)
	if err != nil {
		t.Fatalf("ToFile() should create nested destDir: %v", err)
	}
	if _, err := os.Stat(destDir); err != nil {
		t.Errorf("destDir not created: %v", err)
	}
}
