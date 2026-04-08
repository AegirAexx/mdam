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
	got := IsAvailable()
	if !got {
		t.Skip("git not on PATH, skipping")
	}
	if !got {
		t.Error("IsAvailable() = false, want true")
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

func TestPorcelainStagedFile(t *testing.T) {
	if !IsAvailable() {
		t.Skip("git not available")
	}
	dir := initRepo(t)

	// Create and commit an initial file.
	initial := filepath.Join(dir, "initial.md")
	if err := os.WriteFile(initial, []byte("# Init\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "initial.md"},
		{"git", "commit", "-m", "init"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s: %v", args, string(out), err)
		}
	}

	// Stage a new file.
	if err := os.WriteFile(filepath.Join(dir, "new.md"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "new.md")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %s: %v", string(out), err)
	}

	files, err := porcelain(dir)
	if err != nil {
		t.Fatalf("porcelain() error = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("porcelain() returned %d files, want 1", len(files))
	}
	if files[0].X != 'A' {
		t.Errorf("X = %c, want A", files[0].X)
	}
	if !files[0].IsStaged() {
		t.Error("IsStaged() = false, want true")
	}
}

func TestStashCount(t *testing.T) {
	if !IsAvailable() {
		t.Skip("git not available")
	}
	dir := initRepo(t)

	// Create and commit an initial file.
	initial := filepath.Join(dir, "initial.md")
	if err := os.WriteFile(initial, []byte("# Init\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "initial.md"},
		{"git", "commit", "-m", "init"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s: %v", args, string(out), err)
		}
	}

	// Zero stashes before stashing.
	n, err := stashCount(dir)
	if err != nil {
		t.Fatalf("stashCount() error = %v", err)
	}
	if n != 0 {
		t.Errorf("stashCount() = %d, want 0", n)
	}

	// Modify file and stash.
	if err := os.WriteFile(initial, []byte("# Modified\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stashCmd := exec.Command("git", "stash", "push", "-m", "test stash")
	stashCmd.Dir = dir
	if out, err := stashCmd.CombinedOutput(); err != nil {
		t.Fatalf("git stash push: %s: %v", string(out), err)
	}

	n, err = stashCount(dir)
	if err != nil {
		t.Fatalf("stashCount() after stash error = %v", err)
	}
	if n != 1 {
		t.Errorf("stashCount() = %d, want 1", n)
	}

	// Pop the stash.
	popCmd := exec.Command("git", "stash", "pop")
	popCmd.Dir = dir
	if out, err := popCmd.CombinedOutput(); err != nil {
		t.Fatalf("git stash pop: %s: %v", string(out), err)
	}

	n2, err := stashCount(dir)
	if err != nil {
		t.Fatalf("stashCount() after pop error = %v", err)
	}
	if n2 != 0 {
		t.Errorf("stashCount() after pop = %d, want 0", n2)
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
