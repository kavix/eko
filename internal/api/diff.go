package api

import (
	"io/fs"
	"os"
	"path/filepath"
)

// SnapshotRecord is the JSON-serializable form of a stored snapshot.
type SnapshotRecord struct {
	ID           string `json:"id"`
	Message      string `json:"message"`
	Path         string `json:"path"`
	CreatedAt    string `json:"createdAt"`
	FilesChanged int    `json:"filesChanged"`
}

// DiffFile holds the before/after content of a single changed file.
type DiffFile struct {
	Name     string `json:"name"`
	Original string `json:"original"`
	Modified string `json:"modified"`
}

// BuildDiff walks both snapshot dirs, collecting file pairs where content differs.
func BuildDiff(fromDir, toDir string) ([]DiffFile, error) {
	seen := map[string]bool{}
	var results []DiffFile

	walkDir := func(root string) error {
		return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(root, path)
			seen[rel] = true
			return nil
		})
	}
	if fromDir != "" {
		walkDir(fromDir)
	}
	if toDir != "" {
		walkDir(toDir)
	}

	for rel := range seen {
		orig := readFileSafe(filepath.Join(fromDir, rel))
		mod := readFileSafe(filepath.Join(toDir, rel))
		if orig == mod {
			continue
		}
		results = append(results, DiffFile{
			Name:     rel,
			Original: orig,
			Modified: mod,
		})
	}
	return results, nil
}

func readFileSafe(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}
