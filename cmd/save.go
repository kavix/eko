package cmd

import (
	"eko/internal/db"
	"eko/internal/snapshot"
	"fmt"

	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save project snapshot",
	Long: `Save creates a new snapshot of the current project state.

A snapshot captures all files in the project directory and stores them
for later retrieval. Each snapshot is assigned a unique ID that can be
used with the restore command to revert to this state.`,
	Example: `  # Save a snapshot of the current project state
  eko save

  # Save and immediately view history
  eko save && eko history

  # View history, then restore to a prior snapshot
  eko history
  eko restore <snapshot-id>`,
	PreRunE: requireInitialized,
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
