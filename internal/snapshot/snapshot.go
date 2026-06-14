package snapshot

import (
	"crypto/rand"
	"encoding/hex"
	"os"

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

func RestoreSnapshot(path string) error {
	entries, _ := os.ReadDir(".")
	for _, e := range entries {
		if e.Name() == ".vibe" {
			continue
		}
		os.RemoveAll(e.Name())
	}
	return util.CopyDir(path, ".")
}
