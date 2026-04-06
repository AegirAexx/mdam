package tui

import (
	"encoding/json"
	"fmt"
	"os"
)

// maxPins is the maximum number of pinned documents.
const maxPins = 10

// loadPins reads pinned document paths from path.
// Returns an empty slice (not an error) if the file does not exist.
// The slice order is the insertion order (oldest first).
func loadPins(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading pins: %w", err)
	}
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return nil, fmt.Errorf("parsing pins: %w", err)
	}
	return paths, nil
}

// savePins writes the pinned paths to path as a JSON array preserving order.
func savePins(pinsPath string, pins []string) error {
	data, err := json.Marshal(pins)
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
