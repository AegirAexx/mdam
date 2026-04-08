package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Load from a non-existent path — should return defaults without error.
	cfg, err := LoadFrom("/does/not/exist/config.yml")
	if err != nil {
		t.Fatalf("LoadFrom non-existent path error = %v", err)
	}
	if cfg.Theme != "tokyonight" {
		t.Errorf("Theme = %q, want tokyonight", cfg.Theme)
	}
	if !cfg.Journal.AutoCreate {
		t.Error("Journal.AutoCreate = false, want true")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
editor: vim
author: "Test User"
base_dir: ~/docs
export_dir: ~/Desktop
theme: nord

journal:
  auto_create: false
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if cfg.Editor != "vim" {
		t.Errorf("Editor = %q, want vim", cfg.Editor)
	}
	if cfg.Author != "Test User" {
		t.Errorf("Author = %q, want Test User", cfg.Author)
	}
	if cfg.Theme != "nord" {
		t.Errorf("Theme = %q, want nord", cfg.Theme)
	}
	if cfg.Journal.AutoCreate {
		t.Error("Journal.AutoCreate = true, want false")
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		input string
		want  string
	}{
		{"~/notes", filepath.Join(home, "notes")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~", "~"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandHome(tt.input)
			if got != tt.want {
				t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNerdFontsDefault(t *testing.T) {
	cfg, err := LoadFrom("/does/not/exist/config.yml")
	if err != nil {
		t.Fatalf("LoadFrom error = %v", err)
	}
	if cfg.NerdFonts {
		t.Error("NerdFonts default = true, want false")
	}
}

func TestPinsPath(t *testing.T) {
	cfg := Config{BaseDir: "/base"}
	want := filepath.Join("/base", ".mdam", "pins.json")
	if got := cfg.PinsPath(); got != want {
		t.Errorf("PinsPath() = %q, want %q", got, want)
	}
}

func TestConfigDirs(t *testing.T) {
	cfg := Config{BaseDir: "/base"}
	if cfg.JournalDir() != "/base/journal" {
		t.Errorf("JournalDir = %q", cfg.JournalDir())
	}
	if cfg.KBDir() != "/base/kb" {
		t.Errorf("KBDir = %q", cfg.KBDir())
	}
	if cfg.TemplatesDir() != "/base/.templates" {
		t.Errorf("TemplatesDir = %q", cfg.TemplatesDir())
	}
	if cfg.TodoPath() != "/base/todo.md" {
		t.Errorf("TodoPath = %q", cfg.TodoPath())
	}
	if cfg.ScratchPath() != "/base/scratch.md" {
		t.Errorf("ScratchPath = %q", cfg.ScratchPath())
	}
}

func TestLoadDefaultsBaseDir(t *testing.T) {
	cfg, err := LoadFrom("/does/not/exist/config.yml")
	if err != nil {
		t.Fatalf("LoadFrom error = %v", err)
	}
	if cfg.BaseDir != "" {
		t.Errorf("BaseDir default = %q, want empty string", cfg.BaseDir)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath() error = %v", err)
	}
	if !strings.Contains(path, "mdam") || !strings.HasSuffix(path, "config.yml") {
		t.Errorf("DefaultConfigPath() = %q, want path containing mdam/config.yml", path)
	}
}

func TestLoadFromInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yml")
	if err := os.WriteFile(path, []byte("not: valid: yaml: [["), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFrom(path)
	if err == nil {
		t.Error("LoadFrom() with invalid YAML should return error, got nil")
	}
}
