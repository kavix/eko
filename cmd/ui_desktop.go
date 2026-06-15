//go:build !no_gui

package cmd

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
)

var UIAssets embed.FS

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Start the Eko visual memory UI",
	Long:  `Starts the Eko native Wails desktop application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// This block is never executed in the desktop build because
		// main.go intercepts the "ui" argument and starts Wails.
		fmt.Println("✦ Starting Eko Native UI...")
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}
