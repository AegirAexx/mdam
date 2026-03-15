package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// loadPins reads pinned document paths from path.
// Returns an empty map (not an error) if the file does not exist.
func loadPins(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[string]bool), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading pins: %w", err)
	}
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return nil, fmt.Errorf("parsing pins: %w", err)
	}
	m := make(map[string]bool, len(paths))
	for _, p := range paths {
		m[p] = true
	}
	return m, nil
}

// savePins writes the pinned paths map to path as a sorted JSON array.
func savePins(pinsPath string, pins map[string]bool) error {
	paths := make([]string, 0, len(pins))
	for p := range pins {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	data, err := json.Marshal(paths)
	if err != nil {
		return fmt.Errorf("marshaling pins: %w", err)
	}
	return os.WriteFile(pinsPath, data, 0o644)
}

// togglePin returns a new pins map with path added if absent, or removed if present.
// The original map is not mutated.
func togglePin(pins map[string]bool, path string) map[string]bool {
	next := make(map[string]bool, len(pins))
	for k, v := range pins {
		next[k] = v
	}
	if next[path] {
		delete(next, path)
	} else {
		next[path] = true
	}
	return next
}
