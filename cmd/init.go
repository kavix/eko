package cmd

import (
	"fmt"
	"os"
	"eko/internal/db"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize vibe project",
	Run: func(cmd *cobra.Command, args []string) {
		os.MkdirAll(".vibe/snapshots", 0755)
		database := db.InitDB()
		database.Exec(`
			CREATE TABLE IF NOT EXISTS snapshots (
				id TEXT PRIMARY KEY,
				message TEXT,
				path TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`)
		fmt.Println("Vibe initialized.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
