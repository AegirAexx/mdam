// Package cli wires the mdam engine to Cobra subcommands.
package cli

import (
	"fmt"
	"os"

	"github.com/AegirAexx/mdam/internal/config"
	"github.com/AegirAexx/mdam/internal/setup"
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
	cfgPath := cfgFile
	if cfgPath == "" {
		p, err := config.DefaultConfigPath()
		if err != nil {
			cfg, _ = config.LoadFrom("/dev/null")
			return
		}
		cfgPath = p
	}

	var err error
	cfg, err = config.LoadFrom(cfgPath)
	if err != nil {
		cfg, _ = config.LoadFrom("/dev/null")
	}

	if setup.IsFirstRun(cfgPath, cfg) {
		cfg, err = tui.RunWizard(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mdam: setup: %v\n", err)
			os.Exit(1)
		}
	} else {
		for _, w := range setup.ValidateConfig(cfg) {
			fmt.Fprintln(os.Stderr, "mdam: config: "+w)
		}
	}
}
