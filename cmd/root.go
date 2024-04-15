package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "celebrations",
	Short: "Celebrate your company birthdays and anniversaries",
	Long:  "Celebrartions is a set of tools that will help you manage your company birthdays and anniversaries",
	Run:   func(cmd *cobra.Command, args []string) { sendReminders() },
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(download_users)
	rootCmd.AddCommand(send_reminders)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
