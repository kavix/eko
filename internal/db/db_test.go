package db

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupEkoDir creates a temp project root with a .eko subdir,
// chdirs into it, and returns the cleanup restore.
func setupEkoDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".eko"), 0755); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
}

// TestInitDB_opens verifies InitDB returns a non-nil, usable connection.
func TestInitDB_opens(t *testing.T) {
	setupEkoDir(t)

	conn := InitDB()
	defer conn.Close()

	if conn == nil {
		t.Fatal("InitDB returned nil")
	}
	if err := conn.Ping(); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

// TestInitDB_createSchema verifies the snapshots table can be created and queried.
func TestInitDB_createSchema(t *testing.T) {
	setupEkoDir(t)

	conn := InitDB()
	defer conn.Close()

	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS snapshots (
			id TEXT PRIMARY KEY,
			message TEXT,
			path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	_, err = conn.Exec(
		"INSERT INTO snapshots(id, message, path) VALUES (?, ?, ?)",
		"abc123", "test snap", "/tmp/snap",
	)
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	var id, message, path string
	err = conn.QueryRow("SELECT id, message, path FROM snapshots WHERE id = ?", "abc123").
		Scan(&id, &message, &path)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}
	if id != "abc123" {
		t.Errorf("id: got %q, want %q", id, "abc123")
	}
	if message != "test snap" {
		t.Errorf("message: got %q, want %q", message, "test snap")
	}
	if path != "/tmp/snap" {
		t.Errorf("path: got %q, want %q", path, "/tmp/snap")
	}
}

// TestInitDB_multipleRows inserts several rows and verifies the count.
func TestInitDB_multipleRows(t *testing.T) {
	setupEkoDir(t)

	conn := InitDB()
	defer conn.Close()

	conn.Exec(`CREATE TABLE IF NOT EXISTS snapshots (
		id TEXT PRIMARY KEY, message TEXT, path TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)

	ids := []string{"a1", "b2", "c3"}
	for i, id := range ids {
		_, err := conn.Exec("INSERT INTO snapshots(id,message,path) VALUES(?,?,?)",
			id, "msg", "/p/"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("INSERT %q failed: %v", id, err)
		}
	}

	var count int
	conn.QueryRow("SELECT COUNT(*) FROM snapshots").Scan(&count)
	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}
}

// TestInitDB_uniqueConstraint checks the PRIMARY KEY rejects duplicates.
func TestInitDB_uniqueConstraint(t *testing.T) {
	setupEkoDir(t)

	conn := InitDB()
	defer conn.Close()

	conn.Exec(`CREATE TABLE IF NOT EXISTS snapshots (
		id TEXT PRIMARY KEY, message TEXT, path TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)

	conn.Exec("INSERT INTO snapshots(id,message,path) VALUES('dup','first','/a')")
	_, err := conn.Exec("INSERT INTO snapshots(id,message,path) VALUES('dup','second','/b')")
	if err == nil {
		t.Error("expected UNIQUE constraint error on duplicate id, got nil")
	}
}

// TestInitDB_fileCreated checks that the SQLite file is written to disk.
func TestInitDB_fileCreated(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".eko"), 0755)
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })
	os.Chdir(dir)

	conn := InitDB()
	conn.Exec("SELECT 1") // force file materialisation
	conn.Close()

	if _, err := os.Stat(filepath.Join(dir, ".eko", "db.sqlite")); os.IsNotExist(err) {
		t.Error(".eko/db.sqlite was not created on disk")
	}
}
