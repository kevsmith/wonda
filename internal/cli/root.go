package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	defaultConfig := path.Join(homeDir, ".config", "wonda")
	rootCommand.PersistentFlags().StringVarP(&configDir, "config-dir", "c", defaultConfig, "path to Wonda configuration")
	rootCommand.AddCommand(initCommand, nukeCommand, providersCommand, modelsCommand, charactersCommand, scenariosCommand)
}

var configDir string

var rootCommand = &cobra.Command{
	Use:   "wonda",
	Short: "Watch your characters surprise you",
	Long:  `Your creative sandbox for character-driven storytelling`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
