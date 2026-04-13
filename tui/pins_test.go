package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPinsMissingFileReturnsEmpty(t *testing.T) {
	pins, err := loadPins("/does/not/exist/pins.json", "/base")
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

	// Create real files so the auto-prune doesn't remove them.
	aPath := filepath.Join(dir, "a.md")
	bPath := filepath.Join(dir, "b.md")
	for _, f := range []string{aPath, bPath} {
		if err := os.WriteFile(f, []byte("test"), 0o644); err != nil {
			t.Fatalf("creating test file: %v", err)
		}
	}

	pins := []string{aPath, bPath}
	if err := savePins(path, dir, pins); err != nil {
		t.Fatalf("savePins: %v", err)
	}

	loaded, err := loadPins(path, dir)
	if err != nil {
		t.Fatalf("loadPins: %v", err)
	}
	if len(loaded) != len(pins) {
		t.Fatalf("loaded %d pins, want %d", len(loaded), len(pins))
	}
	for i, p := range pins {
		if loaded[i] != p {
			t.Errorf("loaded[%d] = %q, want %q", i, loaded[i], p)
		}
	}
}

func TestSaveAndLoadPinsRelativePaths(t *testing.T) {
	dir := t.TempDir()
	pinsPath := filepath.Join(dir, "pins.json")

	// Create real files under the base dir.
	aPath := filepath.Join(dir, "journal", "a.md")
	if err := os.MkdirAll(filepath.Dir(aPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(aPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	if err := savePins(pinsPath, dir, []string{aPath}); err != nil {
		t.Fatalf("savePins: %v", err)
	}

	// Verify the on-disk JSON contains a relative path, not an absolute one.
	raw, err := os.ReadFile(pinsPath)
	if err != nil {
		t.Fatalf("reading pins.json: %v", err)
	}
	if filepath.IsAbs(string(raw[1 : len(raw)-2])) { // strip surrounding ["..."]
		t.Errorf("pins.json should store relative path, got: %s", raw)
	}

	// Confirm load reconstructs the original absolute path.
	loaded, err := loadPins(pinsPath, dir)
	if err != nil {
		t.Fatalf("loadPins: %v", err)
	}
	if len(loaded) != 1 || loaded[0] != aPath {
		t.Errorf("round-trip: got %v, want [%s]", loaded, aPath)
	}
}

func TestLoadPinsAbsolutePathsBackwardCompat(t *testing.T) {
	dir := t.TempDir()
	pinsPath := filepath.Join(dir, "pins.json")

	// Simulate a pre-migration pins.json with absolute paths.
	absPath := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(absPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}
	// Write absolute path directly to JSON (old format).
	jsonData := `["` + absPath + `"]`
	if err := os.WriteFile(pinsPath, []byte(jsonData), 0o644); err != nil {
		t.Fatalf("writing pins.json: %v", err)
	}

	// Load should return the absolute path unchanged.
	loaded, err := loadPins(pinsPath, "/some/other/base")
	if err != nil {
		t.Fatalf("loadPins: %v", err)
	}
	if len(loaded) != 1 || loaded[0] != absPath {
		t.Errorf("backward compat: got %v, want [%s]", loaded, absPath)
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

func TestLoadPinsPrunesStaleEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pins.json")

	// One real file, one that doesn't exist.
	realPath := filepath.Join(dir, "real.md")
	if err := os.WriteFile(realPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}
	stalePath := filepath.Join(dir, "gone.md")

	if err := savePins(path, dir, []string{realPath, stalePath}); err != nil {
		t.Fatalf("savePins: %v", err)
	}

	loaded, err := loadPins(path, dir)
	if err != nil {
		t.Fatalf("loadPins: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 pin after prune, got %d", len(loaded))
	}
	if loaded[0] != realPath {
		t.Errorf("expected %q, got %q", realPath, loaded[0])
	}

	// Verify pruned list was persisted.
	reloaded, err := loadPins(path, dir)
	if err != nil {
		t.Fatalf("loadPins after prune: %v", err)
	}
	if len(reloaded) != 1 {
		t.Errorf("persisted pins should have 1 entry, got %d", len(reloaded))
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
