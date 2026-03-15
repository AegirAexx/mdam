package tui

import (
	"reflect"
	"testing"
)

func TestDefaultIconsNonEmpty(t *testing.T) {
	icons := DefaultIcons()
	v := reflect.ValueOf(icons)
	for i := range v.NumField() {
		field := v.Type().Field(i)
		val := v.Field(i).String()
		if val == "" {
			t.Errorf("DefaultIcons.%s is empty", field.Name)
		}
	}
}

func TestPlainIconsNonEmpty(t *testing.T) {
	icons := PlainIcons()
	v := reflect.ValueOf(icons)
	for i := range v.NumField() {
		field := v.Type().Field(i)
		val := v.Field(i).String()
		if val == "" {
			t.Errorf("PlainIcons.%s is empty", field.Name)
		}
	}
}

func TestDefaultAndPlainIconsDiffer(t *testing.T) {
	def := DefaultIcons()
	plain := PlainIcons()
	// At least the cursor should differ (Nerd Font vs plain).
	if def.CursorSel == plain.CursorSel {
		t.Error("DefaultIcons and PlainIcons should have different CursorSel")
	}
}
