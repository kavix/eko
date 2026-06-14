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

func CreateSnapshot() (string, string, error) {
	id := generateID()
	base := ".vibe/snapshots/" + id
	err := util.CopyDir(".", base)
	if err != nil {
		return "", "", err
	}
	return id, base, nil
}

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// RestoreSnapshot restores the working directory to the state in path.
//
// Phase 1 (parallel): delete every top-level entry except ".vibe" concurrently.
// Phase 2 (parallel via CopyDir): copy the snapshot back into ".".
func RestoreSnapshot(path string) error {
	entries, err := os.ReadDir(".")
	if err != nil {
		return err
	}

	// Collect entries to remove (skip .vibe).
	var toRemove []string
	for _, e := range entries {
		if e.Name() != ".vibe" {
			toRemove = append(toRemove, e.Name())
		}
	}

	// Delete concurrently; capture the first error via atomic pointer swap.
	var (
		wg      sync.WaitGroup
		firstErr unsafe.Pointer // *error, nil == no error yet
	)
	for _, name := range toRemove {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			if rmErr := os.RemoveAll(n); rmErr != nil {
				// Only store the very first error.
				ep := (*error)(nil)
				_ = ep
				errVal := rmErr
				atomic.CompareAndSwapPointer(&firstErr, nil, unsafe.Pointer(&errVal))
			}
		}(name)
	}
	wg.Wait()

	if firstErr != nil {
		return *(*error)(firstErr)
	}

	// Phase 2: copy snapshot back (already parallel inside CopyDir).
	return util.CopyDir(path, ".")
}
