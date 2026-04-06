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
	Pinned         string
	Missing        string // missing file indicator
	CursorSel      string // selected item cursor
	CursorInactive string // inactive cursor (focused panel but not selected row)
	Dashboard      string
	Tag            string
	Filter         string
	Loading        string
}

// DefaultIcons returns Nerd Font icon variants.
// Requires a Nerd Font installed in the terminal.
// Codepoints use \U escape sequences to survive any editor/encoding.
func DefaultIcons() Icons {
	return Icons{
		Journal:        "\U000F0068 ", // nf-md-calendar
		KB:             "\U000F02DA ", // nf-md-library
		Todo:           "\U000F0134 ", // nf-md-checkbox_marked
		Scratch:        "\U000F0219 ", // nf-md-file_document_edit
		Unsorted:       "\U000F0224 ", // nf-md-file_question
		GitModified:    "\U000F0415",  // nf-md-pencil
		GitUntracked:   "\U000F0547",  // nf-md-sticker_plus_outline → nf-md-plus_circle
		GitStaged:      "\U000F012C",  // nf-md-check
		Pinned:         "\U000F0403",  // nf-md-pin
		Missing:        "\U000F0159",  // nf-md-close_circle
		CursorSel:      "\U000F0142 ", // nf-md-chevron_right
		CursorInactive: "─ ",
		Dashboard:      "\U000F00EC ", // nf-md-view_dashboard
		Tag:            "\U000F04F9 ", // nf-md-tag
		Filter:         "\U000F0232 ", // nf-md-filter
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
		Missing:        "X",
		CursorSel:      "> ",
		CursorInactive: "- ",
		Dashboard:      "[D] ",
		Tag:            "#",
		Filter:         "~",
		Loading:        "~",
	}
}
