package util

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// copyTask holds a single file-copy job dispatched to the worker pool.
type copyTask struct {
	src string
	dst string
}

// ShouldIgnore reports whether a file or directory should be ignored by Eko.
func ShouldIgnore(name string, isDir bool) bool {
	if isDir {
		switch name {
		case ".eko", ".git", "node_modules", ".next", "build", "out", "dist":
			return true
		}
	} else {
		if exePath, err := os.Executable(); err == nil {
			if name == filepath.Base(exePath) {
				return true
			}
		}
		if name == "eko" || name == "eko.exe" {
			return true
		}
	}
	return false
}

// CopyDir copies all files from src to dst concurrently, skipping ignored items.
//
// Strategy:
//  1. Walk the source tree serially to preserve the guarantee that parent
//     directories always exist before any worker tries to write into them.
//  2. Dispatch every file path to a fixed-size worker pool (NumCPU workers)
//     so multiple files are copied in parallel, saturating both CPU and I/O.
//  3. Collect the first error from any worker and return it after draining.
func CopyDir(src, dst string) error {
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	numWorkers := runtime.NumCPU()
	tasks := make(chan copyTask, numWorkers*2)
	errs := make(chan error, numWorkers)

	// Start worker pool.
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tasks {
				if err := copyFile(t.src, t.dst); err != nil {
					errs <- err
					return
				}
			}
		}()
	}

	// Walk serially: create dirs immediately, enqueue file copies.
	walkErr := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if ShouldIgnore(filepath.Base(path), info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip the destination directory itself.
		absPath, _ := filepath.Abs(path)
		if info.IsDir() && absPath == absDst {
			return filepath.SkipDir
		}

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			// Create directories synchronously so workers never race on mkdir.
			return os.MkdirAll(target, 0755)
		}

		// Non-blocking check: bail early if a worker already reported an error.
		select {
		case err := <-errs:
			errs <- err // put it back so the final drain sees it
			return err
		default:
		}

		tasks <- copyTask{src: path, dst: target}
		return nil
	})

	close(tasks)
	wg.Wait()
	close(errs)

	// Walk error takes priority; otherwise return the first worker error.
	if walkErr != nil {
		return walkErr
	}
	return <-errs // nil if the channel is empty
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
