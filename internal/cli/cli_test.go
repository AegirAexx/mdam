package cli

import (
	"bytes"
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
