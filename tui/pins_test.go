package tui

import (
	"path/filepath"
	"testing"
)

func TestLoadPinsMissingFileReturnsEmpty(t *testing.T) {
	pins, err := loadPins("/does/not/exist/pins.json")
	if err != nil {
		t.Fatalf("loadPins missing file: unexpected error: %v", err)
	}
	if len(pins) != 0 {
		t.Errorf("loadPins missing file: len = %d, want 0", len(pins))
	}
}

func TestSaveAndLoadPinsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pins.json")

	pins := []string{"/notes/a.md", "/notes/b.md"}
	if err := savePins(path, pins); err != nil {
		t.Fatalf("savePins: %v", err)
	}

	loaded, err := loadPins(path)
	if err != nil {
		t.Fatalf("loadPins: %v", err)
	}
	if len(loaded) != len(pins) {
		t.Errorf("loaded %d pins, want %d", len(loaded), len(pins))
	}
	for i, p := range pins {
		if loaded[i] != p {
			t.Errorf("loaded[%d] = %q, want %q", i, loaded[i], p)
		}
	}
}

func TestTogglePinAddsPath(t *testing.T) {
	var pins []string
	result := togglePin(pins, "/notes/a.md")
	if len(result) != 1 || result[0] != "/notes/a.md" {
		t.Errorf("togglePin add: got %v, want [/notes/a.md]", result)
	}
}

func TestTogglePinRemovesPath(t *testing.T) {
	pins := []string{"/notes/a.md"}
	result := togglePin(pins, "/notes/a.md")
	if len(result) != 0 {
		t.Errorf("togglePin remove: got %v, want empty", result)
	}
}

func TestTogglePinEvictsOldest(t *testing.T) {
	var pins []string
	for i := 0; i < maxPins; i++ {
		pins = append(pins, filepath.Join("/notes", string(rune('a'+i))+".md"))
	}
	// Adding 11th should evict the first.
	result := togglePin(pins, "/notes/new.md")
	if len(result) != maxPins {
		t.Fatalf("expected %d pins, got %d", maxPins, len(result))
	}
	if result[0] == "/notes/a.md" {
		t.Error("oldest pin should have been evicted")
	}
	if result[len(result)-1] != "/notes/new.md" {
		t.Error("new pin should be at the end")
	}
}

func TestTogglePinPreservesOrder(t *testing.T) {
	pins := []string{"/notes/a.md", "/notes/b.md", "/notes/c.md"}
	result := togglePin(pins, "/notes/d.md")
	expected := []string{"/notes/a.md", "/notes/b.md", "/notes/c.md", "/notes/d.md"}
	if len(result) != len(expected) {
		t.Fatalf("len = %d, want %d", len(result), len(expected))
	}
	for i, p := range expected {
		if result[i] != p {
			t.Errorf("result[%d] = %q, want %q", i, result[i], p)
		}
	}
}

func TestPinsToMap(t *testing.T) {
	pins := []string{"/notes/a.md", "/notes/b.md"}
	m := pinsToMap(pins)
	if !m["/notes/a.md"] || !m["/notes/b.md"] {
		t.Errorf("pinsToMap missing entries: %v", m)
	}
	if m["/notes/c.md"] {
		t.Error("pinsToMap should not contain absent path")
	}
}
