// Package util provides filesystem helpers used across echo internals.
package util

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// defaultIgnorePrefixes lists path prefixes (relative to the project root)
// that are always excluded from snapshots.
var defaultIgnorePrefixes = []string{
	".vibe",
	".git",
}

// CollectFiles walks root and returns relative paths of all regular files
// that are not excluded by the default ignore list.
func CollectFiles(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Normalise to a slash-separated relative path.
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)

		if isIgnored(rel, d) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			files = append(files, rel)
		}
		return nil
	})

	return files, err
}

// CopyDir copies a directory tree from src to dst, creating dst if needed.
func CopyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// isIgnored returns true when a path should be excluded from snapshots.
func isIgnored(rel string, d fs.DirEntry) bool {
	for _, prefix := range defaultIgnorePrefixes {
		if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
			return true
		}
	}
	return false
}
