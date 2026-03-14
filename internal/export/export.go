// Package export strips frontmatter from managed documents for sharing.
package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Strip removes the YAML frontmatter block from content and returns the
// trimmed markdown body. Returns an error if the file has no frontmatter.
func Strip(content string) (string, error) {
	if !strings.HasPrefix(content, "---") {
		return "", fmt.Errorf("content has no frontmatter delimiter")
	}

	lines := strings.Split(content, "\n")
	closing := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			closing = i
			break
		}
	}
	if closing == -1 {
		return "", fmt.Errorf("content has no closing frontmatter delimiter")
	}

	body := strings.Join(lines[closing+1:], "\n")
	body = strings.TrimLeft(body, "\r\n")
	return body, nil
}

// ToFile reads srcPath, strips its frontmatter, and writes the result to
// destDir/<same filename>. The destination directory is created if needed.
// Returns the destination path.
func ToFile(srcPath, destDir string) (string, error) {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("reading source file: %w", err)
	}

	body, err := Strip(string(data))
	if err != nil {
		return "", fmt.Errorf("stripping frontmatter from %s: %w", srcPath, err)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("creating export directory: %w", err)
	}

	destPath := filepath.Join(destDir, filepath.Base(srcPath))
	if err := os.WriteFile(destPath, []byte(body), 0o644); err != nil {
		return "", fmt.Errorf("writing exported file: %w", err)
	}
	return destPath, nil
}
