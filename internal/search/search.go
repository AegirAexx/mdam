// Package search provides fuzzy search across the mdam managed document tree.
package search

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/AegirAexx/mdam/internal/document"
)

// Result is a single search result.
type Result struct {
	Path        string
	Frontmatter document.Frontmatter
	Score       int
	// Snippet is a short excerpt from the body (empty if body was not searched).
	Snippet string
}

// Filters controls which documents are searched.
type Filters struct {
	Tag           string
	Type          string
	ModifiedAfter time.Time
}

// Search performs a fuzzy search over the managed document tree rooted at baseDir.
// query is matched against frontmatter fields and filenames. Full-text body
// search is not performed by default (see SearchWithBody).
// Results are sorted by score descending.
func Search(baseDir, query string, filters Filters) ([]Result, error) {
	return search(baseDir, query, filters, false)
}

// SearchWithBody includes document body content in the search, which is slower.
func SearchWithBody(baseDir, query string, filters Filters) ([]Result, error) {
	return search(baseDir, query, filters, true)
}

func search(baseDir, query string, filters Filters, includeBody bool) ([]Result, error) {
	var results []Result

	err := filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			// Skip hidden directories.
			if strings.HasPrefix(d.Name(), ".") && path != baseDir {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		doc, err := document.ParseFile(path)
		if err != nil {
			return nil // skip unparseable files
		}

		// Apply filters before scoring.
		if !matchFilters(doc, filters) {
			return nil
		}

		score, snippet := scoreDocument(doc, query, includeBody)
		if query == "" || score > 0 {
			results = append(results, Result{
				Path:        path,
				Frontmatter: doc.Frontmatter,
				Score:       score,
				Snippet:     snippet,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		// Tie-break: more recently modified first.
		return results[i].Frontmatter.Modified.After(results[j].Frontmatter.Modified)
	})

	return results, nil
}

// matchFilters returns true if the document passes all active filters.
func matchFilters(doc document.Document, f Filters) bool {
	if f.Tag != "" {
		found := false
		for _, tag := range doc.Frontmatter.Tags {
			if strings.EqualFold(tag, f.Tag) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if f.Type != "" && !strings.EqualFold(doc.Frontmatter.Type, f.Type) {
		return false
	}
	if !f.ModifiedAfter.IsZero() && doc.Frontmatter.Modified.Before(f.ModifiedAfter) {
		return false
	}
	return true
}

// scoreDocument computes a relevance score for doc against query.
// Higher scores rank first. Returns the score and an optional snippet.
//
// Scoring tiers:
//   - Exact tag match:    100
//   - Fuzzy title match:   50
//   - Filename match:      30
//   - Body match:          10
func scoreDocument(doc document.Document, query string, includeBody bool) (int, string) {
	if query == "" {
		return 0, ""
	}
	q := strings.ToLower(query)
	score := 0
	snippet := ""

	// Exact tag match (highest priority).
	for _, tag := range doc.Frontmatter.Tags {
		if strings.EqualFold(tag, query) {
			score += 100
		} else if fuzzyContains(strings.ToLower(tag), q) {
			score += 40
		}
	}

	// Title match.
	title := strings.ToLower(doc.Frontmatter.Title)
	if strings.Contains(title, q) {
		score += 50
	} else if fuzzyContains(title, q) {
		score += 25
	}

	// Filename match.
	filename := strings.ToLower(filepath.Base(doc.Path))
	if strings.Contains(filename, q) {
		score += 30
	} else if fuzzyContains(filename, q) {
		score += 15
	}

	// Body match (slower path).
	if includeBody && doc.Body != "" {
		body := strings.ToLower(doc.Body)
		if idx := strings.Index(body, q); idx != -1 {
			score += 10
			snippet = extractSnippet(doc.Body, idx, 80)
		} else if fuzzyContains(body, q) {
			score += 5
		}
	}

	return score, snippet
}

// fuzzyContains reports whether all runes of sub appear in s in order.
func fuzzyContains(s, sub string) bool {
	if sub == "" {
		return true
	}
	runes := []rune(sub)
	si := 0
	for _, r := range s {
		if unicode.ToLower(r) == unicode.ToLower(runes[si]) {
			si++
			if si == len(runes) {
				return true
			}
		}
	}
	return false
}

// extractSnippet returns up to maxLen characters of s centred around pos.
func extractSnippet(s string, pos, maxLen int) string {
	start := pos - maxLen/4
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(s) {
		end = len(s)
	}
	snippet := strings.TrimSpace(s[start:end])
	if start > 0 {
		snippet = "…" + snippet
	}
	if end < len(s) {
		snippet = snippet + "…"
	}
	return snippet
}

// ListAll returns all managed documents in baseDir without filtering or ranking.
func ListAll(baseDir string) ([]Result, error) {
	return search(baseDir, "", Filters{}, false)
}
