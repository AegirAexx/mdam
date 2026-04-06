package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWizardConfigYAML(t *testing.T) {
	m := WizardModel{
		editor:    "nvim",
		author:    "Test User",
		baseDir:   "/home/test/notes",
		exportDir: "/home/test/Downloads",
		themeName: "nord",
		nerdFonts: true,
	}

	yaml := wizardConfigYAML(m)

	checks := []string{
		`editor: "nvim"`,
		`author: "Test User"`,
		"base_dir: /home/test/notes",
		"export_dir: /home/test/Downloads",
		"theme: nord",
		"nerd_fonts: true",
		"auto_create: true",
	}
	for _, check := range checks {
		if !strings.Contains(yaml, check) {
			t.Errorf("wizardConfigYAML() missing %q in:\n%s", check, yaml)
		}
	}
}

func TestWizardConfigYAMLDefaults(t *testing.T) {
	m := WizardModel{
		themeName: "tokyonight",
		nerdFonts: false,
	}

	yaml := wizardConfigYAML(m)

	if !strings.Contains(yaml, "theme: tokyonight") {
		t.Errorf("missing theme in:\n%s", yaml)
	}
	if !strings.Contains(yaml, "nerd_fonts: false") {
		t.Errorf("missing nerd_fonts in:\n%s", yaml)
	}
}

func TestWizardStepNavigation(t *testing.T) {
	m := NewWizard("/tmp/test-config.yml")

	if m.step != stepBaseDir {
		t.Fatalf("initial step = %d, want stepBaseDir", m.step)
	}

	// Enter advances from baseDir to editor.
	result, _ := m.Update(keyMsg("enter"))
	m = result.(WizardModel)
	if m.step != stepEditor {
		t.Errorf("after first enter: step = %d, want stepEditor", m.step)
	}

	// Esc goes back.
	result, _ = m.Update(keyMsg("esc"))
	m = result.(WizardModel)
	if m.step != stepBaseDir {
		t.Errorf("after esc: step = %d, want stepBaseDir", m.step)
	}

	// Esc at first step is a no-op.
	result, _ = m.Update(keyMsg("esc"))
	m = result.(WizardModel)
	if m.step != stepBaseDir {
		t.Errorf("esc at step 0: step = %d, want stepBaseDir", m.step)
	}
}

func TestWizardThemeStep(t *testing.T) {
	m := NewWizard("/tmp/test-config.yml")
	// Advance to theme step.
	m.step = stepTheme
	m.cursor = 0
	m = m.syncInputToStep()

	// j moves cursor down.
	result, _ := m.Update(keyMsg("j"))
	m = result.(WizardModel)
	if m.cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", m.cursor)
	}
	if m.themeName != "nord" {
		t.Errorf("after j: themeName = %q, want nord", m.themeName)
	}

	// Enter confirms and advances.
	result, _ = m.Update(keyMsg("enter"))
	m = result.(WizardModel)
	if m.step != stepNerdFonts {
		t.Errorf("after enter on theme: step = %d, want stepNerdFonts", m.step)
	}
}

func TestWizardWriteAndScaffold(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config", "config.yml")
	baseDir := filepath.Join(dir, "notes")

	m := WizardModel{
		cfgPath:   cfgPath,
		baseDir:   baseDir,
		editor:    "vim",
		author:    "Test",
		themeName: "tokyonight",
		exportDir: filepath.Join(dir, "exports"),
	}

	if err := m.writeAndScaffold(); err != nil {
		t.Fatalf("writeAndScaffold() error = %v", err)
	}

	// Config file exists.
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("config file not created: %v", err)
	}

	// Dirs exist.
	for _, d := range []string{"journal", "kb", ".templates"} {
		if _, err := os.Stat(filepath.Join(baseDir, d)); err != nil {
			t.Errorf("missing dir %s: %v", d, err)
		}
	}

	// Singleton files exist.
	for _, f := range []string{"scratch.md", "todo.md"} {
		if _, err := os.Stat(filepath.Join(baseDir, f)); err != nil {
			t.Errorf("missing file %s: %v", f, err)
		}
	}

	// Templates seeded.
	templates, err := os.ReadDir(filepath.Join(baseDir, ".templates"))
	if err != nil {
		t.Fatalf("reading templates dir: %v", err)
	}
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
}

func TestWizardExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		input string
		want  string
	}{
		{"~/notes", filepath.Join(home, "notes")},
		{"/absolute/path", "/absolute/path"},
		{"relative", "relative"},
	}
	for _, tt := range tests {
		got := expandHome(tt.input)
		if got != tt.want {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// keyMsg creates a tea.KeyMsg for testing.
func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}
