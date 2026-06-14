// Package snapshot implements the core snapshot creation and restoration logic for Eko.
// Snapshots are stored under .eko/snapshots/<id>/ relative to the project root.
package snapshot

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"sync"
	"sync/atomic"
	"unsafe"

	"eko/internal/util"
)

// ekoDir is the hidden directory where Eko stores its data.
// All snapshots live under ekoDir/snapshots/.
const ekoDir = ".eko"

// CreateSnapshot captures the current state of the working directory into a new snapshot.
// It generates a random 8-hex-char ID, copies the project tree (excluding .eko itself)
// into .eko/snapshots/<id>/, and returns the snapshot ID and its storage path.
func CreateSnapshot() (string, string, error) {
	id := generateID()
	base := ekoDir + "/snapshots/" + id
	err := util.CopyDir(".", base)
	if err != nil {
		return "", "", err
	}
	return id, base, nil
}

// generateID returns a random 8-character hexadecimal string used as a snapshot identifier.
func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
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
	// We always keep the .eko directory so snapshot metadata is preserved.
	var toRemove []string
	for _, e := range entries {
		if e.Name() != ekoDir {
			toRemove = append(toRemove, e.Name())
		}
	}

	// Phase 1: delete concurrently; capture the first error via atomic pointer swap.
	// Using unsafe.Pointer lets us do a lock-free compare-and-swap on an *error value.
	var (
		wg       sync.WaitGroup
		firstErr unsafe.Pointer // *error; nil means no error has been stored yet
	)
	for _, name := range toRemove {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			if rmErr := os.RemoveAll(n); rmErr != nil {
				// Only record the very first removal error encountered.
				errVal := rmErr
				atomic.CompareAndSwapPointer(&firstErr, nil, unsafe.Pointer(&errVal))
			}
		}(name)
	}
	wg.Wait()

	// If any removal failed, return that error before attempting to copy.
	if firstErr != nil {
		return *(*error)(firstErr)
	}

	// Phase 2: copy the snapshot back into the working directory.
	return util.CopyDir(path, ".")
}
