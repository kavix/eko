# Contributing to Eko

Thank you for your interest in contributing to Eko! This guide will help you get up and running quickly.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Setting Up Your Development Environment](#setting-up-your-development-environment)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Running Tests](#running-tests)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Code Style](#code-style)

## Prerequisites

Make sure you have the following installed:

- **Go 1.21+** — for the CLI and backend ([download](https://go.dev/dl/))
- **Node.js 18+** — for the web UI ([download](https://nodejs.org/))
- **Git**

## Setting Up Your Development Environment

1. **Fork** the repository and clone your fork:

   ```bash
   git clone https://github.com/<your-username>/eko.git
   cd eko
   ```

2. **Build the CLI:**

   ```bash
   go build -o eko main.go
   ```

3. **Start the web UI** (optional, for UI changes):

   ```bash
   cd ui
   npm install
   npm run dev
   ```

   The dev server starts at <http://localhost:3000>.

4. **Verify the setup:**

   ```bash
   ./eko --help
   ```

## Project Structure

```
eko/
├── cmd/           CLI commands (save, restore, history, init, ui)
├── internal/
│   ├── db/        SQLite database helpers
│   ├── snapshot/  Snapshot creation and restoration logic
│   └── util/      File system utilities (CopyDir, etc.)
├── ui/            Next.js web interface
│   └── app/       App Router pages and layouts
│   └── components/ React components (SnapshotCard, etc.)
└── main.go        Entry point
```

Snapshots are stored under `.eko/snapshots/<id>/` relative to your project root.

## Making Changes

1. Create a new branch from `main`:

   ```bash
   git checkout -b feat/your-feature-name
   ```

2. Make your changes following the [code style guidelines](#code-style).

3. Run tests to make sure nothing is broken (see [Running Tests](#running-tests)).

4. Commit with a clear message:

   ```bash
   git commit -m "feat: add confirmation prompt to restore command"
   ```

## Running Tests

```bash
# Run all Go tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific package
go test ./internal/snapshot/...
```

## Submitting a Pull Request

1. Push your branch to your fork:

   ```bash
   git push origin feat/your-feature-name
   ```

2. Open a pull request against `kavix/eko:main`.

3. In your PR description, include:
   - A summary of what changed and why
   - Reference to any related issue (`Closes #123`)
   - Screenshots for UI changes

4. A maintainer will review your PR and may request changes.

## Code Style

**Go:**
- Follow standard Go conventions (`gofmt`, `go vet`).
- Add godoc comments to all exported functions and packages.
- Keep functions small and focused.

**TypeScript / React (UI):**
- Use functional components with hooks.
- Follow the existing file structure under `ui/`.
- Use Tailwind CSS for styling (no inline styles).

---

If you have questions, open a [GitHub Discussion](https://github.com/kavix/eko/discussions) or comment on the relevant issue. We're happy to help!
