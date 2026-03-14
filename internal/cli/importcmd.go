package cli

import (
	"fmt"

	"github.com/AegirAexx/mdam/internal/importer"
	"github.com/spf13/cobra"
)

var (
	importAutoFix bool
	importDryRun  bool
)

var importCmd = &cobra.Command{
	Use:   "import <path>",
	Short: "Import a file or directory into the managed tree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		opts := importer.Options{
			AutoFix: importAutoFix,
			DryRun:  importDryRun,
			DestDir: cfg.BaseDir,
		}

		// Try as directory first, then single file.
		results, err := importer.ImportDir(path, cfg.BaseDir, opts)
		if err != nil {
			// Not a directory or read error — try as single file.
			result, fileErr := importer.ImportFile(path, cfg.BaseDir, opts)
			if fileErr != nil {
				return fmt.Errorf("import: %w", fileErr)
			}
			results = []importer.Result{result}
		}

		ok, skipped, fixed, errCount := 0, 0, 0, 0
		for _, r := range results {
			if r.Skipped {
				skipped++
				cmd.Printf("SKIP  %s (duplicate)\n", r.SourcePath)
				continue
			}
			if len(r.Errors) > 0 {
				errCount++
				for _, e := range r.Errors {
					cmd.Printf("ERROR %s\n", e)
				}
				continue
			}
			if r.Fixed {
				fixed++
			}
			ok++
			if importDryRun {
				cmd.Printf("OK    %s (dry-run)\n", r.SourcePath)
			} else {
				cmd.Printf("OK    %s\n", r.DestPath)
			}
		}

		cmd.Printf("\n%d imported, %d skipped, %d fixed, %d errors\n", ok, skipped, fixed, errCount)
		return nil
	},
}

func init() {
	importCmd.Flags().BoolVar(&importAutoFix, "auto-fix", false, "auto-fix filename and frontmatter issues")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "report issues without modifying files")
	rootCmd.AddCommand(importCmd)
}
