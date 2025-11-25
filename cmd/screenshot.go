package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	screenshotOut string
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Take a screenshot of the guest console",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Stub for now as it requires complex raw API handling
		fmt.Println("Screenshot functionality is currently disabled in this version.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(screenshotCmd)
	screenshotCmd.Flags().StringVar(&screenshotOut, "out", "screenshot.png", "Output file path")
}