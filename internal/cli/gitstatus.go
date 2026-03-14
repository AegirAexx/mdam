package cli

import (
	"fmt"

	"github.com/AegirAexx/mdam/internal/git"
	"github.com/spf13/cobra"
)

var statusPorcelain bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show git status summary for the managed tree",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.IsAvailable() {
			return fmt.Errorf("git is not available on PATH")
		}
		if !git.IsRepo(cfg.BaseDir) {
			cmd.Printf("%s is not a git repository\n", cfg.BaseDir)
			return nil
		}

		s, err := git.Status(cfg.BaseDir)
		if err != nil {
			return fmt.Errorf("git status: %w", err)
		}

		if statusPorcelain {
			for _, f := range s.Files {
				cmd.Printf("%c%c %s\n", f.X, f.Y, f.Path)
			}
			return nil
		}

		cmd.Printf("branch: %s\n", s.Branch)
		if s.Ahead > 0 || s.Behind > 0 {
			cmd.Printf("ahead: %d  behind: %d\n", s.Ahead, s.Behind)
		}
		cmd.Printf("uncommitted: %d\n", s.UncommittedCount())
		if s.StashCount > 0 {
			cmd.Printf("stash: %d\n", s.StashCount)
		}
		return nil
	},
}

func init() {
	statusCmd.Flags().BoolVar(&statusPorcelain, "porcelain", false, "machine-readable output")
	rootCmd.AddCommand(statusCmd)
}
