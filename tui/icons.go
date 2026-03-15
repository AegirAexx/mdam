package tui

// Icons holds display glyphs used throughout the TUI.
// Two variants: Nerd Font (rich unicode glyphs) and plain ASCII fallback.
type Icons struct {
	// Document types
	Journal  string
	KB       string
	Todo     string
	Scratch  string
	Unsorted string

	// Git status
	GitModified  string
	GitUntracked string
	GitStaged    string

	// Navigation / state
	Pinned    string
	CursorSel string // selected item cursor
	CursorInactive string // inactive cursor (focused panel but not selected row)
	Dashboard string
	Tag       string
	Filter    string
	Loading   string
}

// DefaultIcons returns Nerd Font icon variants.
// Requires a Nerd Font installed in the terminal.
func DefaultIcons() Icons {
	return Icons{
		Journal:        " ",
		KB:             " ",
		Todo:           " ",
		Scratch:        " ",
		Unsorted:       " ",
		GitModified:    " ",
		GitUntracked:   " ",
		GitStaged:      " ",
		Pinned:         " ",
		CursorSel:      " ",
		CursorInactive: "─ ",
		Dashboard:      " ",
		Tag:            " ",
		Filter:         " ",
		Loading:        "⣾ ",
	}
}

// PlainIcons returns ASCII-safe icon variants that work on any terminal.
func PlainIcons() Icons {
	return Icons{
		Journal:        "[J] ",
		KB:             "[K] ",
		Todo:           "[T] ",
		Scratch:        "[S] ",
		Unsorted:       "[U] ",
		GitModified:    "[M]",
		GitUntracked:   "[?]",
		GitStaged:      "[A]",
		Pinned:         "[*]",
		CursorSel:      "> ",
		CursorInactive: "- ",
		Dashboard:      "[D] ",
		Tag:            "#",
		Filter:         "~",
		Loading:        "~",
	}
}
