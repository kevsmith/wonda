package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var nukeCommand = &cobra.Command{
	Use:   "nuke",
	Short: "Delete Wonda configuration",
	Run:   deleteConfig,
}

func deleteConfig(cmd *cobra.Command, args []string) {
	info, err := os.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			reportSuccess("Done")
		} else {
			reportErrorAndDieP(configDir, err)
		}
	}
	if !info.IsDir() {
		reportErrorAndDieP(configDir, fmt.Errorf("%s is not a directory", configDir))
	}
	if askForConfirmation("Type 'wonda' with no quotes to delete your configuration>", "wonda") {
		if err := os.RemoveAll(configDir); err != nil {
			reportErrorAndDieP(configDir, err)
		}
		reportSuccess("Done")
	} else {
		os.Exit(1)
	}
}
