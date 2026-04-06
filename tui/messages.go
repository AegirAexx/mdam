package tui

import (
	"github.com/AegirAexx/mdam/internal/git"
	"github.com/AegirAexx/mdam/internal/search"
)

// docsLoadedMsg is sent when the document scan completes.
type docsLoadedMsg struct {
	docs      []search.Result
	skipCount int // number of .md files skipped due to parse errors
	err       error
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

// todoReadyMsg is sent when the todo file is confirmed to exist on disk.
type todoReadyMsg struct {
	path string
}

// dashTodoReadyMsg is sent when the todo file has been glamour-rendered for the dashboard.
type dashTodoReadyMsg struct {
	content string
	err     error
}

// previewReadyMsg is sent when glamour has finished rendering a document preview.
type previewReadyMsg struct {
	content string
}

// pinsLoadedMsg is sent when the pinned document paths file has been read.
type pinsLoadedMsg struct {
	pins []string // ordered list (oldest first)
	err  error
}

// tagIndexMsg is sent when the tag index has been built from the document list.
type tagIndexMsg struct {
	entries []tagEntry
}

// tickMsg is sent by the spinner tick to advance the loading animation frame.
type tickMsg struct{}

// journalAutoCreateMsg is sent when the auto-create journal startup task completes.
type journalAutoCreateMsg struct {
	created bool // true if a new entry was written, false if it already existed
	err     error
}

// readReadyMsg is sent when glamour has rendered the full-screen read content.
type readReadyMsg struct {
	content string
}
