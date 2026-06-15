package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ekoDir is the marker directory that identifies an initialized eko project.
const ekoDir = ".eko"

// isInitialized reports whether the current directory is an initialized eko
// project (i.e. it contains a .eko directory).
func isInitialized() bool {
	info, err := os.Stat(ekoDir)
	return err == nil && info.IsDir()
}

// requireInitialized returns a helpful error when the current directory has not
// been initialized with `eko init`. It is wired into commands as a cobra
// PreRunE so they fail early with a clear message instead of panicking or
// silently doing nothing.
func requireInitialized(_ *cobra.Command, _ []string) error {
	if isInitialized() {
		return nil
	}
	return fmt.Errorf("not an eko project: no %s directory found here\nrun 'eko init' to initialize eko in this directory", ekoDir)
}
