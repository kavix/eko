package util

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCopyDir_basic copies a flat source dir and checks all files arrive.
func TestCopyDir_basic(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	files := map[string]string{
		"hello.txt": "hello world",
		"data.go":   "package main",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(src, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir returned error: %v", err)
	}

	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Errorf("missing file %s after copy: %v", name, err)
			continue
		}
		if string(got) != want {
			t.Errorf("file %s: got %q, want %q", name, got, want)
		}
	}
}

// TestCopyDir_nested verifies subdirectory trees are replicated.
func TestCopyDir_nested(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	subdir := filepath.Join(src, "sub", "deep")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "file.txt"), []byte("deep"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir returned error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dst, "sub", "deep", "file.txt"))
	if err != nil {
		t.Fatalf("nested file not found: %v", err)
	}
	if string(got) != "deep" {
		t.Errorf("content mismatch: got %q, want \"deep\"", got)
	}
}

// TestCopyDir_skipsVibe ensures the .vibe directory is never copied.
func TestCopyDir_skipsVibe(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create a .vibe directory inside src — it should NOT appear in dst.
	if err := os.MkdirAll(filepath.Join(src, ".vibe"), 0755); err != nil {
		t.Fatal(err)
	}
	// Also add a normal file that SHOULD be copied.
	if err := os.WriteFile(filepath.Join(src, "real.txt"), []byte("real"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, ".vibe")); !os.IsNotExist(err) {
		t.Error(".vibe directory should not be copied to dst")
	}
	if _, err := os.Stat(filepath.Join(dst, "real.txt")); err != nil {
		t.Errorf("real.txt should have been copied: %v", err)
	}
}

// TestCopyDir_emptySource copies an empty directory without errors.
func TestCopyDir_emptySource(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("unexpected error on empty source: %v", err)
	}
}

// TestCopyDir_preservesContent checks file bytes are identical after copy.
func TestCopyDir_preservesContent(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	payload := []byte("binary\x00data\xff\xfe")
	if err := os.WriteFile(filepath.Join(src, "bin"), payload, 0644); err != nil {
		t.Fatal(err)
	}

	if err := CopyDir(src, dst); err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(filepath.Join(dst, "bin"))
	if string(got) != string(payload) {
		t.Errorf("binary content mismatch after copy")
	}
}
