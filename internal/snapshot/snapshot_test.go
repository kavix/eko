package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

// chdir changes the working directory for the duration of the test.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
}

// setupProject creates a temp project directory with a .vibe/snapshots subtree
// and some source files, then chdirs into it.
func setupProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".vibe", "snapshots"), 0755)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# project"), 0644)
	chdir(t, dir)
	return dir
}

// ---------------------------------------------------------------------------
// generateID
// ---------------------------------------------------------------------------

// TestGenerateID_length checks the hex string is 8 characters (4 bytes → 8 hex).
func TestGenerateID_length(t *testing.T) {
	id := generateID()
	if len(id) != 8 {
		t.Errorf("expected 8-char hex id, got %q (len=%d)", id, len(id))
	}
}

// TestGenerateID_hex checks the result is valid hex.
func TestGenerateID_hex(t *testing.T) {
	id := generateID()
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("non-hex character %q in id %q", c, id)
		}
	}
}

// TestGenerateID_unique checks two consecutive calls produce different IDs.
func TestGenerateID_unique(t *testing.T) {
	a, b := generateID(), generateID()
	if a == b {
		t.Errorf("generateID produced the same value twice: %q", a)
	}
}

// ---------------------------------------------------------------------------
// CreateSnapshot
// ---------------------------------------------------------------------------

// TestCreateSnapshot_returnsID verifies a non-empty ID is returned.
func TestCreateSnapshot_returnsID(t *testing.T) {
	setupProject(t)

	id, _, err := CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty snapshot ID")
	}
}

// TestCreateSnapshot_dirExists checks the snapshot directory was created.
func TestCreateSnapshot_dirExists(t *testing.T) {
	dir := setupProject(t)

	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot error: %v", err)
	}

	// snapPath is relative; resolve against the project dir.
	full := filepath.Join(dir, snapPath)
	if _, err := os.Stat(full); os.IsNotExist(err) {
		t.Errorf("snapshot directory %s does not exist", full)
	}
}

// TestCreateSnapshot_filesAreCopied checks that project files appear in snapshot.
func TestCreateSnapshot_filesAreCopied(t *testing.T) {
	dir := setupProject(t)

	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot error: %v", err)
	}

	full := filepath.Join(dir, snapPath)
	for _, f := range []string{"main.go", "README.md"} {
		if _, err := os.Stat(filepath.Join(full, f)); os.IsNotExist(err) {
			t.Errorf("expected %s to be in snapshot, but not found", f)
		}
	}
}

// TestCreateSnapshot_vibeNotCopied ensures .vibe is excluded from the snapshot.
func TestCreateSnapshot_vibeNotCopied(t *testing.T) {
	dir := setupProject(t)

	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot error: %v", err)
	}

	full := filepath.Join(dir, snapPath)
	vibeInSnap := filepath.Join(full, ".vibe")
	if _, err := os.Stat(vibeInSnap); !os.IsNotExist(err) {
		t.Error(".vibe directory should not be copied into snapshot")
	}
}

// ---------------------------------------------------------------------------
// RestoreSnapshot
// ---------------------------------------------------------------------------

// TestRestoreSnapshot_restoresFiles creates a snapshot then restores it and
// verifies that the original files come back.
func TestRestoreSnapshot_restoresFiles(t *testing.T) {
	dir := setupProject(t)

	// Save a snapshot of the initial state (contains main.go, README.md).
	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot error: %v", err)
	}

	// Delete a file to simulate a change.
	os.Remove(filepath.Join(dir, "README.md"))

	// Restore.
	full := filepath.Join(dir, snapPath)
	if err := RestoreSnapshot(full); err != nil {
		t.Fatalf("RestoreSnapshot error: %v", err)
	}

	// README.md should be back.
	if _, err := os.Stat(filepath.Join(dir, "README.md")); os.IsNotExist(err) {
		t.Error("README.md should have been restored but is missing")
	}
}

// TestRestoreSnapshot_removesExtraFiles verifies that files NOT in the snapshot
// are deleted from the working directory on restore.
func TestRestoreSnapshot_removesExtraFiles(t *testing.T) {
	dir := setupProject(t)

	// Snapshot the clean state.
	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	// Add a new file after the snapshot.
	extraFile := filepath.Join(dir, "extra.go")
	os.WriteFile(extraFile, []byte("package main"), 0644)

	// Restore to the snapshot — extra.go should be gone.
	full := filepath.Join(dir, snapPath)
	if err := RestoreSnapshot(full); err != nil {
		t.Fatalf("RestoreSnapshot error: %v", err)
	}

	if _, err := os.Stat(extraFile); !os.IsNotExist(err) {
		t.Error("extra.go should have been removed by RestoreSnapshot")
	}
}

// TestRestoreSnapshot_preservesVibe checks that .vibe is never removed during restore.
func TestRestoreSnapshot_preservesVibe(t *testing.T) {
	dir := setupProject(t)

	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	full := filepath.Join(dir, snapPath)
	if err := RestoreSnapshot(full); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".vibe")); os.IsNotExist(err) {
		t.Error(".vibe directory should be preserved after RestoreSnapshot")
	}
}
