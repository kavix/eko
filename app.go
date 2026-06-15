//go:build !no_gui

package main

import (
	"context"
	"database/sql"
	"eko/internal/api"
	"eko/internal/snapshot"
	"fmt"
	"os"
	"path/filepath"
)

// FrontendSnapshot represents the snapshot data structure required by the UI.
type FrontendSnapshot struct {
	ID           string   `json:"id"`
	Timestamp    string   `json:"timestamp"`
	Prompt       string   `json:"prompt"`
	AISummary    string   `json:"aiSummary"`
	FilesChanged []string `json:"filesChanged"`
	Model        string   `json:"model"`
}

// WailsApp handles Wails UI bindings.
type WailsApp struct {
	ctx context.Context
	DB  *sql.DB
}

// Startup is called when the app starts.
func (a *WailsApp) Startup(ctx context.Context) {
	a.ctx = ctx
}

// ListSnapshots returns all snapshots mapped to FrontendSnapshot format.
func (a *WailsApp) ListSnapshots() ([]FrontendSnapshot, error) {
	rows, err := a.DB.Query("SELECT id, message, path, created_at FROM snapshots ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []FrontendSnapshot
	for rows.Next() {
		var id, message, path, createdAt string
		if err := rows.Scan(&id, &message, &path, &createdAt); err != nil {
			continue
		}

		// Calculate actual files changed by diffing with predecessor
		filesChanged, _ := a.getSnapshotChanges(id, path, createdAt)

		prompt := message
		if prompt == "snapshot" {
			prompt = "Saved Snapshot #" + id
		}

		snapshots = append(snapshots, FrontendSnapshot{
			ID:           id,
			Timestamp:    createdAt,
			Prompt:       prompt,
			AISummary:    fmt.Sprintf("Snapshot of the repository state. %d files changed.", len(filesChanged)),
			FilesChanged: filesChanged,
			Model:        "Eko CLI",
		})
	}

	if snapshots == nil {
		snapshots = []FrontendSnapshot{}
	}
	return snapshots, nil
}

// GetSnapshot returns metadata of a single snapshot.
func (a *WailsApp) GetSnapshot(id string) (*api.SnapshotRecord, error) {
	var rec api.SnapshotRecord
	err := a.DB.QueryRow(
		"SELECT id, message, path, created_at FROM snapshots WHERE id = ?", id,
	).Scan(&rec.ID, &rec.Message, &rec.Path, &rec.CreatedAt)
	if err != nil {
		return nil, err
	}
	
	// Count files in the snapshot directory
	count := 0
	filepath.Walk(rec.Path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	rec.FilesChanged = count
	return &rec, nil
}

// RestoreSnapshot restores the workspace to the specified snapshot state.
func (a *WailsApp) RestoreSnapshot(id string) (map[string]string, error) {
	var path string
	err := a.DB.QueryRow("SELECT path FROM snapshots WHERE id = ?", id).Scan(&path)
	if err != nil {
		return nil, err
	}
	if err := snapshot.RestoreSnapshot(path); err != nil {
		return nil, err
	}
	return map[string]string{"status": "restored", "id": id}, nil
}

// DiffSnapshots returns differences between two snapshots.
func (a *WailsApp) DiffSnapshots(fromID, toID string) ([]api.DiffFile, error) {
	fromPath, toPath := "", ""
	a.DB.QueryRow("SELECT path FROM snapshots WHERE id = ?", fromID).Scan(&fromPath)
	a.DB.QueryRow("SELECT path FROM snapshots WHERE id = ?", toID).Scan(&toPath)

	// We call buildDiff helper inside internal/api
	// Wait, is buildDiff exported in package api?
	// Ah! Let's check internal/api/server.go.
	// In server.go: "func buildDiff(fromDir, toDir string) ([]DiffFile, error)"
	// Oh! It is lowercase buildDiff! It is private/unexported!
	// So package main cannot call it unless we export it!
	// Let's check if we should make buildDiff exported as BuildDiff in package api!
	// Yes! That's a simple, clean, and safe change.
	return api.BuildDiff(fromPath, toPath)
}

// getSnapshotChanges diffs current snapshot directory with its predecessor to find exactly which files changed.
func (a *WailsApp) getSnapshotChanges(id, currentPath, createdAt string) ([]string, error) {
	var prevPath string
	err := a.DB.QueryRow(
		"SELECT path FROM snapshots WHERE created_at < ? ORDER BY created_at DESC LIMIT 1",
		createdAt,
	).Scan(&prevPath)

	if err == sql.ErrNoRows {
		var files []string
		err = filepath.Walk(currentPath, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				rel, _ := filepath.Rel(currentPath, path)
				files = append(files, rel)
			}
			return nil
		})
		return files, err
	} else if err != nil {
		return nil, err
	}

	diffs, err := api.BuildDiff(prevPath, currentPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, d := range diffs {
		files = append(files, d.Name)
	}
	return files, nil
}
