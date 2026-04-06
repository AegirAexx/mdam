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
		fmt.Fprintf(out, "nerd_fonts:  %v\n", cfg.NerdFonts)
		fmt.Fprintf(out, "\njournal:\n")
		fmt.Fprintf(out, "  auto_create: %v\n", cfg.Journal.AutoCreate)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
