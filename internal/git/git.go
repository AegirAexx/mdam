// Package git provides git status detection for the mdam managed tree.
// It shells out to the git binary — no git library dependency.
package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FileStatus represents the git status of a single file.
type FileStatus struct {
	// Path is the file path relative to the repo root.
	Path string
	// X is the index status character (porcelain format).
	X rune
	// Y is the work-tree status character (porcelain format).
	Y rune
}

// IsModified reports whether the file has working-tree changes.
func (f FileStatus) IsModified() bool { return f.Y == 'M' }

// IsUntracked reports whether the file is untracked.
func (f FileStatus) IsUntracked() bool { return f.X == '?' && f.Y == '?' }

// IsStaged reports whether the file has staged changes.
func (f FileStatus) IsStaged() bool { return f.X != ' ' && f.X != '?' }

// RepoStatus holds overall repository status information.
type RepoStatus struct {
	Branch     string
	Ahead      int
	Behind     int
	Files      []FileStatus
	StashCount int
}

// UncommittedCount returns the number of files with any change.
func (r RepoStatus) UncommittedCount() int { return len(r.Files) }

// Status runs git status and returns a RepoStatus for the given directory.
// Returns an error if the directory is not a git repository or git is not found.
func Status(dir string) (RepoStatus, error) {
	branch, err := currentBranch(dir)
	if err != nil {
		return RepoStatus{}, fmt.Errorf("getting branch: %w", err)
	}

	files, err := porcelain(dir)
	if err != nil {
		return RepoStatus{}, fmt.Errorf("running git status: %w", err)
	}

	ahead, behind, err := aheadBehind(dir)
	if err != nil {
		// Remote may not exist — treat as 0/0.
		ahead, behind = 0, 0
	}

	stash, err := stashCount(dir)
	if err != nil {
		stash = 0
	}

	return RepoStatus{
		Branch:     branch,
		Ahead:      ahead,
		Behind:     behind,
		Files:      files,
		StashCount: stash,
	}, nil
}

// currentBranch returns the current branch name.
func currentBranch(dir string) (string, error) {
	out, err := run(dir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// porcelain parses `git status --porcelain` output into FileStatus slices.
func porcelain(dir string) ([]FileStatus, error) {
	out, err := run(dir, "git", "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	var files []FileStatus
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 4 {
			continue
		}
		files = append(files, FileStatus{
			X:    rune(line[0]),
			Y:    rune(line[1]),
			Path: strings.TrimSpace(line[3:]),
		})
	}
	return files, nil
}

// aheadBehind returns commits ahead and behind the upstream branch.
func aheadBehind(dir string) (ahead, behind int, err error) {
	out, err := run(dir, "git", "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output: %q", out)
	}
	ahead, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parsing ahead count: %w", err)
	}
	behind, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parsing behind count: %w", err)
	}
	return ahead, behind, nil
}

// stashCount returns the number of stash entries.
func stashCount(dir string) (int, error) {
	out, err := run(dir, "git", "stash", "list")
	if err != nil {
		return 0, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return 0, nil
	}
	return len(strings.Split(out, "\n")), nil
}

// run executes a command in dir and returns combined stdout output.
func run(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("running %s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(out), nil
}

// IsAvailable reports whether the git binary is available on PATH.
func IsAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsRepo reports whether dir is inside a git repository.
func IsRepo(dir string) bool {
	_, err := run(dir, "git", "rev-parse", "--git-dir")
	return err == nil
}
