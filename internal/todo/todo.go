// Package todo handles TODO parsing, sweep logic, and archiving for mdam.
package todo

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Status values for tasks.
const (
	StatusOpen       = "open"
	StatusInProgress = "in-progress"
	StatusDone       = "done"
	StatusCancelled  = "cancelled"
)

// Task represents a single TODO item parsed from markdown.
type Task struct {
	// Raw is the original line text.
	Raw string
	// Status is the task status: open, in-progress, done, cancelled.
	Status string
	// Text is the task description without metadata markers.
	Text string
	// Category is the @category value, if present.
	Category string
	// Priority is the !priority value, if present.
	Priority string
	// Created is the task creation date parsed from (YYYY-MM-DD).
	Created time.Time
	// Completed is the completion date parsed from ✓YYYY-MM-DD, if present.
	Completed time.Time
}

// IsOpen reports whether the task is unfinished.
func (t Task) IsOpen() bool {
	return t.Status == StatusOpen || t.Status == StatusInProgress
}

// IsDone reports whether the task is finished.
func (t Task) IsDone() bool {
	return t.Status == StatusDone || t.Status == StatusCancelled
}

var (
	// taskRe matches markdown checkbox lines: "- [ ] text" or "- [x] text"
	taskRe = regexp.MustCompile(`^(\s*)- \[([x X\-])\] (.+)$`)
	// categoryRe matches @category in task text.
	categoryRe = regexp.MustCompile(`@(\S+)`)
	// priorityRe matches !priority in task text.
	priorityRe = regexp.MustCompile(`!(\S+)`)
	// dateRe matches (YYYY-MM-DD) in task text.
	dateRe = regexp.MustCompile(`\((\d{4}-\d{2}-\d{2})\)`)
	// completedRe matches ✓YYYY-MM-DD in task text.
	completedRe = regexp.MustCompile(`✓(\d{4}-\d{2}-\d{2})`)
)

// ParseTask parses a single markdown checkbox line into a Task.
// Returns (Task{}, false) if the line is not a task.
func ParseTask(line string) (Task, bool) {
	m := taskRe.FindStringSubmatch(line)
	if m == nil {
		return Task{}, false
	}
	checkmark := strings.ToLower(m[2])
	text := m[3]

	status := StatusOpen
	switch checkmark {
	case "x":
		status = StatusDone
	case "-":
		status = StatusCancelled
	case " ":
		// Check for in-progress marker in the text (convention: ~text~).
		if strings.HasPrefix(text, "~") && strings.Contains(text[1:], "~") {
			status = StatusInProgress
		}
	}

	task := Task{
		Raw:    line,
		Status: status,
	}

	// Extract and strip category.
	if cm := categoryRe.FindStringSubmatch(text); cm != nil {
		task.Category = cm[1]
		text = strings.Replace(text, cm[0], "", 1)
	}

	// Extract and strip priority.
	if pm := priorityRe.FindStringSubmatch(text); pm != nil {
		task.Priority = pm[1]
		text = strings.Replace(text, pm[0], "", 1)
	}

	// Extract and strip created date.
	if dm := dateRe.FindStringSubmatch(text); dm != nil {
		if t, err := time.Parse("2006-01-02", dm[1]); err == nil {
			task.Created = t
		}
		text = strings.Replace(text, dm[0], "", 1)
	}

	// Extract and strip completed date.
	if cm := completedRe.FindStringSubmatch(text); cm != nil {
		if t, err := time.Parse("2006-01-02", cm[1]); err == nil {
			task.Completed = t
		}
		text = strings.Replace(text, cm[0], "", 1)
	}

	task.Text = strings.TrimSpace(text)
	return task, true
}

// ParseTasks scans content line by line and returns all tasks found.
func ParseTasks(content string) []Task {
	var tasks []Task
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		if task, ok := ParseTask(scanner.Text()); ok {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// ReadTasks reads and parses all tasks from a file.
func ReadTasks(path string) ([]Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, fmt.Errorf("reading todo file: %w", err)
	}
	return ParseTasks(string(data)), nil
}

// todoSection holds the parsed sections of a document with a TODO section.
type todoSection struct {
	beforeLines []string
	todoLines   []string
	afterLines  []string
	headerLine  int // index in todoLines of the section header
}

// parseTodoSection splits file content into before/TODO-section/after parts.
// The TODO section starts at a line matching "## TODOs" or "## TODO".
func parseTodoSection(content string) todoSection {
	lines := strings.Split(content, "\n")
	todoStart := -1
	todoEnd := len(lines)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## TODOs" || trimmed == "## TODO" {
			todoStart = i
			continue
		}
		// Next heading of same or higher level ends the TODO section.
		if todoStart != -1 && i > todoStart {
			if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "# ") {
				todoEnd = i
				break
			}
		}
	}

	if todoStart == -1 {
		return todoSection{beforeLines: lines}
	}

	return todoSection{
		beforeLines: lines[:todoStart],
		todoLines:   lines[todoStart:todoEnd],
		afterLines:  lines[todoEnd:],
		headerLine:  0,
	}
}

