// Package importer implements the import pipeline for ingesting external
// markdown files into the mdam managed tree.
package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AegirAexx/mdam/internal/document"
)

// ValidationError describes a validation failure for a single file.
type ValidationError struct {
	File    string
	Reason  string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.File, e.Reason)
}

// Result is the outcome of attempting to import a single file.
type Result struct {
	SourcePath string
	DestPath   string
	// Errors contains any validation issues found.
	Errors []ValidationError
	// Fixed is true if auto-fix was applied.
	Fixed bool
	// Skipped is true if the file was skipped (e.g. duplicate).
	Skipped bool
}

// Options controls import pipeline behaviour.
type Options struct {
	// AutoFix automatically renames files and scaffolds missing frontmatter.
	AutoFix bool
	// DryRun reports issues without modifying any files.
	DryRun bool
	// DestDir is the directory to move validated files into. If empty, files
	// stay in place.
	DestDir string
}

// ImportFile validates and optionally imports a single markdown file.
// path is the source file. baseDir is the managed tree root (for duplicate detection).
func ImportFile(path, baseDir string, opts Options) (Result, error) {
	result := Result{SourcePath: path}

	// Check file exists and is readable.
	info, err := os.Stat(path)
	if err != nil {
		return result, fmt.Errorf("stat %s: %w", path, err)
	}
	if info.IsDir() {
		return result, fmt.Errorf("%s is a directory", path)
	}

	// Validate and possibly fix filename.
	name := strings.TrimSuffix(filepath.Base(path), ".md")
	if err := document.ValidateFilename(name); err != nil {
		if !opts.AutoFix {
			result.Errors = append(result.Errors, ValidationError{
				File:   path,
				Reason: "invalid filename: " + err.Error(),
			})
		} else {
			fixed := document.ToKebabCase(name)
			newPath := filepath.Join(filepath.Dir(path), fixed+".md")
			if !opts.DryRun {
				if err := os.Rename(path, newPath); err != nil {
					return result, fmt.Errorf("renaming file: %w", err)
				}
				path = newPath
				result.Fixed = true
			}
			name = fixed
		}
	}

	// Duplicate detection.
	destName := name + ".md"
	if baseDir != "" {
		if duplicate, err := findDuplicate(baseDir, destName); err == nil && duplicate != "" {
			result.Errors = append(result.Errors, ValidationError{
				File:   path,
				Reason: fmt.Sprintf("duplicate: already exists at %s", duplicate),
			})
			result.Skipped = true
			return result, nil
		}
	}

	// Read and validate frontmatter.
	data, err := os.ReadFile(path)
	if err != nil {
		return result, fmt.Errorf("reading %s: %w", path, err)
	}

	fm, _, fmErr := tryParseFrontmatter(string(data))
	if fmErr != nil {
		if !opts.AutoFix {
			result.Errors = append(result.Errors, ValidationError{
				File:   path,
				Reason: "invalid frontmatter: " + fmErr.Error(),
			})
		} else if !opts.DryRun {
			// Scaffold missing frontmatter from file mtime and name.
			scaffolded := scaffoldFrontmatter(name, info.ModTime())
			header, renderErr := document.RenderFrontmatter(scaffolded)
			if renderErr != nil {
				return result, fmt.Errorf("rendering scaffolded frontmatter: %w", renderErr)
			}
			body := stripExistingFrontmatter(string(data))
			newContent := header + "\n" + body
			if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
				return result, fmt.Errorf("writing scaffolded file: %w", err)
			}
			result.Fixed = true
			fm, _, _ = tryParseFrontmatter(newContent)
		}
	} else {
		if valErr := document.ValidateFrontmatter(fm); valErr != nil {
			if !opts.AutoFix {
				result.Errors = append(result.Errors, ValidationError{
					File:   path,
					Reason: "frontmatter validation: " + valErr.Error(),
				})
			} else if !opts.DryRun {
				fm = fixFrontmatter(fm, name, info.ModTime())
				header, _ := document.RenderFrontmatter(fm)
				_, body, _ := tryParseFrontmatter(string(data))
				if err := os.WriteFile(path, []byte(header+"\n"+body), 0o644); err != nil {
					return result, fmt.Errorf("writing fixed frontmatter: %w", err)
				}
				result.Fixed = true
			}
		}
	}

	// Move to dest directory if specified and no blocking errors.
	if opts.DestDir != "" && len(result.Errors) == 0 && !opts.DryRun {
		dest := filepath.Join(opts.DestDir, destName)
		if err := os.MkdirAll(opts.DestDir, 0o755); err != nil {
			return result, fmt.Errorf("creating dest dir: %w", err)
		}
		if err := os.Rename(path, dest); err != nil {
			return result, fmt.Errorf("moving file to %s: %w", dest, err)
		}
		result.DestPath = dest
	} else {
		result.DestPath = path
	}

	return result, nil
}

