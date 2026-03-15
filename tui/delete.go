package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// deleteDoneMsg is sent when a document deletion completes.
type deleteDoneMsg struct {
	path string
	err  error
}

// cmdDeleteDoc removes the file at path and sends deleteDoneMsg.
func cmdDeleteDoc(path string) tea.Cmd {
	return func() tea.Msg {
		if err := os.Remove(path); err != nil {
			return deleteDoneMsg{path: path, err: fmt.Errorf("removing file: %w", err)}
		}
		return deleteDoneMsg{path: path}
	}
}
