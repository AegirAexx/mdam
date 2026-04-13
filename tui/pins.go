package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// maxPins is the maximum number of pinned documents.
const maxPins = 10

// loadPins reads pinned document paths from pinsPath.
// Returns an empty slice (not an error) if the file does not exist.
// The slice order is the insertion order (oldest first).
// Stale entries (files that no longer exist on disk) are pruned automatically.
// Stored paths are relative to baseDir; absolute paths are accepted for
// backward compatibility with existing pins.json files.
func loadPins(pinsPath, baseDir string) ([]string, error) {
	data, err := os.ReadFile(pinsPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading pins: %w", err)
	}
	var raw []string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing pins: %w", err)
	}
	// Prune stale entries whose files no longer exist.
	live := raw[:0]
	for _, stored := range raw {
		var abs string
		if filepath.IsAbs(stored) {
			abs = stored // backward compat: old absolute path
		} else {
			abs = filepath.Join(baseDir, stored)
		}
		if _, err := os.Stat(abs); err == nil {
			live = append(live, abs)
		}
	}
	if len(live) != len(raw) {
		// Persist the pruned list so stale entries don't accumulate.
		_ = savePins(pinsPath, baseDir, live)
	}
	return live, nil
}

// savePins writes the pinned paths to pinsPath as a JSON array preserving order.
// Paths are stored relative to baseDir so that pins.json is portable across machines.
// Paths outside baseDir are stored as-is (absolute) as a fallback.
func savePins(pinsPath, baseDir string, pins []string) error {
	rel := make([]string, len(pins))
	for i, abs := range pins {
		r, err := filepath.Rel(baseDir, abs)
		if err != nil || filepath.IsAbs(r) {
			r = abs // outside baseDir: store absolute as fallback
		}
		rel[i] = r
	}
	data, err := json.Marshal(rel)
	if err != nil {
		return fmt.Errorf("marshaling pins: %w", err)
	}
	return os.WriteFile(pinsPath, data, 0o644)
}

// pinsToMap builds an O(1) lookup map from the ordered pin list.
func pinsToMap(pins []string) map[string]bool {
	m := make(map[string]bool, len(pins))
	for _, p := range pins {
		m[p] = true
	}
	return m
}

// togglePin returns a new ordered pin list with path added (at end) if absent,
// or removed if present. If adding would exceed maxPins, the oldest pin is evicted.
func togglePin(pins []string, path string) []string {
	// Check if already pinned — if so, remove.
	for i, p := range pins {
		if p == path {
			return append(pins[:i], pins[i+1:]...)
		}
	}
	// Add new pin.
	next := append(pins, path)
	// Evict oldest if over limit.
	if len(next) > maxPins {
		next = next[len(next)-maxPins:]
	}
	return next
}
