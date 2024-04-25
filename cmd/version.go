package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Send remidners via configured handlers",
	Long:  "Send remidners via configured handlers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("celebrations v1.0.0")
	},
}
