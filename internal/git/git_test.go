package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initRepo creates a temporary git repository for testing.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init: %s: %v", string(out), err)
		}
	}
	return dir
}

func TestIsAvailable(t *testing.T) {
	// git should be available in the test environment.
	if !IsAvailable() {
		t.Skip("git not available")
	}
}

func TestIsRepo(t *testing.T) {
	dir := initRepo(t)
	if !IsRepo(dir) {
		t.Errorf("IsRepo() = false for initialized repo")
	}
	if IsRepo(t.TempDir()) {
		t.Errorf("IsRepo() = true for non-repo directory")
	}
}

func TestStatus(t *testing.T) {
	if !IsAvailable() {
		t.Skip("git not available")
	}
	dir := initRepo(t)

	// Create and commit an initial file.
	path := filepath.Join(dir, "README.md")
	if err := os.WriteFile(path, []byte("# Readme\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "init"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s: %v", args, string(out), err)
		}
	}

	// Clean state — no uncommitted changes.
	status, err := Status(dir)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.UncommittedCount() != 0 {
		t.Errorf("UncommittedCount() = %d, want 0", status.UncommittedCount())
	}

	// Add an untracked file.
	if err := os.WriteFile(filepath.Join(dir, "new.md"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err = Status(dir)
	if err != nil {
		t.Fatalf("Status() after new file error = %v", err)
	}
	if status.UncommittedCount() != 1 {
		t.Errorf("UncommittedCount() = %d, want 1", status.UncommittedCount())
	}
	if len(status.Files) < 1 || !status.Files[0].IsUntracked() {
		t.Errorf("expected untracked file, got %+v", status.Files)
	}
}

func TestStatusBranch(t *testing.T) {
	if !IsAvailable() {
		t.Skip("git not available")
	}
	dir := initRepo(t)

	// Need at least one commit to have a branch.
	path := filepath.Join(dir, "f.md")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "f.md"},
		{"git", "commit", "-m", "init"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s: %v", args, string(out), err)
		}
	}

	status, err := Status(dir)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Branch == "" {
		t.Error("Branch should not be empty")
	}
}

func TestFileStatusHelpers(t *testing.T) {
	tests := []struct {
		fs          FileStatus
		modified    bool
		untracked   bool
		staged      bool
	}{
		{FileStatus{X: '?', Y: '?'}, false, true, false},
		{FileStatus{X: ' ', Y: 'M'}, true, false, false},
		{FileStatus{X: 'M', Y: ' '}, false, false, true},
		{FileStatus{X: 'A', Y: ' '}, false, false, true},
	}
	for _, tt := range tests {
		if tt.fs.IsModified() != tt.modified {
			t.Errorf("IsModified() = %v, want %v for %+v", tt.fs.IsModified(), tt.modified, tt.fs)
		}
		if tt.fs.IsUntracked() != tt.untracked {
			t.Errorf("IsUntracked() = %v, want %v for %+v", tt.fs.IsUntracked(), tt.untracked, tt.fs)
		}
		if tt.fs.IsStaged() != tt.staged {
			t.Errorf("IsStaged() = %v, want %v for %+v", tt.fs.IsStaged(), tt.staged, tt.fs)
		}
	}
}
