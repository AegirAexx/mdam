package document

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid kebab", "my-document", false},
		{"valid single word", "notes", false},
		{"valid with digits", "notes-2026", false},
		{"valid journal date", "2026-03-14", false},
		{"uppercase", "MyDoc", true},
		{"spaces", "my doc", true},
		{"leading hyphen", "-doc", true},
		{"trailing hyphen", "doc-", true},
		{"double hyphen", "my--doc", true},
		{"underscore", "my_doc", true},
		{"empty", "", true},
		{"dot in name", "my.doc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilename(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilename(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Document", "my-document"},
		{"setup NGINX server", "setup-nginx-server"},
		{"  leading spaces  ", "leading-spaces"},
		{"multiple   spaces", "multiple-spaces"},
		{"under_score", "under-score"},
		{"already-kebab", "already-kebab"},
		{"123 numbers 456", "123-numbers-456"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToKebabCase(tt.input)
			if got != tt.want {
				t.Errorf("ToKebabCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFrontmatter(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name      string
		content   string
		wantTitle string
		wantBody  string
		wantErr   bool
	}{
		{
			name: "valid full frontmatter",
			content: `---
title: My Note
tags:
  - go
  - test
created: 2026-03-14T10:00:00Z
modified: 2026-03-14T10:00:00Z
type: kb
---

# Body here
`,
			wantTitle: "My Note",
			wantBody:  "# Body here\n",
			wantErr:   false,
		},
		{
			name:    "missing opening delimiter",
			content: "title: foo\n---\n",
			wantErr: true,
		},
		{
			name:    "missing closing delimiter",
			content: "---\ntitle: foo\n",
			wantErr: true,
		},
		{
			name:    "invalid YAML",
			content: "---\ntitle: [\nbad yaml\n---\n",
			wantErr: true,
		},
		{
			name: "empty body",
			content: `---
title: No Body
tags: []
created: 2026-03-14T10:00:00Z
modified: 2026-03-14T10:00:00Z
type: scratch
---
`,
			wantTitle: "No Body",
			wantBody:  "",
			wantErr:   false,
		},
	}

	_ = now
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := parseFrontmatter([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if fm.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", fm.Title, tt.wantTitle)
			}
			if body != tt.wantBody {
				t.Errorf("Body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestValidateFrontmatter(t *testing.T) {
	good := Frontmatter{
		Title:    "Test",
		Tags:     []string{"a"},
		Created:  time.Now(),
		Modified: time.Now(),
		Type:     "kb",
	}

	tests := []struct {
		name    string
		fm      Frontmatter
		wantErr bool
	}{
		{"valid", good, false},
		{"missing title", func() Frontmatter { f := good; f.Title = ""; return f }(), true},
		{"nil tags", func() Frontmatter { f := good; f.Tags = nil; return f }(), true},
		{"zero created", func() Frontmatter { f := good; f.Created = time.Time{}; return f }(), true},
		{"zero modified", func() Frontmatter { f := good; f.Modified = time.Time{}; return f }(), true},
		{"missing type", func() Frontmatter { f := good; f.Type = ""; return f }(), true},
		{"invalid type", func() Frontmatter { f := good; f.Type = "bogus"; return f }(), true},
		{"type journal", func() Frontmatter { f := good; f.Type = "journal"; return f }(), false},
		{"type todo", func() Frontmatter { f := good; f.Type = "todo"; return f }(), false},
		{"type scratch", func() Frontmatter { f := good; f.Type = "scratch"; return f }(), false},
		{"type unsorted", func() Frontmatter { f := good; f.Type = "unsorted"; return f }(), false},
		{"whitespace title", func() Frontmatter { f := good; f.Title = "   "; return f }(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFrontmatter(tt.fm)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRenderFrontmatterFieldOrder(t *testing.T) {
	fm := Frontmatter{
		Type:     "kb",
		Title:    "Test",
		Tags:     []string{},
		Created:  time.Now(),
		Modified: time.Now(),
	}
	out, err := RenderFrontmatter(fm)
	if err != nil {
		t.Fatalf("RenderFrontmatter() error = %v", err)
	}
	// Walk lines between the --- delimiters and collect keys in order.
	wantOrder := []string{"type:", "title:", "tags:", "created:", "modified:"}
	var foundOrder []string
	for _, line := range strings.Split(out, "\n") {
		for _, key := range wantOrder {
			if strings.HasPrefix(line, key) {
				foundOrder = append(foundOrder, key)
				break
			}
		}
	}
	if len(foundOrder) != len(wantOrder) {
		t.Fatalf("RenderFrontmatter() found keys %v, want %v", foundOrder, wantOrder)
	}
	for i := range wantOrder {
		if foundOrder[i] != wantOrder[i] {
			t.Errorf("field order[%d] = %q, want %q", i, foundOrder[i], wantOrder[i])
		}
	}
}

func TestParseFileAndWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-note.md")

	content := `---
title: Test Note
tags:
  - testing
created: 2026-03-14T10:00:00Z
modified: 2026-03-14T10:00:00Z
type: kb
---

# Hello
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	if doc.Frontmatter.Title != "Test Note" {
		t.Errorf("Title = %q, want %q", doc.Frontmatter.Title, "Test Note")
	}
	if doc.Frontmatter.Type != "kb" {
		t.Errorf("Type = %q, want kb", doc.Frontmatter.Type)
	}
	if len(doc.Frontmatter.Tags) != 1 || doc.Frontmatter.Tags[0] != "testing" {
		t.Errorf("Tags = %v, want [testing]", doc.Frontmatter.Tags)
	}

	// Round-trip: write and re-parse.
	doc.Frontmatter.Title = "Updated Note"
	if err := doc.Write(); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	doc2, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() after Write() error = %v", err)
	}
	if doc2.Frontmatter.Title != "Updated Note" {
		t.Errorf("After write, Title = %q, want Updated Note", doc2.Frontmatter.Title)
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := ParseFile("/does/not/exist.md")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestExtraFields(t *testing.T) {
	content := `---
title: KB Note
tags: []
created: 2026-03-14T10:00:00Z
modified: 2026-03-14T10:00:00Z
type: kb
kb_type: howto
---
`
	fm, _, err := parseFrontmatter([]byte(content))
	if err != nil {
		t.Fatalf("parseFrontmatter() error = %v", err)
	}
	if fm.Extra["kb_type"] != "howto" {
		t.Errorf("Extra[kb_type] = %v, want howto", fm.Extra["kb_type"])
	}
}
