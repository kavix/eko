# Contributing to Eko ✦

Thank you for your interest in contributing to Eko! Eko is an AI-powered snapshot versioning tool that runs as both a lightweight CLI and a rich Wails desktop application. 

This guide will help you set up your development environment, understand our architecture, and walk you through adding new features.

---

## 1. Prerequisites

To develop Eko, you need:
- **Go 1.21+** ([Download](https://go.dev/dl/))
- **Node.js 18+** ([Download](https://nodejs.org/))
- **Wails v2 CLI** — required for packaging and local HMR dev server:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

---

## 2. Project Architecture

Eko is structured as a dual-build project to allow headless operation on CI/CD servers (like GitHub Actions) and native desktop UI execution on local systems:

```
eko/
├── cmd/              # Cobra CLI commands (init, save, history, ui, etc.)
│   ├── ui_desktop.go # UI subcommand (included in desktop builds)
│   └── ui_cli.go     # UI subcommand stub (included in headless builds)
├── internal/
│   ├── api/          # Internal diffing logic & structures
│   ├── db/           # Local SQLite database initialization
│   └── snapshot/     # Concurrent snapshot creation/restoration engine
├── ui/               # Next.js 16/React UI
│   ├── app/          # App Router components & layouts
│   ├── components/   # Visual UI components (DetailPanel, DiffViewer, etc.)
│   └── lib/wailsjs/  # Auto-generated Wails JavaScript/TypeScript bindings
├── entry_desktop.go  # Starts Wails desktop runtime (go:build !no_gui)
├── entry_cli.go      # Starts standard CLI runtime (go:build no_gui)
├── app.go            # Go-to-Javascript Wails binding declarations
└── main.go           # Entry point containing embedded UI assets
```

### Build Constraints (`no_gui` Tag)
- **Standard Mode (GUI + CLI):** Compiled by default (or with `wails build`). Includes the desktop WebKit dependencies.
- **Lightweight Mode (CLI Only):** Excludes Wails and native OS browser components. Compiled using `-tags no_gui`. Essential for headless Linux servers and CI runners.

---

## 3. Development Setup

1. **Fork and Clone the Repo:**
   ```bash
   git clone https://github.com/<your-username>/eko.git
   cd eko
   ```

2. **Run in Development Mode (with Live Reload):**
   Start the backend and frontend development server:
   ```bash
   wails dev
   ```
   *This launches the native desktop application and a local web server (usually at `http://localhost:3000`). Modifying React code or Go bindings will automatically trigger hot-reload.*

3. **Building the Production Binary Locally:**
   - **On macOS:**
     ```bash
     (cd ui && npm run build)
     CGO_LDFLAGS="-framework UniformTypeIdentifiers" go build -tags desktop,production -o eko .
     ```
   - **On Windows / Linux:**
     ```bash
     (cd ui && npm run build)
     go build -tags desktop,production -o eko .
     ```

---

## 4. How to Add a New Feature

Adding a new feature typically involves four layers: internal engine, Cobra CLI, Wails bindings, and the React UI.

### Step 1: Implement the Engine Logic
Write the core Go logic in the `internal/` package. Keep database operations inside `internal/db` and filesystem operations inside `internal/snapshot` or `internal/util`.

### Step 2: Register the Cobra CLI Command
Add a new command file in `cmd/` (e.g., `cmd/yourfeature.go`):
```go
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var yourFeatureCmd = &cobra.Command{
	Use:   "yourfeature",
	Short: "Description of your feature",
	Run: func(cmd *cobra.Command, args []string) {
		// Invoke internal engine logic here
		fmt.Println("Feature executed via CLI!")
	},
}

func init() {
	rootCmd.AddCommand(yourFeatureCmd)
}
```

### Step 3: Declare the Wails Binding
To make this feature available to the Desktop UI, add a public method to the `WailsApp` struct in `app.go`:
```go
// YourFeatureBinding is exposed to the frontend.
func (a *WailsApp) YourFeatureBinding(param string) (string, error) {
	// Call internal engine logic here
	return "Result of feature", nil
}
```
*Note: Public methods of `WailsApp` starting with uppercase letters are automatically scanned and exported by Wails.*

### Step 4: Wire Up the Frontend
1. **Regenerate JS Bindings:** During `wails dev` or `wails build`, Wails automatically outputs TypeScript bindings to `ui/lib/wailsjs/go/main/WailsApp.js`.
2. **Import in React:** Import the generated binding in your page or component (e.g., `ui/app/page.tsx`):
   ```typescript
   import { YourFeatureBinding } from "@/lib/wailsjs/go/main/WailsApp";
   ```
3. **Invoke and Render:** Call the binding asynchronously and manage the response using standard React state hooks.

---

## 5. Coding Standards & Conventions

### Go Guidelines
- **Paths:** Always resolve directories relative to the current working directory. Eko can be executed in any user folder.
- **SQLite Safety:** Database operations should handle errors gracefully. Avoid panic blocks in UI bindings to prevent app window crashes.
- **Formatting:** Format Go files using `gofmt` and run `go vet ./...` before submitting a PR.

### React / Next.js Guidelines
- **Static Compilation:** Next.js is configured for a static HTML export (`output: 'export'` in `next.config.ts`). Do not use server-side React features (SSR, Server Actions, Middleware).
- **External Imports:** Wrap any references to browser APIs (`window`, `localStorage`) inside a `useEffect` hook or check `typeof window !== 'undefined'` to avoid build-time compilation errors.
- **Tailwind styling**: Follow the unified look & feel (vibrant colors, glassmorphism, smooth animations) defined in the component stylesheets.

---

## 6. Submitting a Pull Request

1. Create a feature branch:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. Commit your modifications with a semantic commit message (e.g., `feat: add custom snapshot labels`, `fix: sqlite directory lock`).
3. Run automated tests to ensure everything is correct:
   ```bash
   go test ./...
   ```
4. Push to your fork and submit a PR to `kavix/eko:main`. Please include screenshots or screen recordings for any changes modifying the desktop UI layout.
