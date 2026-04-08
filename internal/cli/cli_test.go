package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSubcommandsRegistered(t *testing.T) {
	names := []string{"journal", "todo", "search", "import", "export", "status", "template", "config"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{name})
			if err != nil {
				t.Fatalf("Find(%q) error = %v", name, err)
			}
			if cmd == nil || cmd.Name() == rootCmd.Name() {
				t.Errorf("subcommand %q not found (got root)", name)
			}
		})
	}
}

func TestJournalSubcommandsRegistered(t *testing.T) {
	for _, sub := range []string{"create", "list"} {
		t.Run(sub, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{"journal", sub})
			if err != nil {
				t.Fatalf("Find(journal %q) error = %v", sub, err)
			}
			if cmd == nil || cmd.Name() == rootCmd.Name() {
				t.Errorf("journal %q not found (got root)", sub)
			}
		})
	}
}

func TestTodoSubcommandsRegistered(t *testing.T) {
	for _, sub := range []string{"list", "sweep", "archive"} {
		t.Run(sub, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{"todo", sub})
			if err != nil {
				t.Fatalf("Find(todo %q) error = %v", sub, err)
			}
			if cmd == nil || cmd.Name() == rootCmd.Name() {
				t.Errorf("todo %q not found (got root)", sub)
			}
		})
	}
}

func TestRootHelpOutput(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})
	// --help exits with code 0 but cobra may return nil; ignore the error.
	_ = rootCmd.Execute()
	if !strings.Contains(buf.String(), "mdam") {
		t.Errorf("help output does not contain 'mdam': %q", buf.String())
	}
}

// writeTempConfig writes a minimal config file with the given base_dir and
// returns its path. The config avoids triggering the first-run wizard.
func writeTempConfig(t *testing.T, baseDir string) string {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	content := fmt.Sprintf("base_dir: %s\ntheme: tokyonight\n", baseDir)
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return cfgPath
}

func TestJournalListCommand(t *testing.T) {
	base := t.TempDir()
	journalDir := filepath.Join(base, "journal")
	if err := os.MkdirAll(journalDir, 0o755); err != nil {
		t.Fatal(err)
	}

	fm := func(title string) string {
		return "---\ntitle: " + title + "\ntags: []\ncreated: 2026-03-14T10:00:00Z\nmodified: 2026-03-14T10:00:00Z\ntype: journal\n---\n"
	}
	for _, name := range []string{"2026-03-14.md", "2026-03-15.md"} {
		if err := os.WriteFile(filepath.Join(journalDir, name), []byte(fm(name)), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cfgPath := writeTempConfig(t, base)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--config", cfgPath, "journal", "list"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("journal list error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "2026-03-14") {
		t.Errorf("output missing 2026-03-14: %q", out)
	}
	if !strings.Contains(out, "2026-03-15") {
		t.Errorf("output missing 2026-03-15: %q", out)
	}
}

func TestTodoListCommand(t *testing.T) {
	base := t.TempDir()
	todoPath := filepath.Join(base, "todo.md")
	content := "---\ntitle: Todo\ntags: []\ncreated: 2026-03-14T10:00:00Z\nmodified: 2026-03-14T10:00:00Z\ntype: todo\n---\n\n- [ ] Buy groceries\n- [ ] Write tests\n"
	if err := os.WriteFile(todoPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfgPath := writeTempConfig(t, base)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--config", cfgPath, "todo", "list"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("todo list error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Buy groceries") {
		t.Errorf("output missing 'Buy groceries': %q", out)
	}
	if !strings.Contains(out, "Write tests") {
		t.Errorf("output missing 'Write tests': %q", out)
	}
}
