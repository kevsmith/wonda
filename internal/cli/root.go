package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

func init() {
	// Determine default config directory with precedence:
	// 1. --config-dir flag (handled by cobra automatically)
	// 2. $WONDA_HOME environment variable
	// 3. ~/.config/wonda (fallback)
	defaultConfig, source := getDefaultConfigDirWithSource()

	flagDescription := fmt.Sprintf("Path to Wonda configuration (source: %s)", source)
	rootCommand.PersistentFlags().StringVarP(&configDir, "config-dir", "c", defaultConfig, flagDescription)
	rootCommand.AddCommand(initCommand, nukeCommand, providersCommand, embeddingsCommand, modelsCommand, charactersCommand, scenariosCommand, versionCommand)
}

// getDefaultConfigDirWithSource returns the default configuration directory
// and a description of where it came from.
// Checks $WONDA_HOME first, then falls back to ~/.config/wonda
func getDefaultConfigDirWithSource() (string, string) {
	// Check for WONDA_HOME environment variable
	if wandaHome := os.Getenv("WONDA_HOME"); wandaHome != "" {
		return wandaHome, "$WONDA_HOME"
	}

	// Fallback to ~/.config/wonda
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return path.Join(homeDir, ".config", "wonda"), "default"
}

var configDir string

var rootCommand = &cobra.Command{
	Use:   "wonda",
	Short: "Watch your characters surprise you",
	Long:  `Your creative sandbox for character-driven storytelling`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
