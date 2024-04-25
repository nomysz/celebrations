package cmd

import (
	"fmt"
	"os"

	"github.com/nomysz/celebrations/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "celebrations",
	Short: "Celebrate your company birthdays and anniversaries",
	Long:  "Set of tools facilitating company anniversaries and birthdays",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	cobra.OnInitialize(config.InitConfig)
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(DownloadUsers)
	rootCmd.AddCommand(SendRemindersCmd)
	rootCmd.AddCommand(VersionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
