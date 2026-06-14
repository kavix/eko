package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"eko/internal/api"
	"eko/internal/db"

	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Start the Eko visual memory UI",
	Long: `Starts the Eko REST API on :7700 and opens the UI in your browser.
Run 'cd ui && npm run dev' in a separate terminal to start the Next.js frontend.`,
	Run: func(cmd *cobra.Command, args []string) {
		database := db.InitDB()

		fmt.Println("✦ Eko UI")
		fmt.Println("  API  → http://localhost:7700")
		fmt.Println("  App  → http://localhost:3000 (run: cd ui && npm run dev)")
		fmt.Println()

		// Open the browser after a short delay so the server can start.
		go openBrowser("http://localhost:3000")

		if err := api.Serve(":7700", database); err != nil {
			fmt.Println("API server error:", err)
		}
	},
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "linux":
		cmd, args = "xdg-open", []string{url}
	default:
		cmd, args = "cmd", []string{"/c", "start", url}
	}
	exec.Command(cmd, args...).Start()
}

func init() {
	rootCmd.AddCommand(uiCmd)
}
