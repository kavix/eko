package cmd

import (
	"bufio"
	"eko/internal/db"
	"eko/internal/snapshot"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var restoreYes bool

// errRestoreNeedsTTY is returned when confirmation is required but there is no
// terminal to ask on. Failing is the only safe option here: assuming "yes" would
// delete the working directory of a script that never opted in, and blocking on a
// read would hang a CI job until it times out.
var errRestoreNeedsTTY = errors.New(
	"restore needs confirmation but input is not a terminal; re-run with --yes to confirm")

var restoreCmd = &cobra.Command{
	Use:     "restore [id]",
	Short:   "Restore snapshot",
	Args:    cobra.ExactArgs(1),
	PreRunE: requireInitialized,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		database := db.InitDB()
		defer database.Close()

		var path string
		if err := database.QueryRow("SELECT path FROM snapshots WHERE id=?", id).Scan(&path); err != nil {
			return fmt.Errorf("snapshot not found: %w", err)
		}

		if !restoreYes {
			confirmed, err := confirmRestore(cmd, id)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintln(cmd.OutOrStdout(), "Restore cancelled. Nothing was deleted.")
				return nil
			}
		}

		err := snapshot.RestoreSnapshot(path)
		if err != nil {
			return err
		}
		fmt.Println("Restored:", id)

		return nil
	},
}

// confirmRestore lists what restore is about to delete and waits for an explicit
// "y". Anything else — including a bare Enter — cancels.
func confirmRestore(cmd *cobra.Command, id string) (bool, error) {
	in := cmd.InOrStdin()
	if f, ok := in.(*os.File); ok {
		info, statErr := f.Stat()
		if statErr != nil {
			return false, fmt.Errorf("cannot inspect input stream: %w", statErr)
		}
		if info.Mode()&os.ModeCharDevice == 0 {
			return false, errRestoreNeedsTTY
		}
	}

	toRemove, err := snapshot.PendingRemovals()
	if err != nil {
		return false, fmt.Errorf("cannot determine what restore would delete: %w", err)
	}

	out := cmd.OutOrStdout()
	if len(toRemove) == 0 {
		fmt.Fprintf(out, "Restoring snapshot %s. Nothing in the working directory will be deleted.\n", id)
	} else {
		fmt.Fprintf(out,
			"Restoring snapshot %s will permanently delete %s from the working directory:\n",
			id, pluralEntries(len(toRemove)))
		for _, name := range toRemove {
			fmt.Fprintf(out, "  %s\n", name)
		}
		fmt.Fprintln(out, "Any changes made since your last save will be lost.")
	}
	fmt.Fprint(out, "Continue? [y/N]: ")

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("cannot read confirmation: %w", err)
	}

	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func pluralEntries(n int) string {
	if n == 1 {
		return "1 entry"
	}
	return fmt.Sprintf("%d entries", n)
}

func init() {
	restoreCmd.Flags().BoolVarP(&restoreYes, "yes", "y", false,
		"skip the confirmation prompt (required when stdin is not a terminal)")
	rootCmd.AddCommand(restoreCmd)
}
