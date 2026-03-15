package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the TUI.
// Keybindings will be finalised at the Phase 3→4 transition (see KEYBINDINGS.md).
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Top      key.Binding // gg
	Bottom   key.Binding // G
	NextPane key.Binding // Tab
	PrevPane key.Binding // Shift+Tab

	// Modes
	Search  key.Binding // /
	Command key.Binding // :
	Help    key.Binding // ?
	Confirm key.Binding // Enter
	Cancel  key.Binding // Esc

	// Views
	ViewDashboard key.Binding // 1
	ViewJournal   key.Binding // 2
	ViewKB        key.Binding // 3
	ViewTodo      key.Binding // 4
	ViewRecent    key.Binding // 5
	ViewTags      key.Binding // 6

	// Actions
	Quit          key.Binding // q
	Rescan        key.Binding // R
	NewDoc        key.Binding // n
	Open          key.Binding // Enter (same as Confirm)
	Scratch       key.Binding // s
	Export        key.Binding // e
	Delete        key.Binding // d
	DeleteConfirm key.Binding // y (in ModeDeleteConfirm)
	Pin           key.Binding // p
	SmartFilter   key.Binding // f
	Lazygit       key.Binding // ctrl+g
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
		Down:     key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
		Left:     key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("h/←", "left panel")),
		Right:    key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("l/→", "right panel")),
		Top:      key.NewBinding(key.WithKeys("home"), key.WithHelp("gg", "top")),
		Bottom:   key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "bottom")),
		NextPane: key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "next panel")),
		PrevPane: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("Shift+Tab", "prev panel")),

		Search:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Command: key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "command")),
		Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "confirm")),
		Cancel:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "cancel")),

		ViewDashboard: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "dashboard")),
		ViewJournal:   key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "journal")),
		ViewKB:        key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "kb")),
		ViewTodo:      key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "todos")),
		ViewRecent:    key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "recent")),
		ViewTags:      key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "tags")),

		Quit:          key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		Rescan:        key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "rescan")),
		NewDoc:        key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new doc")),
		Open:          key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "open")),
		Scratch:       key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "scratch")),
		Export:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "export")),
		Delete:        key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		DeleteConfirm: key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm delete")),
		Pin:           key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pin/unpin")),
		SmartFilter:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "smart filter")),
		Lazygit:       key.NewBinding(key.WithKeys("ctrl+g"), key.WithHelp("ctrl+g", "lazygit")),
	}
}

// shortHelp returns a condensed list of keybindings for the status bar.
func (k KeyMap) shortHelp() []key.Binding {
	return []key.Binding{k.Search, k.Command, k.SmartFilter, k.Help, k.Quit}
}
