package cli

import (
	"fmt"
	"time"

	"github.com/AegirAexx/mdam/internal/journal"
	"github.com/spf13/cobra"
)

var journalCmd = &cobra.Command{
	Use:   "journal",
	Short: "Manage daily journal entries",
}

var journalCreateCmd = &cobra.Command{
	Use:   "create [date]",
	Short: "Create a journal entry (defaults to today)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		date := time.Now()
		if len(args) == 1 {
			var err error
			date, err = journal.ParseDate(args[0])
			if err != nil {
				return err
			}
		}

		path, err := journal.Create(cfg.JournalDir(), cfg.TemplatesDir(), date)
		if err != nil {
			return fmt.Errorf("creating journal entry: %w", err)
		}
		if journal.Exists(cfg.JournalDir(), date) {
			cmd.Printf("journal entry: %s\n", path)
		}
		return nil
	},
}

var journalListMonth string

var journalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List journal entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		var paths []string
		var err error

		if journalListMonth != "" {
			paths, err = journal.ListByMonth(cfg.JournalDir(), journalListMonth)
		} else {
			paths, err = journal.List(cfg.JournalDir())
		}
		if err != nil {
			return fmt.Errorf("listing journals: %w", err)
		}

		if len(paths) == 0 {
			cmd.Println("no journal entries found")
			return nil
		}
		for _, p := range paths {
			cmd.Println(p)
		}
		return nil
	},
}

func init() {
	journalListCmd.Flags().StringVar(&journalListMonth, "month", "", "filter by month (YYYY-MM)")
	journalCmd.AddCommand(journalCreateCmd, journalListCmd)
	rootCmd.AddCommand(journalCmd)
}
