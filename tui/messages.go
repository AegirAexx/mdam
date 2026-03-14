package tui

import (
	"github.com/AegirAexx/mdam/internal/git"
	"github.com/AegirAexx/mdam/internal/search"
	"github.com/AegirAexx/mdam/internal/todo"
)

// docsLoadedMsg is sent when the document scan completes.
type docsLoadedMsg struct {
	docs []search.Result
	err  error
}

// todosLoadedMsg is sent when TODO tasks are loaded from disk.
type todosLoadedMsg struct {
	tasks []todo.Task
	err   error
}

// gitStatusMsg is sent when git status detection completes.
type gitStatusMsg struct {
	status git.RepoStatus
	err    error
}

// searchDoneMsg is sent when a fuzzy search completes.
type searchDoneMsg struct {
	results []search.Result
	query   string
	err     error
}

// exportDoneMsg is sent when an export operation completes.
type exportDoneMsg struct {
	dest string
	err  error
}

// sweepDoneMsg is sent when a TODO sweep or archive completes.
type sweepDoneMsg struct {
	err error
}

// fileCreatedMsg is sent when a new document has been written to disk.
type fileCreatedMsg struct {
	path string
	err  error
}
