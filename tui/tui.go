package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/AegirAexx/mdam/internal/config"
)

// Run starts the mdam TUI with the given configuration. It blocks until the user quits.
func Run(cfg config.Config) error {
	m := New(cfg)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // use the alternate screen buffer
		tea.WithMouseCellMotion(), // enable mouse support (Phase 5 will use it)
	)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}

// RunWithOutput is like Run but writes debug output to the given file path.
// Used for development; not exposed in the CLI.
func RunWithOutput(cfg config.Config, logPath string) error {
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer f.Close()

	m := New(cfg)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithoutRenderer(), // suppress rendering when logging
	)
	_, _ = tea.LogToFile(logPath, "tui")
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
