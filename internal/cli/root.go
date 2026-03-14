// Package cli wires the mdam engine to Cobra subcommands.
package cli

import (
	"github.com/AegirAexx/mdam/internal/config"
	"github.com/AegirAexx/mdam/tui"
	"github.com/spf13/cobra"
)

var cfgFile string
var cfg config.Config

var rootCmd = &cobra.Command{
	Use:   "mdam",
	Short: "Markdown Admin Management — organise your markdown documents",
	Long: `mdam is a terminal tool for managing, organising, and navigating
markdown documents, journals, and TODOs. Run without a subcommand to
launch the interactive TUI (coming in Phase 2).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run(cfg)
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/mdam/config.yml)")
}

func initConfig() {
	var err error
	if cfgFile != "" {
		cfg, err = config.LoadFrom(cfgFile)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		// Non-fatal: fall back to defaults.
		cfg, _ = config.LoadFrom("/dev/null")
	}
}
