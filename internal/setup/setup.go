// Package setup handles first-run initialization for mdam: config creation,
// base directory prompting, folder scaffolding, template seeding, and scratch pad creation.
package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AegirAexx/mdam/internal/config"
	"github.com/AegirAexx/mdam/internal/template"
)

const defaultConfigYAML = `# MadaM configuration — generated on first run.

# Text editor to open documents (defaults to $EDITOR).
editor: ""

# Your name, used for document metadata.
author: ""

# Root directory where MadaM manages your documents.
# Leave empty to be prompted on every startup until set.
base_dir: ""

# Directory for exported documents (frontmatter stripped).
export_dir: ~/Downloads

# TUI color theme. Options: tokyonight, nord, gruvbox, catppuccin, dracula
theme: tokyonight

# Use Nerd Font icons in the TUI (requires a patched terminal font).
nerd_fonts: false

import:
  # Directory for files dropped in for import. Defaults to {base_dir}/.inbox.
  inbox_dir: ""
  # Auto-fix invalid filenames and frontmatter during import.
  auto_fix: false

git:
  # Enable git integration (shows modified/untracked indicators).
  enabled: true
  # Automatically commit after edits.
  auto_commit: false
  # Use lazygit for git operations.
  lazygit: true

todo:
  # Default category for new TODO items.
  default_category: personal
  # Days before completed tasks are moved to archive.
  archive_after_days: 30

journal:
  # Automatically create today's journal entry on startup.
  auto_create: true
  # Sweep incomplete tasks from past entries when creating a new journal.
  sweep_on_create: true
`

var validThemes = map[string]bool{
	"tokyonight": true,
	"nord":       true,
	"gruvbox":    true,
	"catppuccin": true,
	"dracula":    true,
}

// IsFirstRun returns true if the config file is missing OR base_dir is empty/non-existent.
func IsFirstRun(cfgPath string, cfg config.Config) bool {
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return true
	}
	if cfg.BaseDir == "" {
		return true
	}
	if _, err := os.Stat(cfg.BaseDir); os.IsNotExist(err) {
		return true
	}
	return false
}

// WriteDefaultConfig creates parent dirs and writes a commented config.yml.
// No-op if the file already exists.
func WriteDefaultConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(defaultConfigYAML), 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// ValidateConfig returns human-readable warnings for invalid/missing config values.
func ValidateConfig(cfg config.Config) []string {
	var warnings []string
	if cfg.Theme != "" && !validThemes[cfg.Theme] {
		warnings = append(warnings, fmt.Sprintf("unknown theme %q — valid options: tokyonight, nord, gruvbox, catppuccin, dracula", cfg.Theme))
	}
	return warnings
}

// PromptBaseDir writes a prompt to w, reads a line from r, expands ~, defaults to ~/notes.
func PromptBaseDir(r io.Reader, w io.Writer) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	defaultDir := filepath.Join(home, "notes")

	fmt.Fprintf(w, "mdam: Enter the base directory for your documents [%s]: ", defaultDir)
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	line := strings.TrimSpace(scanner.Text())
	if line == "" {
		return defaultDir, nil
	}
	return expandHome(line), nil
}

// ScaffoldDirs creates the 6 expected subdirs under baseDir. Idempotent.
func ScaffoldDirs(baseDir string) error {
	dirs := []string{
		filepath.Join(baseDir, "journal"),
		filepath.Join(baseDir, "kb"),
		filepath.Join(baseDir, "todo"),
		filepath.Join(baseDir, "scratch"),
		filepath.Join(baseDir, ".inbox"),
		filepath.Join(baseDir, ".templates"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating %s: %w", d, err)
		}
	}
	return nil
}

// SeedTemplates copies built-in templates to templatesDir. No-op if all exist.
func SeedTemplates(templatesDir string) error {
	return template.WriteBuiltins(templatesDir)
}

// EnsureScratch creates scratchPath with valid frontmatter if it doesn't exist.
// Creates parent directory if needed.
func EnsureScratch(scratchPath string) error {
	if err := os.MkdirAll(filepath.Dir(scratchPath), 0o755); err != nil {
		return fmt.Errorf("creating scratch dir: %w", err)
	}
	if _, err := os.Stat(scratchPath); err == nil {
		return nil // already exists
	}
	now := time.Now().UTC().Format("2006-01-02")
	content := fmt.Sprintf("---\ntype: scratch\ntitle: Scratch Pad\ntags: []\ncreated: %s\nmodified: %s\n---\n", now, now)
	if err := os.WriteFile(scratchPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing scratch: %w", err)
	}
	return nil
}

// Run orchestrates the full first-run flow, returning the updated config.
func Run(cfgPath string, cfg config.Config, r io.Reader, w io.Writer) (config.Config, error) {
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := WriteDefaultConfig(cfgPath); err != nil {
			return cfg, fmt.Errorf("writing default config: %w", err)
		}
	}

	if cfg.BaseDir == "" {
		baseDir, err := PromptBaseDir(r, w)
		if err != nil {
			return cfg, fmt.Errorf("prompting base dir: %w", err)
		}

		ok, err := promptYN(r, w, fmt.Sprintf("Create folder structure at %s? (Y/n): ", baseDir))
		if err != nil {
			return cfg, fmt.Errorf("prompting scaffold: %w", err)
		}

		if ok {
			if err := ScaffoldDirs(baseDir); err != nil {
				return cfg, fmt.Errorf("scaffolding dirs: %w", err)
			}
			templatesDir := filepath.Join(baseDir, ".templates")
			if err := SeedTemplates(templatesDir); err != nil {
				return cfg, fmt.Errorf("seeding templates: %w", err)
			}
			scratchPath := filepath.Join(baseDir, "scratch", "scratch.md")
			if err := EnsureScratch(scratchPath); err != nil {
				return cfg, fmt.Errorf("ensuring scratch: %w", err)
			}
		}

		if err := updateConfigBaseDir(cfgPath, baseDir); err != nil {
			return cfg, fmt.Errorf("updating config: %w", err)
		}

		updated, err := config.LoadFrom(cfgPath)
		if err != nil {
			return cfg, fmt.Errorf("reloading config: %w", err)
		}
		return updated, nil
	}

	return cfg, nil
}

// promptYN reads y/n/Enter from r, defaulting to true (yes).
func promptYN(r io.Reader, w io.Writer, prompt string) (bool, error) {
	fmt.Fprint(w, prompt)
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("reading input: %w", err)
	}
	line := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return line == "" || line == "y" || line == "yes", nil
}

// updateConfigBaseDir rewrites the base_dir line in the YAML config file.
func updateConfigBaseDir(cfgPath, baseDir string) error {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	replaced := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "base_dir:") {
			lines[i] = "base_dir: " + baseDir
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, "base_dir: "+baseDir)
	}
	if err := os.WriteFile(cfgPath, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// expandHome expands a leading ~/ to the user's home directory.
func expandHome(path string) string {
	if len(path) < 2 || path[:2] != "~/" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}
