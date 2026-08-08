package snapshot

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

// setupProject creates a temp project directory with a .eko/snapshots subtree
// and some source files, then chdirs into it.
func setupProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".eko", "snapshots"), 0755)
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
	id, err := generateID()
	if err != nil {
		t.Fatalf("generateID error: %v", err)
	}

	if len(id) != 8 {
		t.Errorf("expected 8-char hex id, got %q (len=%d)", id, len(id))
	}
}

// TestGenerateID_hex checks the result is valid hex.
func TestGenerateID_hex(t *testing.T) {
	id, err := generateID()
	if err != nil {
		t.Fatalf("generateID error: %v", err)
	}

	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("non-hex character %q in id %q", c, id)
		}
	}
}

// TestGenerateID_unique checks two consecutive calls produce different IDs.
func TestGenerateID_unique(t *testing.T) {
	a, err := generateID()
	if err != nil {
		t.Fatalf("generateID error for a: %v", err)
	}

	b, err := generateID()
	if err != nil {
		t.Fatalf("generateID error for b: %v", err)
	}

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

// TestCreateSnapshot_ekoNotCopied ensures .eko is excluded from the snapshot.
func TestCreateSnapshot_ekoNotCopied(t *testing.T) {
	dir := setupProject(t)

	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot error: %v", err)
	}

	full := filepath.Join(dir, snapPath)
	ekoInSnap := filepath.Join(full, ".eko")
	if _, err := os.Stat(ekoInSnap); !os.IsNotExist(err) {
		t.Error(".eko directory should not be copied into snapshot")
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

// TestRestoreSnapshot_preservesEko checks that .eko is never removed during restore.
func TestRestoreSnapshot_preservesEko(t *testing.T) {
	dir := setupProject(t)

	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	full := filepath.Join(dir, snapPath)
	if err := RestoreSnapshot(full); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".eko")); os.IsNotExist(err) {
		t.Error(".eko directory should be preserved after RestoreSnapshot")
	}
}

// TestCreateSnapshot_ignoresFiles ensures common files/directories like .git and node_modules are ignored.
func TestCreateSnapshot_ignoresFiles(t *testing.T) {
	dir := setupProject(t)

	// Create some folders/files that should be ignored
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("[core]"), 0644)
	os.MkdirAll(filepath.Join(dir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "some-dep"), []byte("some-dep"), 0644)
	os.WriteFile(filepath.Join(dir, "eko"), []byte("binary content"), 0755)

	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot error: %v", err)
	}

	full := filepath.Join(dir, snapPath)
	for _, ign := range []string{".git", "node_modules", "eko"} {
		if _, err := os.Stat(filepath.Join(full, ign)); !os.IsNotExist(err) {
			t.Errorf("%s should not be copied into snapshot", ign)
		}
	}
}

// TestRestoreSnapshot_preservesIgnored checks that ignored folders/files are preserved during restore.
func TestRestoreSnapshot_preservesIgnored(t *testing.T) {
	dir := setupProject(t)

	// Create some folders/files that should be ignored
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(dir, "eko"), []byte("binary content"), 0755)

	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	full := filepath.Join(dir, snapPath)
	if err := RestoreSnapshot(full); err != nil {
		t.Fatal(err)
	}

	for _, ign := range []string{".git", "node_modules", "eko"} {
		if _, err := os.Stat(filepath.Join(dir, ign)); os.IsNotExist(err) {
			t.Errorf("%s directory should be preserved after RestoreSnapshot", ign)
		}
	}
}

