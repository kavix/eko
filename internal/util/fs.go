package util

import (
	"io"
	"os"
	"path/filepath"
)

// CopyDir copies all files from src to dst, skipping any directory named ".vibe".
func CopyDir(src string, dst string) error {
	// Resolve dst to an absolute path so we can skip it during the walk.
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .vibe directories entirely (prevents infinite recursion when
		// the snapshot destination lives inside the project tree).
		if info.IsDir() && filepath.Base(path) == ".vibe" {
			return filepath.SkipDir
		}

		// Skip the destination directory itself to avoid recursive copies.
		absPath, _ := filepath.Abs(path)
		if info.IsDir() && absPath == absDst {
			return filepath.SkipDir
		}

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
