package tui

import "testing"

// Smoke tests: NewTheme must not panic and must produce valid styles.

func TestNewThemeTokyoNight(t *testing.T) {
	th := NewTheme("tokyonight")
	if th.GlamourStyle == "" {
		t.Error("tokyonight: GlamourStyle is empty")
	}
	// Render must produce non-empty output.
	if th.StatusNormal.Render("NORMAL") == "" {
		t.Error("tokyonight: StatusNormal.Render returned empty")
	}
}

func TestNewThemeNord(t *testing.T) {
	th := NewTheme("nord")
	if th.GlamourStyle == "" {
		t.Error("nord: GlamourStyle is empty")
	}
	if th.HeaderFocused.Render("Preview") == "" {
		t.Error("nord: HeaderFocused.Render returned empty")
	}
}

func TestNewThemeGruvbox(t *testing.T) {
	th := NewTheme("gruvbox")
	if th.GlamourStyle != "dark" {
		t.Errorf("gruvbox: GlamourStyle = %q, want dark", th.GlamourStyle)
	}
}

func TestNewThemeCatppuccin(t *testing.T) {
	th := NewTheme("catppuccin")
	if th.GlamourStyle == "" {
		t.Error("catppuccin: GlamourStyle is empty")
	}
}

func TestNewThemeDracula(t *testing.T) {
	th := NewTheme("dracula")
	if th.GlamourStyle == "" {
		t.Error("dracula: GlamourStyle is empty")
	}
}

func TestNewThemeUnknownFallsBackToTokyoNight(t *testing.T) {
	th := NewTheme("nonexistent-theme")
	ref := NewTheme("tokyonight")
	if th.GlamourStyle != ref.GlamourStyle {
		t.Errorf("unknown theme glamour = %q, want %q", th.GlamourStyle, ref.GlamourStyle)
	}
}
