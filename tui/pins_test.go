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

	pins := map[string]bool{
		"/notes/a.md": true,
		"/notes/b.md": true,
	}
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
	for p := range pins {
		if !loaded[p] {
			t.Errorf("pin %q not found after round-trip", p)
		}
	}
}

func TestTogglePinAddsPath(t *testing.T) {
	pins := map[string]bool{}
	result := togglePin(pins, "/notes/a.md")
	if !result["/notes/a.md"] {
		t.Error("togglePin should add path when absent")
	}
}

func TestTogglePinRemovesPath(t *testing.T) {
	pins := map[string]bool{"/notes/a.md": true}
	result := togglePin(pins, "/notes/a.md")
	if result["/notes/a.md"] {
		t.Error("togglePin should remove path when present")
	}
}

func TestTogglePinDoesNotMutateOriginal(t *testing.T) {
	pins := map[string]bool{"/notes/a.md": true}
	_ = togglePin(pins, "/notes/a.md")
	if !pins["/notes/a.md"] {
		t.Error("togglePin should not mutate the original map")
	}
}
