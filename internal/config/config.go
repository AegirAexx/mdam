// Package config handles loading and validating mdam configuration from
// ~/.config/mdam/config.yml using Viper.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// JournalConfig holds journal behaviour settings.
type JournalConfig struct {
	AutoCreate bool `mapstructure:"auto_create"`
}

// Config is the top-level mdam configuration.
type Config struct {
	Editor    string        `mapstructure:"editor"`
	Author    string        `mapstructure:"author"`
	BaseDir   string        `mapstructure:"base_dir"`
	ExportDir string        `mapstructure:"export_dir"`
	Theme     string        `mapstructure:"theme"`
	NerdFonts bool          `mapstructure:"nerd_fonts"`
	Journal   JournalConfig `mapstructure:"journal"`
}

// DefaultConfigPath returns the default path for the config file.
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	return filepath.Join(home, ".config", "mdam", "config.yml"), nil
}

// Load reads the config file at the default location and returns a Config.
// If the file does not exist, defaults are returned without error.
func Load() (Config, error) {
	path, err := DefaultConfigPath()
	if err != nil {
		return Config{}, err
	}
	return LoadFrom(path)
}

// LoadFrom reads a config file from a specific path.
func LoadFrom(path string) (Config, error) {
	v := viper.New()
	setDefaults(v)

	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			// No config file — return defaults.
			return unmarshal(v)
		}
		return Config{}, fmt.Errorf("reading config: %w", err)
	}
	return unmarshal(v)
}

func unmarshal(v *viper.Viper) (Config, error) {
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	cfg = expandPaths(cfg)
	return cfg, nil
}

// setDefaults configures Viper defaults matching the spec's config.yml example.
func setDefaults(v *viper.Viper) {
	home, _ := os.UserHomeDir()

	v.SetDefault("editor", os.Getenv("EDITOR"))
	v.SetDefault("author", "")
	v.SetDefault("base_dir", "")
	v.SetDefault("export_dir", filepath.Join(home, "Downloads"))
	v.SetDefault("theme", "tokyonight")
	v.SetDefault("nerd_fonts", false)

	v.SetDefault("journal.auto_create", true)
}

// expandPaths expands ~ in directory paths.
func expandPaths(cfg Config) Config {
	cfg.BaseDir = expandHome(cfg.BaseDir)
	cfg.ExportDir = expandHome(cfg.ExportDir)
	return cfg
}

func expandHome(path string) string {
	if len(path) < 2 || path[:2] != "~/" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}

// JournalDir returns the journal subdirectory within BaseDir.
func (c Config) JournalDir() string {
	return filepath.Join(c.BaseDir, "journal")
}

// KBDir returns the knowledge base subdirectory within BaseDir.
func (c Config) KBDir() string {
	return filepath.Join(c.BaseDir, "kb")
}

// TemplatesDir returns the templates subdirectory within BaseDir.
func (c Config) TemplatesDir() string {
	return filepath.Join(c.BaseDir, ".templates")
}

// TodoPath returns the path for the global TODO file.
func (c Config) TodoPath() string {
	return filepath.Join(c.BaseDir, "todo.md")
}

// ScratchPath returns the path for the scratch pad file.
func (c Config) ScratchPath() string {
	return filepath.Join(c.BaseDir, "scratch.md")
}

// MdamDir returns the .mdam metadata directory within BaseDir.
// This directory lives in version control alongside the managed documents.
func (c Config) MdamDir() string {
	return filepath.Join(c.BaseDir, ".mdam")
}

// PinsPath returns the path for the pinned documents file.
// Stored in {BaseDir}/.mdam/ so pins live in version control.
func (c Config) PinsPath() string {
	return filepath.Join(c.MdamDir(), "pins.json")
}
