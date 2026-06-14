package cmd

import (
	"fmt"
	"eko/internal/db"
	"eko/internal/snapshot"

	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save project snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		id, path, err := snapshot.CreateSnapshot()
		if err != nil {
			panic(err)
		}
		database := db.InitDB()
		database.Exec(
			"INSERT INTO snapshots(id, message, path) VALUES (?, ?, ?)",
			id,
			"snapshot",
			path,
		)
		fmt.Println("Snapshot saved:", id)
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
}