// TestRestoreSnapshot_restoresEnvVars verifies that environment variables are captured
// during CreateSnapshot and successfully restored to a .eko_env_restore.sh script.
func TestRestoreSnapshot_restoresEnvVars(t *testing.T) {
	dir := setupProject(t)

	// Set a custom environment variable
	os.Setenv("EKO_TEST_VAR", "eko-value-123")
	defer os.Unsetenv("EKO_TEST_VAR")

	// Snapshot the state.
	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	// Restore the state.
	full := filepath.Join(dir, snapPath)
	if err := RestoreSnapshot(full); err != nil {
		t.Fatalf("RestoreSnapshot error: %v", err)
	}

	// Check if .eko_env_restore.sh was created.
	restoreScript := filepath.Join(dir, ".eko_env_restore.sh")
	if _, err := os.Stat(restoreScript); os.IsNotExist(err) {
		t.Fatal(".eko_env_restore.sh was not created on restore")
	}

	// Read content and check if it contains the variable.
	content, err := os.ReadFile(restoreScript)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "export EKO_TEST_VAR='eko-value-123'") {
		t.Errorf("expected script to contain EKO_TEST_VAR export, got:\n%s", string(content))
	}
}

// TestRestoreSnapshot_returnsRemovalError verifies the parallel-deletion error
// path: when a top-level entry cannot be removed, RestoreSnapshot reports that
// error and returns *before* copying the snapshot back, so the working tree is
// not left in a half-restored state.
func TestRestoreSnapshot_returnsRemovalError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: directory permissions do not block removal")
	}

	dir := setupProject(t)

	_, snapPath, err := CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot error: %v", err)
	}
	full := filepath.Join(dir, snapPath)

	// A directory whose child cannot be unlinked because the parent is not
	// writable: os.RemoveAll on it fails with a permission error.
	blocked := filepath.Join(dir, "blocked")
	if err := os.Mkdir(blocked, 0755); err != nil {
		t.Fatalf("mkdir blocked: %v", err)
	}
	if err := os.WriteFile(filepath.Join(blocked, "child.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("write child: %v", err)
	}
	if err := os.Chmod(blocked, 0500); err != nil {
		t.Fatalf("chmod blocked: %v", err)
	}
	// Restore write permission so t.TempDir cleanup can remove the tree.
	t.Cleanup(func() { os.Chmod(blocked, 0700) })

	if err := RestoreSnapshot(full); err == nil {
		t.Fatal("RestoreSnapshot should have returned the removal error, got nil")
	}

	// Phase 2 must not have run: README.md was removed by phase 1 and, because
	// the error short-circuits the copy, it should not have been restored.
	if _, err := os.Stat(filepath.Join(dir, "README.md")); !os.IsNotExist(err) {
		t.Error("snapshot was copied back despite a removal error; expected early return")
	}
}

// ---------------------------------------------------------------------------
// discardPartial (#93)
// ---------------------------------------------------------------------------

// A CreateSnapshot that aborts after util.CopyDir has begun would otherwise leave the
// partial tree under .eko/snapshots/<id>/ with nothing pointing at it: the id and path
// are never returned, so no caller can reference or remove it, and `eko clean` only walks
// snapshots recorded in the database.
func TestDiscardPartial_removesTheDirectoryAndKeepsTheCause(t *testing.T) {
	setupProject(t)

	base := filepath.Join(".eko", "snapshots", "deadbeef")
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "partial.txt"), []byte("half a snapshot"), 0644); err != nil {
		t.Fatal(err)
	}

	cause := errors.New("copy failed midway")
	got := discardPartial(base, cause)

	if !errors.Is(got, cause) {
		t.Errorf("the abort cause must survive cleanup, got %v", got)
	}
	if _, err := os.Stat(base); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected %s to be removed, stat error was %v", base, err)
	}
}

// Cleanup must not invent a snapshot directory or fail when there is nothing to remove:
// CreateSnapshot can abort before CopyDir creates anything.
func TestDiscardPartial_isANoOpWhenNothingWasWritten(t *testing.T) {
	setupProject(t)

	cause := errors.New("aborted before any write")
	got := discardPartial(filepath.Join(".eko", "snapshots", "never-created"), cause)

	if !errors.Is(got, cause) {
		t.Errorf("expected the cause unchanged, got %v", got)
	}
	if got.Error() != cause.Error() {
		t.Errorf("a no-op cleanup must not decorate the error, got %q", got)
	}
}
