package cli

import (
	"fmt"

	"github.com/AegirAexx/mdam/internal/export"
	"github.com/spf13/cobra"
)

var exportTo string

var exportCmd = &cobra.Command{
	Use:   "export <filename>",
	Short: "Export a document with frontmatter stripped",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		dest := exportTo
		if dest == "" {
			dest = cfg.ExportDir
		}

		destPath, err := export.ToFile(src, dest)
		if err != nil {
			return fmt.Errorf("export: %w", err)
		}
		cmd.Printf("exported to %s\n", destPath)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportTo, "to", "", "destination directory (default: export_dir from config)")
	rootCmd.AddCommand(exportCmd)
}
