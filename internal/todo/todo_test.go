package todo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseTask(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantOk       bool
		wantStatus   string
		wantText     string
		wantCategory string
		wantPriority string
	}{
		{
			name:       "open task",
			line:       "- [ ] Review PR #42",
			wantOk:     true,
			wantStatus: StatusOpen,
			wantText:   "Review PR #42",
		},
		{
			name:       "done task",
			line:       "- [x] Update DNS records",
			wantOk:     true,
			wantStatus: StatusDone,
			wantText:   "Update DNS records",
		},
		{
			name:       "cancelled task",
			line:       "- [-] Old idea",
			wantOk:     true,
			wantStatus: StatusCancelled,
			wantText:   "Old idea",
		},
		{
			name:         "task with category",
			line:         "- [ ] Buy groceries @personal",
			wantOk:       true,
			wantStatus:   StatusOpen,
			wantText:     "Buy groceries",
			wantCategory: "personal",
		},
		{
			name:         "task with priority",
			line:         "- [ ] Deploy fix !high",
			wantOk:       true,
			wantStatus:   StatusOpen,
			wantText:     "Deploy fix",
			wantPriority: "high",
		},
		{
			name:         "full task",
			line:         "- [ ] Review PR #42 @work !high (2026-03-14)",
			wantOk:       true,
			wantStatus:   StatusOpen,
			wantText:     "Review PR #42",
			wantCategory: "work",
			wantPriority: "high",
		},
		{
			name:   "not a task",
			line:   "# Heading",
			wantOk: false,
		},
		{
			name:   "blank line",
			line:   "",
			wantOk: false,
		},
		{
			name:       "indented task",
			line:       "  - [ ] Subtask",
			wantOk:     true,
			wantStatus: StatusOpen,
			wantText:   "Subtask",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, ok := ParseTask(tt.line)
			if ok != tt.wantOk {
				t.Fatalf("ParseTask() ok = %v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if task.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", task.Status, tt.wantStatus)
			}
			if tt.wantText != "" && task.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", task.Text, tt.wantText)
			}
			if tt.wantCategory != "" && task.Category != tt.wantCategory {
				t.Errorf("Category = %q, want %q", task.Category, tt.wantCategory)
			}
			if tt.wantPriority != "" && task.Priority != tt.wantPriority {
				t.Errorf("Priority = %q, want %q", task.Priority, tt.wantPriority)
			}
		})
	}
}

func TestParseTasks(t *testing.T) {
	content := `## TODOs

- [ ] Open task @work
- [x] Done task ✓2026-03-13
- [-] Cancelled task

Some plain text that isn't a task.
`
	tasks := ParseTasks(content)
	if len(tasks) != 3 {
		t.Fatalf("ParseTasks() returned %d tasks, want 3", len(tasks))
	}
	if tasks[0].Status != StatusOpen {
		t.Errorf("tasks[0].Status = %q, want open", tasks[0].Status)
	}
	if tasks[1].Status != StatusDone {
		t.Errorf("tasks[1].Status = %q, want done", tasks[1].Status)
	}
	if tasks[2].Status != StatusCancelled {
		t.Errorf("tasks[2].Status = %q, want cancelled", tasks[2].Status)
	}
}

func TestTaskHelpers(t *testing.T) {
	open := Task{Status: StatusOpen}
	inProgress := Task{Status: StatusInProgress}
	done := Task{Status: StatusDone}
	cancelled := Task{Status: StatusCancelled}

	if !open.IsOpen() || !inProgress.IsOpen() {
		t.Error("IsOpen() failed for open/in-progress")
	}
	if done.IsOpen() || cancelled.IsOpen() {
		t.Error("IsOpen() true for done/cancelled")
	}
	if !done.IsDone() || !cancelled.IsDone() {
		t.Error("IsDone() failed for done/cancelled")
	}
	if open.IsDone() || inProgress.IsDone() {
		t.Error("IsDone() true for open/in-progress")
	}
}

