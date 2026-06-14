package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"eko/internal/snapshot"
)

// SnapshotRecord is the JSON-serialisable form of a stored snapshot.
type SnapshotRecord struct {
	ID          string `json:"id"`
	Message     string `json:"message"`
	Path        string `json:"path"`
	CreatedAt   string `json:"createdAt"`
	FilesChanged int   `json:"filesChanged"`
}

// DiffFile holds the before/after content of a single changed file.
type DiffFile struct {
	Name     string `json:"name"`
	Original string `json:"original"`
	Modified string `json:"modified"`
}

// Serve starts the HTTP API on the given addr (e.g. ":7700").
func Serve(addr string, db *sql.DB) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/snapshots", cors(func(w http.ResponseWriter, r *http.Request) {
		listSnapshots(w, r, db)
	}))

	mux.HandleFunc("/api/snapshots/", cors(func(w http.ResponseWriter, r *http.Request) {
		// Route: /api/snapshots/{id}  or  /api/snapshots/{id}/restore
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/snapshots/"), "/")
		id := parts[0]
		if len(parts) == 2 && parts[1] == "restore" {
			restoreSnapshot(w, r, db, id)
			return
		}
		getSnapshot(w, r, db, id)
	}))

	mux.HandleFunc("/api/diff", cors(func(w http.ResponseWriter, r *http.Request) {
		diffSnapshots(w, r, db)
	}))

	fmt.Printf("eko API listening on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func listSnapshots(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.Query("SELECT id, message, path, created_at FROM snapshots ORDER BY created_at DESC")
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []SnapshotRecord
	for rows.Next() {
		var rec SnapshotRecord
		if err := rows.Scan(&rec.ID, &rec.Message, &rec.Path, &rec.CreatedAt); err != nil {
			continue
		}
		rec.FilesChanged = countFiles(rec.Path)
		records = append(records, rec)
	}
	if records == nil {
		records = []SnapshotRecord{}
	}
	jsonOK(w, records)
}

func getSnapshot(w http.ResponseWriter, r *http.Request, db *sql.DB, id string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var rec SnapshotRecord
	err := db.QueryRow(
		"SELECT id, message, path, created_at FROM snapshots WHERE id = ?", id,
	).Scan(&rec.ID, &rec.Message, &rec.Path, &rec.CreatedAt)
	if err == sql.ErrNoRows {
		jsonError(w, "snapshot not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rec.FilesChanged = countFiles(rec.Path)
	jsonOK(w, rec)
}

func restoreSnapshot(w http.ResponseWriter, r *http.Request, db *sql.DB, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var path string
	err := db.QueryRow("SELECT path FROM snapshots WHERE id = ?", id).Scan(&path)
	if err == sql.ErrNoRows {
		jsonError(w, "snapshot not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := snapshot.RestoreSnapshot(path); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "restored", "id": id})
}

func diffSnapshots(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fromID := r.URL.Query().Get("from")
	toID := r.URL.Query().Get("to")

	fromPath, toPath := "", ""
	db.QueryRow("SELECT path FROM snapshots WHERE id = ?", fromID).Scan(&fromPath)
	db.QueryRow("SELECT path FROM snapshots WHERE id = ?", toID).Scan(&toPath)

	diffs, err := buildDiff(fromPath, toPath)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, diffs)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildDiff walks both snapshot dirs, collecting file pairs where content differs.
func buildDiff(fromDir, toDir string) ([]DiffFile, error) {
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

func countFiles(dir string) int {
	count := 0
	filepath.Walk(dir, func(_ string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func cors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
