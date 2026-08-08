package cmd

import (
	"context"
	"eko/internal/ai"
	"eko/internal/db"
	"eko/internal/snapshot"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	saveMessage string
	saveAI      bool
	saveAIProv  string
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

  # Save with a custom message
  eko save -m "fixed db lock issue"

  # Save and auto-generate an AI change summary
  eko save --ai

  # Save with AI summary using a specific provider
  eko save --ai --provider heuristic`,
	PreRunE: requireInitialized,
	RunE: func(cmd *cobra.Command, args []string) error {
		database := db.InitDB()
		defer database.Close()

		// Get previous snapshot path before creating a new one
		var prevPath string
		_ = database.QueryRow("SELECT path FROM snapshots ORDER BY created_at DESC, rowid DESC LIMIT 1").Scan(&prevPath)

		id, path, err := snapshot.CreateSnapshot()
		if err != nil {
			return err
		}

		var summaryText string
		if saveAI {
			ctx := context.Background()
			res, err := ai.GenerateSnapshotSummary(ctx, prevPath, path, saveAIProv)
			if err == nil && res != nil {
				summaryText = res.Summary
				if saveMessage == "snapshot" {
					saveMessage = res.Summary
				}
			}
		}

		if _, err := database.Exec(
			"INSERT INTO snapshots(id, message, path, summary) VALUES (?, ?, ?, ?)",
			id,
			saveMessage,
			path,
			summaryText,
		); err != nil {
			// The snapshot tree is already on disk at this point, and the row that would
			// have made it reachable does not exist. Leaving it there strands the bytes:
			// `eko history` cannot list it, `eko restore <id>` fails with "snapshot not
			// found", and `eko clean` only walks snapshots recorded in the database, so it
			// never sees an orphan directory either. Every retry against a still-unwritable
			// database would add another one (#93).
			//
			// Removing it here keeps `save` all-or-nothing without changing the existing
			// write-then-insert order, so a partially-failed save leaves behind exactly what
			// it did before this fix: nothing.
			dbErr := fmt.Errorf("failed to save snapshot to db: %w", err)
			if rmErr := os.RemoveAll(path); rmErr != nil {
				return errors.Join(dbErr, fmt.Errorf("could not remove orphaned snapshot %s: %w", path, rmErr))
			}
			return dbErr
		}

		fmt.Println("Snapshot saved:", id)
		if summaryText != "" {
			fmt.Println("AI Summary:", summaryText)
		}

		return nil
	},
}

func init() {
	saveCmd.Flags().StringVarP(&saveMessage, "message", "m", "snapshot", "log message describing the snapshot")
	saveCmd.Flags().BoolVarP(&saveAI, "ai", "a", false, "auto-generate AI summary of changes")
	saveCmd.Flags().StringVar(&saveAIProv, "provider", "auto", "AI provider for auto-generated summary (auto, heuristic, openai, gemini)")
	rootCmd.AddCommand(saveCmd)
}
