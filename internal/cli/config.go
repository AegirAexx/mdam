package cli

import (
	"fmt"

	"github.com/AegirAexx/mdam/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := cfgFile
		if path == "" {
			var err error
			path, err = config.DefaultConfigPath()
			if err != nil {
				return err
			}
		}
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "config file: %s\n\n", path)
		fmt.Fprintf(out, "editor:      %s\n", cfg.Editor)
		fmt.Fprintf(out, "author:      %s\n", cfg.Author)
		fmt.Fprintf(out, "base_dir:    %s\n", cfg.BaseDir)
		fmt.Fprintf(out, "export_dir:  %s\n", cfg.ExportDir)
		fmt.Fprintf(out, "theme:       %s\n", cfg.Theme)
		fmt.Fprintf(out, "\ngit:\n")
		fmt.Fprintf(out, "  enabled:     %v\n", cfg.Git.Enabled)
		fmt.Fprintf(out, "  auto_commit: %v\n", cfg.Git.AutoCommit)
		fmt.Fprintf(out, "\njournal:\n")
		fmt.Fprintf(out, "  auto_create:     %v\n", cfg.Journal.AutoCreate)
		fmt.Fprintf(out, "  sweep_on_create: %v\n", cfg.Journal.SweepOnCreate)
		fmt.Fprintf(out, "\ntodo:\n")
		fmt.Fprintf(out, "  default_category:  %s\n", cfg.Todo.DefaultCategory)
		fmt.Fprintf(out, "  archive_after_days: %d\n", cfg.Todo.ArchiveAfterDays)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
