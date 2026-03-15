package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AegirAexx/mdam/internal/config"
)

func TestIsFirstRun(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")

	// Missing file + empty BaseDir → true
	cfg := config.Config{}
	if !IsFirstRun(cfgPath, cfg) {
		t.Error("IsFirstRun = false, want true (missing file + empty base_dir)")
	}

	// File exists but BaseDir still empty → true
	if err := os.WriteFile(cfgPath, []byte("base_dir: \"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !IsFirstRun(cfgPath, cfg) {
		t.Error("IsFirstRun = false, want true (file exists + empty base_dir)")
	}

	// File exists + BaseDir set to existing dir → false
	cfg.BaseDir = dir
	if IsFirstRun(cfgPath, cfg) {
		t.Error("IsFirstRun = true, want false (file exists + valid base_dir)")
	}

	// File exists + BaseDir set to non-existent dir → true
	cfg.BaseDir = filepath.Join(dir, "nonexistent")
	if !IsFirstRun(cfgPath, cfg) {
		t.Error("IsFirstRun = false, want true (base_dir does not exist)")
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.yml")

	// Creates file with parent dirs.
	if err := WriteDefaultConfig(path); err != nil {
		t.Fatalf("WriteDefaultConfig() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading created config: %v", err)
	}
	if !strings.Contains(string(data), "base_dir:") {
		t.Error("config missing base_dir field")
	}

	// Second call is no-op (content unchanged).
	if err := WriteDefaultConfig(path); err != nil {
		t.Fatalf("WriteDefaultConfig() second call error = %v", err)
	}
	data2, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Error("WriteDefaultConfig() second call modified the file")
	}
}

func TestUpdateConfigBaseDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	content := "base_dir: \"\"\ntheme: tokyonight\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := updateConfigBaseDir(path, "/my/notes"); err != nil {
		t.Fatalf("updateConfigBaseDir() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "base_dir: /my/notes") {
		t.Errorf("updateConfigBaseDir() result = %q, missing base_dir: /my/notes", string(data))
	}
	if !strings.Contains(string(data), "theme: tokyonight") {
		t.Error("updateConfigBaseDir() removed other fields")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.Config
		wantLen  int
	}{
		{
			name:    "valid config",
			cfg:     config.Config{Theme: "tokyonight"},
			wantLen: 0,
		},
		{
			name:    "unknown theme",
			cfg:     config.Config{Theme: "unknowntheme"},
			wantLen: 1,
		},
		{
			name:    "empty theme",
			cfg:     config.Config{Theme: ""},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateConfig(tt.cfg)
			if len(got) != tt.wantLen {
				t.Errorf("ValidateConfig() len = %d, want %d (warnings: %v)", len(got), tt.wantLen, got)
			}
		})
	}
}

func TestPromptBaseDir(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty input defaults to ~/notes",
			input: "\n",
			want:  filepath.Join(home, "notes"),
		},
		{
			name:  "tilde expansion",
			input: "~/docs\n",
			want:  filepath.Join(home, "docs"),
		},
		{
			name:  "absolute path",
			input: "/tmp/mynotess\n",
			want:  "/tmp/mynotess",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			var w strings.Builder
			got, err := PromptBaseDir(r, &w)
			if err != nil {
				t.Fatalf("PromptBaseDir() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("PromptBaseDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestScaffoldDirs(t *testing.T) {
	base := t.TempDir()
	if err := ScaffoldDirs(base); err != nil {
		t.Fatalf("ScaffoldDirs() error = %v", err)
	}

	expected := []string{"journal", "kb", "todo", "scratch", ".inbox", ".templates"}
	for _, d := range expected {
		path := filepath.Join(base, d)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("ScaffoldDirs() missing dir %s", path)
		}
	}

	// Second call must not error.
	if err := ScaffoldDirs(base); err != nil {
		t.Fatalf("ScaffoldDirs() second call error = %v", err)
	}
}

func TestSeedTemplates(t *testing.T) {
	dir := t.TempDir()
	if err := SeedTemplates(dir); err != nil {
		t.Fatalf("SeedTemplates() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Error("SeedTemplates() wrote no files")
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			t.Errorf("SeedTemplates() unexpected file %s", e.Name())
		}
	}

	// Collect content before second call.
	snapshots := map[string][]byte{}
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		data, _ := os.ReadFile(p)
		snapshots[e.Name()] = data
	}

	// Second call must not error and must not change existing files.
	if err := SeedTemplates(dir); err != nil {
		t.Fatalf("SeedTemplates() second call error = %v", err)
	}
	for name, before := range snapshots {
		after, _ := os.ReadFile(filepath.Join(dir, name))
		if string(before) != string(after) {
			t.Errorf("SeedTemplates() second call modified %s", name)
		}
	}
}

func TestEnsureScratch(t *testing.T) {
	dir := t.TempDir()
	scratchPath := filepath.Join(dir, "scratch", "scratch.md")

	// Creates file with correct frontmatter.
	if err := EnsureScratch(scratchPath); err != nil {
		t.Fatalf("EnsureScratch() error = %v", err)
	}
	data, err := os.ReadFile(scratchPath)
	if err != nil {
		t.Fatalf("reading scratch: %v", err)
	}
	if !strings.Contains(string(data), "type: scratch") {
		t.Errorf("EnsureScratch() missing type: scratch in %q", string(data))
	}

	// Second call must not overwrite existing file.
	original := string(data)
	if err := EnsureScratch(scratchPath); err != nil {
		t.Fatalf("EnsureScratch() second call error = %v", err)
	}
	data2, _ := os.ReadFile(scratchPath)
	if original != string(data2) {
		t.Error("EnsureScratch() second call overwrote existing file")
	}
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	notesDir := filepath.Join(dir, "testdocs")

	// Write a minimal config with empty base_dir.
	initial := "base_dir: \"\"\ntheme: tokyonight\n"
	if err := os.WriteFile(cfgPath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{}
	input := strings.NewReader(notesDir + "\ny\n")
	var out strings.Builder

	updated, err := Run(cfgPath, cfg, input, &out)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if updated.BaseDir != notesDir {
		t.Errorf("Run() BaseDir = %q, want %q", updated.BaseDir, notesDir)
	}

	// Dirs were scaffolded.
	for _, d := range []string{"journal", "kb", "todo", "scratch", ".inbox", ".templates"} {
		if _, err := os.Stat(filepath.Join(notesDir, d)); os.IsNotExist(err) {
			t.Errorf("Run() missing dir %s", d)
		}
	}

	// Scratch file exists.
	scratchPath := filepath.Join(notesDir, "scratch", "scratch.md")
	if _, err := os.Stat(scratchPath); os.IsNotExist(err) {
		t.Error("Run() scratch file not created")
	}
}
