//go:build no_gui

package cmd

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
)

var UIAssets embed.FS

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Start the Eko visual memory UI (Disabled)",
	Long:  `The graphical UI is not enabled in this build of Eko. Please use the Wails desktop build.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⚠ Eko UI is not compiled in this binary.")
		fmt.Println("Please run Eko using the Wails desktop build or install the Wails version.")
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}
