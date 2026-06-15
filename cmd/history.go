package cmd

import (
	"eko/internal/db"
	"fmt"

	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:     "history",
	Short:   "Show snapshots",
	PreRunE: requireInitialized,
	Run: func(cmd *cobra.Command, args []string) {
		database := db.InitDB()
		rows, _ := database.Query("SELECT id, created_at FROM snapshots")
		for rows.Next() {
			var id, created string
			rows.Scan(&id, &created)
			fmt.Println(id, created)
		}
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
