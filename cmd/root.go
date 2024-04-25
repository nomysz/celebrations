package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/nomysz/celebrations/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "celebrations",
	Short: "Celebrate your company birthdays and anniversaries",
	Long:  "Celebrartions is a set of tools that will help you manage your company birthdays and anniversaries",
	Run:   func(cmd *cobra.Command, args []string) { log.Fatalln("Command not specified") },
}

func init() {
	cobra.OnInitialize(config.InitConfig)
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(download_users)
	rootCmd.AddCommand(SendRemindersCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
