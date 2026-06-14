package cmd

import (
	"fmt"
	"eko/internal/db"
	"eko/internal/snapshot"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [id]",
	Short: "Restore snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		database := db.InitDB()
		var path string
		database.QueryRow("SELECT path FROM snapshots WHERE id=?", id).Scan(&path)
		err := snapshot.RestoreSnapshot(path)
		if err != nil {
			panic(err)
		}
		fmt.Println("Restored:", id)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
