package cli

import (
	"fmt"

	"github.com/AegirAexx/mdam/internal/journal"
	"github.com/AegirAexx/mdam/internal/todo"
	"github.com/spf13/cobra"
)

var todoCmd = &cobra.Command{
	Use:    "todo",
	Short:  "Manage TODO tasks",
	Hidden: true,
}

var (
	todoListStatus   string
	todoListCategory string
	todoListAll      bool
)

var todoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TODO tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := todo.ReadTasks(cfg.TodoPath())
		if err != nil {
			return fmt.Errorf("reading todos: %w", err)
		}

		var filtered []todo.Task
		if todoListAll {
			filtered = tasks
		} else if todoListStatus != "" {
			filtered = todo.FilterTasks(tasks, todoListStatus, todoListCategory)
		} else {
			// Default: open and in-progress tasks.
			for _, t := range tasks {
				if t.IsOpen() && (todoListCategory == "" || t.Category == todoListCategory) {
					filtered = append(filtered, t)
				}
			}
		}

		if len(filtered) == 0 {
			cmd.Println("no tasks found")
			return nil
		}
		for _, t := range filtered {
			marker := " "
			switch t.Status {
			case todo.StatusDone:
				marker = "x"
			case todo.StatusCancelled:
				marker = "-"
			case todo.StatusInProgress:
				marker = "~"
			}
			cat := ""
			if t.Category != "" {
				cat = " @" + t.Category
			}
			cmd.Printf("[%s] %s%s\n", marker, t.Text, cat)
		}
		return nil
	},
}

var todoSweepCmd = &cobra.Command{
	Use:   "sweep",
	Short: "Sweep incomplete tasks from past journal entries to global TODO",
	RunE: func(cmd *cobra.Command, args []string) error {
		past, err := journal.PastEntries(cfg.JournalDir())
		if err != nil {
			return fmt.Errorf("listing journal entries: %w", err)
		}
		swept := 0
		for _, p := range past {
			if err := todo.Sweep(p, cfg.TodoPath()); err != nil {
				return fmt.Errorf("sweeping %s: %w", p, err)
			}
			swept++
		}
		cmd.Printf("swept %d journal entries\n", swept)
		return nil
	},
}

var todoArchiveOlderThan int

var todoArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive old completed tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("todo archive is not yet available")
		return nil
	},
}

func init() {
	todoListCmd.Flags().StringVar(&todoListStatus, "status", "", "filter by status (open, in-progress, done, cancelled)")
	todoListCmd.Flags().StringVar(&todoListCategory, "category", "", "filter by category")
	todoListCmd.Flags().BoolVar(&todoListAll, "all", false, "show all tasks including done and cancelled")

	todoArchiveCmd.Flags().IntVar(&todoArchiveOlderThan, "older-than", 30, "archive tasks completed more than N days ago")

	todoCmd.AddCommand(todoListCmd, todoSweepCmd, todoArchiveCmd)
	rootCmd.AddCommand(todoCmd)
}
