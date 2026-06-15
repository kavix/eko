//go:build !no_gui

package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"eko/cmd"
	"eko/internal/db"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func run() {
	// Only run Cobra CLI if the command is a known CLI subcommand.
	// Otherwise, default to starting the Wails UI (so that Wails bindings generation
	// and desktop double-clicking both launch Wails correctly).
	isCLI := false
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init", "save", "restore", "history", "completion", "help", "-h", "--help":
			isCLI = true
		}
	}

	if isCLI {
		cmd.UIAssets = assets
		cmd.Execute()
		return
	}

	// Default behavior (no args, or "ui") is to start Wails UI
	database := db.InitDB()
	wailsApp := &WailsApp{DB: database}

	fmt.Println("✦ Starting Eko Native UI...")

	subFS, err := fs.Sub(assets, "ui/out")
	if err != nil {
		fmt.Println("Error loading embedded UI assets:", err)
		os.Exit(1)
	}

	err = wails.Run(&options.App{
		Title:  "Eko Visual Memory",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: subFS,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Intercept Next.js App Router static prefetch requests
				if r.URL.Query().Has("_rsc") {
					path := r.URL.Path
					if path == "/" || path == "" || path == "/index.html" {
						content, err := fs.ReadFile(subFS, "index.txt")
						if err == nil {
							w.Header().Set("Content-Type", "text/plain; charset=utf-8")
							w.Write(content)
							return
						}
					} else if path == "/_not-found" || path == "/_not-found.html" {
						content, err := fs.ReadFile(subFS, "_not-found.txt")
						if err == nil {
							w.Header().Set("Content-Type", "text/plain; charset=utf-8")
							w.Write(content)
							return
						}
					}
				}
				w.WriteHeader(http.StatusNotFound)
			}),
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 1}, // Slate-900 style background
		OnStartup:        wailsApp.Startup,
		Bind: []interface{}{
			wailsApp,
		},
	})

	if err != nil {
		fmt.Println("Wails application error:", err)
		os.Exit(1)
	}
}
