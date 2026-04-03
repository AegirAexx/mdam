package tui

import "github.com/charmbracelet/lipgloss"

// Theme holds all lipgloss styles for the TUI, computed once per palette.
// Never re-computed during a render cycle.
type Theme struct {
	// Status bar
	StatusNormal  lipgloss.Style
	StatusCommand lipgloss.Style
	StatusSearch  lipgloss.Style
	StatusNewDoc  lipgloss.Style
	StatusGit     lipgloss.Style
	StatusInfo    lipgloss.Style
	StatusMsg     lipgloss.Style
	StatusHint    lipgloss.Style

	// Panel headers
	HeaderFocused   lipgloss.Style
	HeaderUnfocused lipgloss.Style
	Divider         lipgloss.Style

	// File list
	FileSelected lipgloss.Style
	FileNormal   lipgloss.Style
	FilePinned   lipgloss.Style
	FileCursor   lipgloss.Style

	// Git markers
	GitModified  lipgloss.Style
	GitUntracked lipgloss.Style
	GitStaged    lipgloss.Style

	// Preview panel
	PreviewTitle lipgloss.Style
	PreviewMeta  lipgloss.Style
	PreviewKey   lipgloss.Style

	// TODO items
	TodoOpen   lipgloss.Style
	TodoDone   lipgloss.Style

	// FilterBar is reserved for future use.
	FilterBar lipgloss.Style

	// Dashboard
	DashboardHeader lipgloss.Style
	DashboardItem   lipgloss.Style

	// Tab bar
	TabActive   lipgloss.Style // active pane tab — inverted colors
	TabInactive lipgloss.Style // inactive pane tab — muted text

	// GlamourStyle is the name passed to glamour.Render.
	GlamourStyle string
}

// NewTheme returns a Theme for the given palette name.
// Unknown names fall back to tokyonight.
func NewTheme(name string) Theme {
	switch name {
	case "nord":
		return nordTheme()
	case "gruvbox":
		return gruvboxTheme()
	case "catppuccin":
		return catppuccinTheme()
	case "dracula":
		return draculaTheme()
	default:
		return tokyonightTheme()
	}
}