// ImportDir imports all .md files found in dir (non-recursive by default).
func ImportDir(dir, baseDir string, opts Options) ([]Result, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}
	var results []Result
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		r, err := ImportFile(path, baseDir, opts)
		if err != nil {
			return results, err
		}
		results = append(results, r)
	}
	return results, nil
}

// tryParseFrontmatter is a thin wrapper that returns the frontmatter without
// panicking on parse errors.
func tryParseFrontmatter(content string) (document.Frontmatter, string, error) {
	if !strings.HasPrefix(content, "---") {
		return document.Frontmatter{}, content, fmt.Errorf("no frontmatter")
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
		return document.Frontmatter{}, content, fmt.Errorf("no closing delimiter")
	}
	// Delegate to document package via a temp file approach is too heavy;
	// replicate the logic inline for now.
	yamlContent := strings.Join(lines[1:closing], "\n")
	body := strings.TrimLeft(strings.Join(lines[closing+1:], "\n"), "\r\n")

	// Parse using yaml.
	var fm document.Frontmatter
	unmarshalFrontmatter(&fm, yamlContent)
	return fm, body, nil
}

// unmarshalFrontmatter parses yamlContent into out using yaml.v3.
func unmarshalFrontmatter(out interface{}, yamlContent string) {
	_ = yamlUnmarshal([]byte(yamlContent), out)
}

// scaffoldFrontmatter creates a best-effort Frontmatter from filename and mtime.
func scaffoldFrontmatter(name string, mtime time.Time) document.Frontmatter {
	title := strings.ReplaceAll(name, "-", " ")
	// Capitalise first letter.
	if len(title) > 0 {
		title = strings.ToUpper(title[:1]) + title[1:]
	}
	return document.Frontmatter{
		Title:    title,
		Tags:     []string{},
		Created:  mtime.UTC(),
		Modified: mtime.UTC(),
		Type:     "unsorted",
	}
}

// fixFrontmatter fills in missing required fields with safe defaults.
func fixFrontmatter(fm document.Frontmatter, name string, mtime time.Time) document.Frontmatter {
	if strings.TrimSpace(fm.Title) == "" {
		fm.Title = strings.ReplaceAll(name, "-", " ")
	}
	if fm.Tags == nil {
		fm.Tags = []string{}
	}
	if fm.Created.IsZero() {
		fm.Created = mtime.UTC()
	}
	if fm.Modified.IsZero() {
		fm.Modified = mtime.UTC()
	}
	if strings.TrimSpace(fm.Type) == "" || !document.ValidTypes[fm.Type] {
		fm.Type = "unsorted"
	}
	return fm
}

// stripExistingFrontmatter returns content with any existing frontmatter block removed.
func stripExistingFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	lines := strings.Split(content, "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			return strings.TrimLeft(strings.Join(lines[i+1:], "\n"), "\r\n")
		}
	}
	return content
}

// findDuplicate walks baseDir looking for any file named destName.
// Returns the first match found, or "" if none.
func findDuplicate(baseDir, destName string) (string, error) {
	var found string
	err := filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && d.Name() == destName {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found, err
}
