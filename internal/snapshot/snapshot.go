// Package snapshot implements the core snapshot creation and restoration logic for Eko.
// Snapshots are stored under .eko/snapshots/<id>/ relative to the project root.
package snapshot

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"eko/internal/util"
)

// ekoDir is the hidden directory where Eko stores its data.
// All snapshots live under ekoDir/snapshots/.
const ekoDir = ".eko"

// CreateSnapshot captures the current state of the working directory into a new snapshot.
// It generates a random 8-hex-char ID, copies the project tree (excluding .eko itself)
// into .eko/snapshots/<id>/, and returns the snapshot ID and its storage path.
func CreateSnapshot() (string, string, error) {
	id, err := generateID()
	if err != nil {
		return "", "", err
	}

	base := ekoDir + "/snapshots/" + id
	if err := util.CopyDir(".", base); err != nil {
		return "", "", discardPartial(base, err)
	}
	if err := captureEnvVars(base); err != nil {
		return "", "", discardPartial(base, err)
	}
	return id, base, nil
}

// discardPartial removes a snapshot directory that was created but never completed, and
// returns the error that caused the abort.
//
// Without this, a CreateSnapshot that fails after util.CopyDir has begun leaves the
// partial tree under .eko/snapshots/<id>/ with nothing pointing at it: the id and path
// are not returned to the caller, so nothing can reference or remove it, and `eko clean`
// only ever walks snapshots recorded in the database (#93).
//
// A cleanup failure is joined to the original error rather than replacing it. The reason
// the snapshot aborted is what the user needs; that a leftover directory could not be
// removed is extra, and swallowing either one loses information.
func discardPartial(base string, cause error) error {
	if rmErr := os.RemoveAll(base); rmErr != nil {
		return errors.Join(cause, fmt.Errorf("could not remove partial snapshot %s: %w", base, rmErr))
	}
	return cause
}

// generateID returns a random 8-character hexadecimal string used as a snapshot identifier.
func generateID() (string, error) {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// RestoreSnapshot reverts the working directory to the state captured in path.
//
// The restoration happens in two phases:
//
//  1. (Parallel delete) Every top-level entry in "." except the .eko directory is
//     removed concurrently. The first removal error is captured and returned after
//     all goroutines finish, so the working tree is never left in a half-deleted state
//     while an error is silently swallowed.
//
//  2. (Parallel copy) util.CopyDir copies the snapshot tree back into ".", also using
//     internal concurrency for large directory trees.
func RestoreSnapshot(path string) error {
	entries, err := os.ReadDir(".")
	if err != nil {
		return err
	}

	// Collect top-level entries that should be removed.
	// We always keep the .eko directory and other ignored folders/files so metadata/dependencies are preserved.
	var toRemove []string
	for _, e := range entries {
		if !util.ShouldIgnore(e.Name(), e.IsDir()) {
			toRemove = append(toRemove, e.Name())
		}
	}

	// Phase 1: delete concurrently; capture the first error via atomic pointer swap.
	// atomic.Pointer[error] gives the same lock-free compare-and-swap as a raw
	// unsafe.Pointer, but is type-safe and needs no unsafe import.
	var (
		wg       sync.WaitGroup
		firstErr atomic.Pointer[error] // nil means no error has been stored yet
	)
	for _, name := range toRemove {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			if rmErr := os.RemoveAll(n); rmErr != nil {
				// Only record the very first removal error encountered.
				firstErr.CompareAndSwap(nil, &rmErr)
			}
		}(name)
	}
	wg.Wait()

	// If any removal failed, return that error before attempting to copy.
	if err := firstErr.Load(); err != nil {
		return *err
	}

	// Phase 2: copy the snapshot back into the working directory.
	if err := util.CopyDir(path, "."); err != nil {
		return err
	}

	// Restore environment variables by writing a shell script.
	return restoreEnvVars(path)
}

func captureEnvVars(destDir string) error {
	env := os.Environ()
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	data, err := json.MarshalIndent(envMap, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(destDir, ".eko_env_vars.json"), data, 0600)
}

func restoreEnvVars(snapDir string) error {
	var envMap map[string]string
	data, err := os.ReadFile(filepath.Join(snapDir, ".eko_env_vars.json"))
	if err != nil {
		// If the file doesn't exist, it might be an older snapshot. Just skip.
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &envMap); err != nil {
		return err
	}

	// Create .eko_env_restore.sh
	f, err := os.OpenFile(".eko_env_restore.sh", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header
	if _, err := f.WriteString("#!/bin/sh\n# Eko Shell Environment Restore Script\n# Run: source .eko_env_restore.sh\n\n"); err != nil {
		return err
	}

	for k, v := range envMap {
		// Escape values for single-quoted strings in shell
		escapedVal := strings.ReplaceAll(v, "'", "'\\''")
		line := "export " + k + "='" + escapedVal + "'\n"
		if _, err := f.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}