// Sweep processes a past journal entry's TODO section:
//   - Completed tasks remain in the journal.
//   - Incomplete tasks are removed from the journal and promoted to the global TODO.
//   - New tasks (not in global TODO) are added to the global TODO.
//
// Both journalPath and todoPath are read and written atomically via os.WriteFile.
func Sweep(journalPath, todoPath string) error {
	journalData, err := os.ReadFile(journalPath)
	if err != nil {
		return fmt.Errorf("reading journal: %w", err)
	}

	sec := parseTodoSection(string(journalData))
	if sec.todoLines == nil {
		// No TODO section — nothing to sweep.
		return nil
	}

	// Separate tasks into keep (done/cancelled) and promote (open/in-progress).
	var keepLines []string
	var promoteTasks []Task

	for _, line := range sec.todoLines {
		if task, ok := ParseTask(line); ok {
			if task.IsDone() {
				keepLines = append(keepLines, line)
			} else {
				promoteTasks = append(promoteTasks, task)
			}
		} else {
			// Non-task lines (header, blank lines) stay in the journal.
			keepLines = append(keepLines, line)
		}
	}

	// Rewrite journal keeping only the kept lines.
	newJournalLines := append(sec.beforeLines, keepLines...)
	newJournalLines = append(newJournalLines, sec.afterLines...)
	journalContent := strings.Join(newJournalLines, "\n")
	if err := os.WriteFile(journalPath, []byte(journalContent), 0o644); err != nil {
		return fmt.Errorf("writing journal after sweep: %w", err)
	}

	if len(promoteTasks) == 0 {
		return nil
	}

	// Read existing global TODO to avoid duplicates.
	existingTasks, err := ReadTasks(todoPath)
	if err != nil {
		return fmt.Errorf("reading global todo: %w", err)
	}
	existingTexts := make(map[string]bool, len(existingTasks))
	for _, t := range existingTasks {
		existingTexts[t.Text] = true
	}

	// Build new task lines to append.
	var newLines []string
	for _, task := range promoteTasks {
		if !existingTexts[task.Text] {
			newLines = append(newLines, task.Raw)
		}
	}

	if len(newLines) == 0 {
		return nil
	}

	// Append to global TODO (create if missing).
	todoData, err := os.ReadFile(todoPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading global todo: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(todoPath), 0o755); err != nil {
		return fmt.Errorf("creating todo directory: %w", err)
	}

	existing := strings.TrimRight(string(todoData), "\n")
	appended := existing + "\n" + strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(todoPath, []byte(appended), 0o644); err != nil {
		return fmt.Errorf("writing global todo: %w", err)
	}
	return nil
}

// Archive moves completed/cancelled tasks older than olderThan from todoPath
// to archivePath, preserving their content. Tasks are archived if their
// Completed date is set and older than the threshold.
func Archive(todoPath, archivePath string, olderThan time.Duration) error {
	data, err := os.ReadFile(todoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading todo file: %w", err)
	}

	threshold := time.Now().Add(-olderThan)
	var keepLines, archiveLines []string

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		task, ok := ParseTask(line)
		if !ok {
			keepLines = append(keepLines, line)
			continue
		}
		if task.IsDone() && !task.Completed.IsZero() && task.Completed.Before(threshold) {
			archiveLines = append(archiveLines, line)
		} else {
			keepLines = append(keepLines, line)
		}
	}

	if len(archiveLines) == 0 {
		return nil
	}

	// Write updated todo file.
	if err := os.WriteFile(todoPath, []byte(strings.Join(keepLines, "\n")+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing todo after archive: %w", err)
	}

	// Append to archive file.
	archiveData, _ := os.ReadFile(archivePath)
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		return fmt.Errorf("creating archive directory: %w", err)
	}
	existing := strings.TrimRight(string(archiveData), "\n")
	content := existing + "\n" + strings.Join(archiveLines, "\n") + "\n"
	if err := os.WriteFile(archivePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing archive: %w", err)
	}
	return nil
}

// FilterTasks returns tasks matching the given status (empty string = all).
func FilterTasks(tasks []Task, status, category string) []Task {
	var result []Task
	for _, t := range tasks {
		if status != "" && t.Status != status {
			continue
		}
		if category != "" && t.Category != category {
			continue
		}
		result = append(result, t)
	}
	return result
}
