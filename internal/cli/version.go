package cli

import (
	"fmt"

	"github.com/poiesic/wonda/internal/version"
	"github.com/spf13/cobra"
)

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display version, commit SHA, and build time",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Info())
	},
}
