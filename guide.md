# Eko User Guide ✦

Eko is an AI snapshot versioning tool designed to capture, inspect, diff, and restore directory states. It can be run either as a lightweight command-line interface (CLI) or as a rich native desktop application.

---

## 1. Installation

Depending on your use case, you can compile Eko in one of two modes:

### Option A: Native Desktop App (Recommended)
This compiles Eko with the native Wails visual UI window.

1. Ensure you have the Wails CLI installed:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```
2. Build the application package:
   ```bash
   wails build
   ```
   *Note: On macOS, this creates a native app bundle at `build/bin/eko.app`. You can copy the executable inside `/build/bin/eko.app/Contents/MacOS/eko` to your system path (e.g., `/usr/local/bin/eko`) to use the `eko` command globally.*

### Option B: Lightweight CLI Only
This compiles only the command-line interface, skipping Wails dependencies. Ideal for servers, headless environments, or if you do not need the graphical interface.

1. Compile the binary with the `no_gui` build tag:
   ```bash
   go build -tags no_gui -o eko .
   ```
2. Copy the resulting `eko` binary to a folder in your `$PATH` (e.g., `/usr/local/bin/`).

---

## 2. Basic CLI Usage

To use Eko inside any project or directory:

### Step 1: Initialize Eko
Create the hidden SQLite database and backups folder (`.eko/`) in the target directory:
```bash
cd /path/to/your-project
eko init
```
*Output:* `Eko initialized.`

### Step 2: Save a Snapshot
Capture the current state of all files in the directory (excluding the `.eko` folder itself):
```bash
eko save
```
*Output:* `Snapshot saved: <id>` (where `<id>` is a unique 8-character hexadecimal identifier).

### Step 3: View Snapshot History
List all saved snapshots with their creation timestamps:
```bash
eko history
```
*Example Output:*
```text
3b7f2a1e 2026-06-15 13:28:45
8c9d1a2f 2026-06-15 13:30:10
```

### Step 4: Restore a Previous State
Revert all files in your directory concurrently to the exact state captured in a given snapshot:
```bash
eko restore <snapshot-id>
```
*Example:* `eko restore 3b7f2a1e`
*Output:* `Restored: 3b7f2a1e`

---

## 3. Graphical UI Usage (Wails Only)

If you compiled Eko using **Option A (Desktop App)**, you can launch the native visual memory timeline interface directly from your command line:

```bash
eko ui
```

### Visual UI Features:
- **Interactive Timeline**: Scroll through snapshots in chronological order.
- **Changed Files List**: Inspect exactly how many files and which specific paths were added, modified, or deleted in each snapshot.
- **Monaco Diff Viewer**: Click on any changed file to view side-by-side split diffs with full syntax highlighting.
- **Graphical Restore**: Click the **Restore** button on any snapshot in the timeline or details panel to revert your workspace instantly.
