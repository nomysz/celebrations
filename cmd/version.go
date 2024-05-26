package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.2.1"

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show app version",
	Long:  "Show app version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("celebrations v%s", Version))
	},
}
