package tui

import (
	"github.com/AegirAexx/mdam/internal/git"
	"github.com/AegirAexx/mdam/internal/search"
	"github.com/AegirAexx/mdam/internal/todo"
)

// docsLoadedMsg is sent when the document scan completes.
type docsLoadedMsg struct {
	docs      []search.Result
	skipCount int // number of .md files skipped due to parse errors
	err       error
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

// editorReturnMsg is sent when an external process ($EDITOR) exits.
type editorReturnMsg struct {
	err error
}

// scratchReadyMsg is sent when the scratch pad file is confirmed to exist on disk.
type scratchReadyMsg struct {
	path string
}

// previewReadyMsg is sent when glamour has finished rendering a document preview.
type previewReadyMsg struct {
	content string
}

// pinsLoadedMsg is sent when the pinned document paths file has been read.
type pinsLoadedMsg struct {
	pins map[string]bool
	err  error
}

// tagIndexMsg is sent when the tag index has been built from the document list.
type tagIndexMsg struct {
	entries []tagEntry
}

// tickMsg is sent by the spinner tick to advance the loading animation frame.
type tickMsg struct{}

// readReadyMsg is sent when glamour has rendered the full-screen read content.
type readReadyMsg struct {
	content string
}
