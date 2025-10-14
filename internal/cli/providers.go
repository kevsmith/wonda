package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var providersCommand = &cobra.Command{
	Use:     "providers",
	Short:   "Work with wonda provider configuration",
	Aliases: []string{"p"},
}

var showProviderCommand = &cobra.Command{
	Use:   "show",
	Short: "View providers configuration",
	Run:   showProvider,
}

var editProviderCommand = &cobra.Command{
	Use:   "edit",
	Short: "Open providers.toml in $EDITOR",
	Run:   editProvider,
}

var editors = []string{"vi", "vim", "nvi", "nano"}

func init() {
	providersCommand.AddCommand(showProviderCommand, editProviderCommand)
}

func showProvider(cmd *cobra.Command, args []string) {
	tomlFile := path.Join(configDir, "providers.toml")
	contents, err := os.ReadFile(tomlFile)
	if err != nil {
		reportErrorAndDieP(tomlFile, err)
	}
	fmt.Printf("PATH: %s\n", tomlFile)
	fmt.Println(string(contents))
}

func editProvider(cmd *cobra.Command, args []string) {
	tomlFile := path.Join(configDir, "providers.toml")
	if _, err := os.Stat(tomlFile); err != nil {
		reportErrorAndDieP(tomlFile, err)
	}
	editFile(tomlFile)
}
