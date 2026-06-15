package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "eko",
	Short: "eko – AI Snapshot Versioning CLI",
	// Errors (including the "not an eko project" guard) are reported once by
	// Execute below; don't let Cobra also print them or dump usage on failure.
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
