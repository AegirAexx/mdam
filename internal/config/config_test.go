package config

import (
	"os"
	"path/filepath"
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
	if cfg.Todo.ArchiveAfterDays != 30 {
		t.Errorf("ArchiveAfterDays = %d, want 30", cfg.Todo.ArchiveAfterDays)
	}
	if !cfg.Git.Enabled {
		t.Error("Git.Enabled = false, want true")
	}
	if cfg.Git.AutoCommit {
		t.Error("Git.AutoCommit = true, want false")
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

import:
  inbox_dir: ~/docs/.inbox
  auto_fix: true

git:
  enabled: true
  auto_commit: true
  lazygit: false

todo:
  default_category: work
  archive_after_days: 60

journal:
  auto_create: false
  sweep_on_create: false
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
	if cfg.Import.AutoFix != true {
		t.Error("Import.AutoFix = false, want true")
	}
	if cfg.Git.AutoCommit != true {
		t.Error("Git.AutoCommit = false, want true")
	}
	if cfg.Git.Lazygit != false {
		t.Error("Git.Lazygit = true, want false")
	}
	if cfg.Todo.DefaultCategory != "work" {
		t.Errorf("Todo.DefaultCategory = %q, want work", cfg.Todo.DefaultCategory)
	}
	if cfg.Todo.ArchiveAfterDays != 60 {
		t.Errorf("Todo.ArchiveAfterDays = %d, want 60", cfg.Todo.ArchiveAfterDays)
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
	home, _ := os.UserHomeDir()
	cfg := Config{}
	want := filepath.Join(home, ".config", "mdam", "pins.json")
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
	if cfg.TodoDir() != "/base/todo" {
		t.Errorf("TodoDir = %q", cfg.TodoDir())
	}
	if cfg.TodoPath() != "/base/todo/todo.md" {
		t.Errorf("TodoPath = %q", cfg.TodoPath())
	}
	if cfg.ScratchDir() != "/base/scratch" {
		t.Errorf("ScratchDir = %q", cfg.ScratchDir())
	}
	if cfg.ScratchPath() != "/base/scratch/scratch.md" {
		t.Errorf("ScratchPath = %q", cfg.ScratchPath())
	}
	if cfg.ArchivePath() != "/base/todo/archive.md" {
		t.Errorf("ArchivePath = %q", cfg.ArchivePath())
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