func TestSweep(t *testing.T) {
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "2026-03-13.md")
	todoPath := filepath.Join(dir, "todo.md")

	journalContent := `---
title: Journal 2026-03-13
type: journal
---

## Notes

Some notes.

## TODOs

- [ ] Carry forward task @work
- [x] Completed task ✓2026-03-13
- [ ] Another open task

## Extra section

More content.
`
	if err := os.WriteFile(journalPath, []byte(journalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Sweep(journalPath, todoPath); err != nil {
		t.Fatalf("Sweep() error = %v", err)
	}

	// Check journal: open tasks should be removed, done should remain.
	journalData, _ := os.ReadFile(journalPath)
	j := string(journalData)
	if strings.Contains(j, "Carry forward task") {
		t.Error("Sweep() did not remove open task from journal")
	}
	if strings.Contains(j, "Another open task") {
		t.Error("Sweep() did not remove second open task from journal")
	}
	if !strings.Contains(j, "Completed task") {
		t.Error("Sweep() removed done task from journal")
	}
	if !strings.Contains(j, "Some notes") {
		t.Error("Sweep() removed non-TODO content from journal")
	}
	if !strings.Contains(j, "Extra section") {
		t.Error("Sweep() removed content after TODO section")
	}

	// Check global TODO: open tasks should have been promoted.
	todoData, _ := os.ReadFile(todoPath)
	td := string(todoData)
	if !strings.Contains(td, "Carry forward task") {
		t.Error("Sweep() did not promote open task to global todo")
	}
	if !strings.Contains(td, "Another open task") {
		t.Error("Sweep() did not promote second open task to global todo")
	}
	if strings.Contains(td, "Completed task") {
		t.Error("Sweep() promoted done task to global todo")
	}
}

func TestSweepNoDuplicates(t *testing.T) {
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "journal.md")
	todoPath := filepath.Join(dir, "todo.md")

	existing := "- [ ] Already in todo @work\n"
	if err := os.WriteFile(todoPath, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	journalContent := `---
title: Journal
---
## TODOs

- [ ] Already in todo @work
- [ ] New task
`
	if err := os.WriteFile(journalPath, []byte(journalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Sweep(journalPath, todoPath); err != nil {
		t.Fatalf("Sweep() error = %v", err)
	}

	todoData, _ := os.ReadFile(todoPath)
	count := strings.Count(string(todoData), "Already in todo")
	if count != 1 {
		t.Errorf("Sweep() created duplicate: found %d occurrences", count)
	}
}

func TestSweepNoTodoSection(t *testing.T) {
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "journal.md")
	todoPath := filepath.Join(dir, "todo.md")

	journalContent := "---\ntitle: Journal\n---\n\n## Notes\n\nJust notes.\n"
	if err := os.WriteFile(journalPath, []byte(journalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Sweep(journalPath, todoPath); err != nil {
		t.Fatalf("Sweep() on no-TODO journal error = %v", err)
	}

	// Todo file should not have been created.
	if _, err := os.Stat(todoPath); err == nil {
		t.Error("Sweep() created todo file when journal had no TODO section")
	}
}

func TestArchive(t *testing.T) {
	dir := t.TempDir()
	todoPath := filepath.Join(dir, "todo.md")
	archivePath := filepath.Join(dir, "archive.md")

	oldDate := time.Now().AddDate(0, 0, -40).Format("2006-01-02")
	recentDate := time.Now().AddDate(0, 0, -5).Format("2006-01-02")

	todoContent := "- [ ] Open task\n" +
		"- [x] Old done task ✓" + oldDate + "\n" +
		"- [x] Recent done task ✓" + recentDate + "\n"

	if err := os.WriteFile(todoPath, []byte(todoContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Archive(todoPath, archivePath, 30*24*time.Hour); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}

	todoData, _ := os.ReadFile(todoPath)
	td := string(todoData)
	if strings.Contains(td, "Old done task") {
		t.Error("Archive() did not remove old done task from todo")
	}
	if !strings.Contains(td, "Open task") {
		t.Error("Archive() removed open task from todo")
	}
	if !strings.Contains(td, "Recent done task") {
		t.Error("Archive() removed recently done task from todo")
	}

	archiveData, _ := os.ReadFile(archivePath)
	if !strings.Contains(string(archiveData), "Old done task") {
		t.Error("Archive() did not move old done task to archive")
	}
}

func TestArchiveMissingTodo(t *testing.T) {
	dir := t.TempDir()
	if err := Archive(filepath.Join(dir, "todo.md"), filepath.Join(dir, "archive.md"), 30*24*time.Hour); err != nil {
		t.Fatalf("Archive() on missing file error = %v", err)
	}
}

func TestFilterTasks(t *testing.T) {
	tasks := []Task{
		{Status: StatusOpen, Category: "work"},
		{Status: StatusDone, Category: "personal"},
		{Status: StatusOpen, Category: "personal"},
		{Status: StatusInProgress, Category: "work"},
	}

	open := FilterTasks(tasks, StatusOpen, "")
	if len(open) != 2 {
		t.Errorf("FilterTasks(open) = %d, want 2", len(open))
	}

	work := FilterTasks(tasks, "", "work")
	if len(work) != 2 {
		t.Errorf("FilterTasks(work) = %d, want 2", len(work))
	}

	openWork := FilterTasks(tasks, StatusOpen, "work")
	if len(openWork) != 1 {
		t.Errorf("FilterTasks(open, work) = %d, want 1", len(openWork))
	}

	all := FilterTasks(tasks, "", "")
	if len(all) != 4 {
		t.Errorf("FilterTasks(all) = %d, want 4", len(all))
	}
}

func TestReadTasks(t *testing.T) {
	dir := t.TempDir()

	// Missing file should return empty, not error.
	tasks, err := ReadTasks(filepath.Join(dir, "missing.md"))
	if err != nil {
		t.Fatalf("ReadTasks() on missing file error = %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("ReadTasks() on missing file = %v, want empty", tasks)
	}
}

func TestCompletedDateParsing(t *testing.T) {
	line := "- [x] Fix bug @work ✓2026-03-13"
	task, ok := ParseTask(line)
	if !ok {
		t.Fatal("ParseTask() returned false")
	}
	if task.Completed.IsZero() {
		t.Error("Completed date not parsed")
	}
	if task.Completed.Format("2006-01-02") != "2026-03-13" {
		t.Errorf("Completed = %v, want 2026-03-13", task.Completed)
	}
}