func tokyonightTheme() Theme {
	const (
		bg      = "#1a1b26"
		fg      = "#c0caf5"
		dim     = "#565f89"
		blue    = "#7aa2f7"
		cyan    = "#7dcfff"
		green   = "#9ece6a"
		orange  = "#ff9e64"
		yellow  = "#e0af68"
		red     = "#f7768e"
		purple  = "#9d7cd8"
		comment = "#444b6a"
	)
	return Theme{
		StatusNormal:    lipgloss.NewStyle().Background(lipgloss.Color(blue)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusCommand:   lipgloss.NewStyle().Background(lipgloss.Color(yellow)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusSearch:    lipgloss.NewStyle().Background(lipgloss.Color(cyan)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusNewDoc:    lipgloss.NewStyle().Background(lipgloss.Color(green)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusGit:       lipgloss.NewStyle().Foreground(lipgloss.Color(purple)),
		StatusInfo:      lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		StatusMsg:       lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		StatusHint:      lipgloss.NewStyle().Foreground(lipgloss.Color(comment)),
		HeaderFocused:   lipgloss.NewStyle().Foreground(lipgloss.Color(blue)).Bold(true),
		HeaderUnfocused: lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		Divider:         lipgloss.NewStyle().Foreground(lipgloss.Color(comment)),
		FileSelected:    lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		FileNormal:      lipgloss.NewStyle().Foreground(lipgloss.Color(fg)),
		FilePinned:      lipgloss.NewStyle().Foreground(lipgloss.Color(cyan)),
		FileCursor:      lipgloss.NewStyle().Foreground(lipgloss.Color(blue)).Bold(true),
		GitModified:     lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		GitUntracked:    lipgloss.NewStyle().Foreground(lipgloss.Color(red)),
		GitStaged:       lipgloss.NewStyle().Foreground(lipgloss.Color(green)),
		PreviewTitle:    lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		PreviewMeta:     lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		PreviewKey:      lipgloss.NewStyle().Foreground(lipgloss.Color(blue)),
		TodoOpen:        lipgloss.NewStyle().Foreground(lipgloss.Color(orange)),
		TodoDone:        lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		FilterBar:       lipgloss.NewStyle().Background(lipgloss.Color(comment)).Foreground(lipgloss.Color(cyan)).Padding(0, 1),
		DashboardHeader: lipgloss.NewStyle().Foreground(lipgloss.Color(blue)).Bold(true),
		DashboardItem:   lipgloss.NewStyle().Foreground(lipgloss.Color(fg)),
		TabActive:       lipgloss.NewStyle().Reverse(true).Bold(true).Padding(0, 1),
		TabInactive:     lipgloss.NewStyle().Foreground(lipgloss.Color(dim)).Padding(0, 1),
		GlamourStyle:    "dark",
	}
}

func nordTheme() Theme {
	const (
		bg     = "#2e3440"
		fg     = "#d8dee9"
		dim    = "#4c566a"
		blue   = "#81a1c1"
		cyan   = "#88c0d0"
		green  = "#a3be8c"
		yellow = "#ebcb8b"
		red    = "#bf616a"
		purple = "#b48ead"
	)
	return Theme{
		StatusNormal:    lipgloss.NewStyle().Background(lipgloss.Color(blue)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusCommand:   lipgloss.NewStyle().Background(lipgloss.Color(yellow)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusSearch:    lipgloss.NewStyle().Background(lipgloss.Color(cyan)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusNewDoc:    lipgloss.NewStyle().Background(lipgloss.Color(green)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusGit:       lipgloss.NewStyle().Foreground(lipgloss.Color(purple)),
		StatusInfo:      lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		StatusMsg:       lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		StatusHint:      lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		HeaderFocused:   lipgloss.NewStyle().Foreground(lipgloss.Color(cyan)).Bold(true),
		HeaderUnfocused: lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		Divider:         lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		FileSelected:    lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		FileNormal:      lipgloss.NewStyle().Foreground(lipgloss.Color(fg)),
		FilePinned:      lipgloss.NewStyle().Foreground(lipgloss.Color(cyan)),
		FileCursor:      lipgloss.NewStyle().Foreground(lipgloss.Color(blue)).Bold(true),
		GitModified:     lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		GitUntracked:    lipgloss.NewStyle().Foreground(lipgloss.Color(red)),
		GitStaged:       lipgloss.NewStyle().Foreground(lipgloss.Color(green)),
		PreviewTitle:    lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		PreviewMeta:     lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		PreviewKey:      lipgloss.NewStyle().Foreground(lipgloss.Color(blue)),
		TodoOpen:        lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		TodoDone:        lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		FilterBar:       lipgloss.NewStyle().Background(lipgloss.Color(dim)).Foreground(lipgloss.Color(cyan)).Padding(0, 1),
		DashboardHeader: lipgloss.NewStyle().Foreground(lipgloss.Color(blue)).Bold(true),
		DashboardItem:   lipgloss.NewStyle().Foreground(lipgloss.Color(fg)),
		TabActive:       lipgloss.NewStyle().Reverse(true).Bold(true).Padding(0, 1),
		TabInactive:     lipgloss.NewStyle().Foreground(lipgloss.Color(dim)).Padding(0, 1),
		GlamourStyle:    "notty",
	}
}

func gruvboxTheme() Theme {
	const (
		bg     = "#282828"
		fg     = "#ebdbb2"
		dim    = "#504945"
		blue   = "#83a598"
		cyan   = "#8ec07c"
		green  = "#b8bb26"
		yellow = "#fabd2f"
		red    = "#fb4934"
		orange = "#fe8019"
	)
	return Theme{
		StatusNormal:    lipgloss.NewStyle().Background(lipgloss.Color(blue)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusCommand:   lipgloss.NewStyle().Background(lipgloss.Color(yellow)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusSearch:    lipgloss.NewStyle().Background(lipgloss.Color(green)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusNewDoc:    lipgloss.NewStyle().Background(lipgloss.Color(cyan)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusGit:       lipgloss.NewStyle().Foreground(lipgloss.Color(orange)),
		StatusInfo:      lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		StatusMsg:       lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		StatusHint:      lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		HeaderFocused:   lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)).Bold(true),
		HeaderUnfocused: lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		Divider:         lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		FileSelected:    lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		FileNormal:      lipgloss.NewStyle().Foreground(lipgloss.Color(fg)),
		FilePinned:      lipgloss.NewStyle().Foreground(lipgloss.Color(cyan)),
		FileCursor:      lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)).Bold(true),
		GitModified:     lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		GitUntracked:    lipgloss.NewStyle().Foreground(lipgloss.Color(red)),
		GitStaged:       lipgloss.NewStyle().Foreground(lipgloss.Color(green)),
		PreviewTitle:    lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		PreviewMeta:     lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		PreviewKey:      lipgloss.NewStyle().Foreground(lipgloss.Color(blue)),
		TodoOpen:        lipgloss.NewStyle().Foreground(lipgloss.Color(orange)),
		TodoDone:        lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		FilterBar:       lipgloss.NewStyle().Background(lipgloss.Color(dim)).Foreground(lipgloss.Color(yellow)).Padding(0, 1),
		DashboardHeader: lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)).Bold(true),
		DashboardItem:   lipgloss.NewStyle().Foreground(lipgloss.Color(fg)),
		TabActive:       lipgloss.NewStyle().Reverse(true).Bold(true).Padding(0, 1),
		TabInactive:     lipgloss.NewStyle().Foreground(lipgloss.Color(dim)).Padding(0, 1),
		GlamourStyle:    "dark",
	}
}

func catppuccinTheme() Theme {
	const (
		bg     = "#1e1e2e"
		fg     = "#cdd6f4"
		dim    = "#585b70"
		blue   = "#89b4fa"
		cyan   = "#89dceb"
		green  = "#a6e3a1"
		yellow = "#f9e2af"
		red    = "#f38ba8"
		pink   = "#f5c2e7"
	)
	return Theme{
		StatusNormal:    lipgloss.NewStyle().Background(lipgloss.Color(blue)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusCommand:   lipgloss.NewStyle().Background(lipgloss.Color(yellow)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusSearch:    lipgloss.NewStyle().Background(lipgloss.Color(cyan)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusNewDoc:    lipgloss.NewStyle().Background(lipgloss.Color(green)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusGit:       lipgloss.NewStyle().Foreground(lipgloss.Color(pink)),
		StatusInfo:      lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		StatusMsg:       lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		StatusHint:      lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		HeaderFocused:   lipgloss.NewStyle().Foreground(lipgloss.Color(blue)).Bold(true),
		HeaderUnfocused: lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		Divider:         lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		FileSelected:    lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		FileNormal:      lipgloss.NewStyle().Foreground(lipgloss.Color(fg)),
		FilePinned:      lipgloss.NewStyle().Foreground(lipgloss.Color(cyan)),
		FileCursor:      lipgloss.NewStyle().Foreground(lipgloss.Color(blue)).Bold(true),
		GitModified:     lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		GitUntracked:    lipgloss.NewStyle().Foreground(lipgloss.Color(red)),
		GitStaged:       lipgloss.NewStyle().Foreground(lipgloss.Color(green)),
		PreviewTitle:    lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		PreviewMeta:     lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		PreviewKey:      lipgloss.NewStyle().Foreground(lipgloss.Color(blue)),
		TodoOpen:        lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		TodoDone:        lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		FilterBar:       lipgloss.NewStyle().Background(lipgloss.Color(dim)).Foreground(lipgloss.Color(cyan)).Padding(0, 1),
		DashboardHeader: lipgloss.NewStyle().Foreground(lipgloss.Color(blue)).Bold(true),
		DashboardItem:   lipgloss.NewStyle().Foreground(lipgloss.Color(fg)),
		TabActive:       lipgloss.NewStyle().Reverse(true).Bold(true).Padding(0, 1),
		TabInactive:     lipgloss.NewStyle().Foreground(lipgloss.Color(dim)).Padding(0, 1),
		GlamourStyle:    "dark",
	}
}

func draculaTheme() Theme {
	const (
		bg     = "#282a36"
		fg     = "#f8f8f2"
		dim    = "#44475a"
		blue   = "#6272a4"
		cyan   = "#8be9fd"
		green  = "#50fa7b"
		yellow = "#f1fa8c"
		red    = "#ff5555"
		pink   = "#ff79c6"
		purple = "#bd93f9"
	)
	return Theme{
		StatusNormal:    lipgloss.NewStyle().Background(lipgloss.Color(purple)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusCommand:   lipgloss.NewStyle().Background(lipgloss.Color(yellow)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusSearch:    lipgloss.NewStyle().Background(lipgloss.Color(cyan)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusNewDoc:    lipgloss.NewStyle().Background(lipgloss.Color(green)).Foreground(lipgloss.Color(bg)).Bold(true).Padding(0, 1),
		StatusGit:       lipgloss.NewStyle().Foreground(lipgloss.Color(pink)),
		StatusInfo:      lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		StatusMsg:       lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		StatusHint:      lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		HeaderFocused:   lipgloss.NewStyle().Foreground(lipgloss.Color(purple)).Bold(true),
		HeaderUnfocused: lipgloss.NewStyle().Foreground(lipgloss.Color(blue)),
		Divider:         lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		FileSelected:    lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		FileNormal:      lipgloss.NewStyle().Foreground(lipgloss.Color(fg)),
		FilePinned:      lipgloss.NewStyle().Foreground(lipgloss.Color(cyan)),
		FileCursor:      lipgloss.NewStyle().Foreground(lipgloss.Color(purple)).Bold(true),
		GitModified:     lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		GitUntracked:    lipgloss.NewStyle().Foreground(lipgloss.Color(red)),
		GitStaged:       lipgloss.NewStyle().Foreground(lipgloss.Color(green)),
		PreviewTitle:    lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		PreviewMeta:     lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		PreviewKey:      lipgloss.NewStyle().Foreground(lipgloss.Color(purple)),
		TodoOpen:        lipgloss.NewStyle().Foreground(lipgloss.Color(pink)),
		TodoDone:        lipgloss.NewStyle().Foreground(lipgloss.Color(dim)),
		FilterBar:       lipgloss.NewStyle().Background(lipgloss.Color(dim)).Foreground(lipgloss.Color(cyan)).Padding(0, 1),
		DashboardHeader: lipgloss.NewStyle().Foreground(lipgloss.Color(purple)).Bold(true),
		DashboardItem:   lipgloss.NewStyle().Foreground(lipgloss.Color(fg)),
		TabActive:       lipgloss.NewStyle().Reverse(true).Bold(true).Padding(0, 1),
		TabInactive:     lipgloss.NewStyle().Foreground(lipgloss.Color(dim)).Padding(0, 1),
		GlamourStyle:    "dark",
	}
}
