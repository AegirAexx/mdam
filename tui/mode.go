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
	default:
		return "UNKNOWN"
	}
}

// View identifies which document set is currently displayed in the files panel.
type View int

const (
	ViewAll     View = iota // all documents (default)
	ViewJournal             // journal entries only
	ViewKB                  // knowledge base only
	ViewTodo                // focus TODO panel
	ViewRecent              // top 20 by modified date
)

// PanelID identifies which panel currently has focus.
type PanelID int

const (
	PanelFiles PanelID = iota
	PanelPreview
	PanelTodo
	panelCount // sentinel — total number of panels
)

func (p PanelID) String() string {
	switch p {
	case PanelFiles:
		return "Files"
	case PanelPreview:
		return "Preview"
	case PanelTodo:
		return "TODOs"
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
