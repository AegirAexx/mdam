// Package tui implements the BubbleTea TUI for mdam.
package tui

// Mode represents the current input mode of the TUI.
type Mode int

const (
	// ModeNormal is the default navigation mode.
	ModeNormal Mode = iota
	// ModeCommand is activated by ":" — accepts colon-prefixed commands.
	ModeCommand
	// ModeSearch is activated by "/" — accepts fuzzy search queries.
	ModeSearch
	// ModeTemplatePicker is the template selection overlay for new documents.
	ModeTemplatePicker
	// ModeTemplateVars collects values for unresolved template variables.
	ModeTemplateVars
	// ModeDeleteConfirm prompts the user to confirm a document deletion.
	ModeDeleteConfirm
	// ModeRead is a full-screen glamour read overlay, triggered by "o".
	ModeRead
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeCommand:
		return "COMMAND"
	case ModeSearch:
		return "SEARCH"
	case ModeTemplatePicker, ModeTemplateVars:
		return "NEW DOC"
	case ModeDeleteConfirm:
		return "DELETE?"
	case ModeRead:
		return "READ"
	default:
		return "UNKNOWN"
	}
}

// View identifies which pane is currently active.
type View int

const (
	ViewDashboard View = iota // key 1, startup default
	ViewJournal               // key 2
	ViewKB                    // key 3
	ViewTags                  // key 4
)

// cycleView advances the active view by delta (1 or -1), wrapping around.
func cycleView(v View, delta int) View {
	const n = 4 // total number of named views
	return View((int(v) + delta + n) % n)
}

// PanelID identifies which panel currently has focus.
type PanelID int

const (
	PanelFiles   PanelID = iota
	PanelPreview
	panelCount // sentinel — total number of panels
)

func (p PanelID) String() string {
	switch p {
	case PanelFiles:
		return "Files"
	case PanelPreview:
		return "Preview"
	default:
		return "?"
	}
}

// next returns the next panel in cycle order.
func (p PanelID) next() PanelID {
	return (p + 1) % panelCount
}

// prev returns the previous panel in cycle order.
func (p PanelID) prev() PanelID {
	return (p + panelCount - 1) % panelCount
}
