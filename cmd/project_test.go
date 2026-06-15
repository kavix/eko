package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestIsInitialized(t *testing.T) {
	t.Run("missing .eko", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if isInitialized() {
			t.Fatal("expected an empty directory to be uninitialized")
		}
	})

	t.Run("with .eko", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		if err := os.Mkdir(ekoDir, 0755); err != nil {
			t.Fatal(err)
		}
		if !isInitialized() {
			t.Fatal("expected a directory with .eko to be initialized")
		}
	})

	t.Run("a file named .eko does not count", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		if err := os.WriteFile(ekoDir, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		if isInitialized() {
			t.Fatal("a plain file named .eko must not be treated as a project")
		}
	})
}

func TestRequireInitialized(t *testing.T) {
	t.Run("uninitialized returns a helpful error", func(t *testing.T) {
		t.Chdir(t.TempDir())
		err := requireInitialized(nil, nil)
		if err == nil {
			t.Fatal("expected an error in an uninitialized directory")
		}
		if !strings.Contains(err.Error(), "eko init") {
			t.Errorf("error should suggest 'eko init', got: %v", err)
		}
	})

	t.Run("initialized returns nil", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if err := os.Mkdir(ekoDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := requireInitialized(nil, nil); err != nil {
			t.Errorf("expected no error in an initialized directory, got: %v", err)
		}
	})
}
