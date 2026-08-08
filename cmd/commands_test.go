package cmd

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"eko/internal/db"
	"eko/internal/snapshot"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDir creates a temp directory, changes to it, and registers a cleanup.
func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestInitCommand(t *testing.T) {
	dir := setupTestDir(t)

	// Run the init command
	_ = initCmd.RunE(initCmd, []string{})

	// Check that .eko/snapshots was created
	snapDir := filepath.Join(dir, ".eko", "snapshots")
	if info, err := os.Stat(snapDir); err != nil || !info.IsDir() {
		t.Fatalf("expected .eko/snapshots to be a directory, error: %v", err)
	}

	// Check that the database file was created
	dbFile := filepath.Join(dir, ".eko", "db.sqlite")
	if _, err := os.Stat(dbFile); err != nil {
		t.Fatalf("expected .eko/db.sqlite to exist, error: %v", err)
	}

	// Open database and verify snapshots table exists
	database, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	var name string
	err = database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='snapshots'").Scan(&name)
	if err != nil {
		t.Fatalf("expected snapshots table to exist: %v", err)
	}
}

func TestSaveCommand(t *testing.T) {
	dir := setupTestDir(t)

	// First initialize the project
	_ = initCmd.RunE(initCmd, []string{})

	// Create a dummy file to snapshot
	testFile := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run the save command
	if err := saveCmd.RunE(saveCmd, []string{}); err != nil {
		t.Fatal(err)
	}

	// Verify database record
	database := db.InitDB()
	defer database.Close()

	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM snapshots").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 snapshot record, got %d", count)
	}

	var id, message, path string
	err = database.QueryRow("SELECT id, message, path FROM snapshots LIMIT 1").Scan(&id, &message, &path)
	if err != nil {
		t.Fatal(err)
	}

	if id == "" {
		t.Error("expected non-empty snapshot ID")
	}
	if message != "snapshot" {
		t.Errorf("expected message 'snapshot', got %q", message)
	}

	// Verify files exist in the snapshot path
	snapFilePath := filepath.Join(dir, path, "hello.txt")
	if content, err := os.ReadFile(snapFilePath); err != nil || string(content) != "hello world" {
		t.Errorf("expected snapshot to contain 'hello world', got err=%v, content=%s", err, string(content))
	}
}

func TestHistoryCommand(t *testing.T) {
	dir := setupTestDir(t)

	// Initialize and create dummy file
	_ = initCmd.RunE(initCmd, []string{})
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Save snapshot
	_ = saveCmd.RunE(saveCmd, []string{})

	// Setup capture of stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = historyCmd.RunE(historyCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if len(output) == 0 {
		t.Error("expected history output to contain snapshot entries, got empty string")
	}
}

func TestHistoryCommand_jsonOutput(t *testing.T) {
	dir := setupTestDir(t)

	_ = initCmd.RunE(initCmd, []string{})
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	_ = saveCmd.RunE(saveCmd, []string{})

	if err := historyCmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		jsonOutput = false
		if err := historyCmd.Flags().Set("json", "false"); err != nil {
			t.Fatal(err)
		}
	}()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = historyCmd.RunE(historyCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var entries []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("expected valid JSON output, got %q: %v", buf.String(), err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(entries))
	}
	if entries[0]["id"] == "" {
		t.Error("expected JSON entry to include id")
	}
	if entries[0]["created_at"] == "" {
		t.Error("expected JSON entry to include created_at")
	}
}

func TestRestoreCommand(t *testing.T) {
	dir := setupTestDir(t)

	// Initialize and save initial snapshot with hello.txt
	_ = initCmd.RunE(initCmd, []string{})
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello version 1"), 0644); err != nil {
		t.Fatal(err)
	}
	_ = saveCmd.RunE(saveCmd, []string{})

	// Get snapshot ID
	database := db.InitDB()
	var id string
	err := database.QueryRow("SELECT id FROM snapshots LIMIT 1").Scan(&id)
	database.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Modify file
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello version 2"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run restore. Restore now confirms before deleting, so a non-interactive
	// caller has to opt in explicitly — the same thing a script would do.
	withRestoreYes(t, true)
	_ = restoreCmd.RunE(restoreCmd, []string{id})

	// Check file restored to version 1
	content, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello version 1" {
		t.Errorf("expected content to be restored to 'hello version 1', got %q", string(content))
	}
}

// withRestoreYes sets the --yes flag for one test and restores it afterwards, so
// the package-level flag cannot leak between tests.
func withRestoreYes(t *testing.T, v bool) {
	t.Helper()
	prev := restoreYes
	restoreYes = v
	t.Cleanup(func() { restoreYes = prev })
}

// seedSnapshot initialises a project, saves one snapshot of hello.txt, then dirties
// the file. Returns the snapshot id and the directory.
func seedSnapshot(t *testing.T) (string, string) {
	t.Helper()
	dir := setupTestDir(t)
	_ = initCmd.RunE(initCmd, []string{})
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("saved"), 0644); err != nil {
		t.Fatal(err)
	}
	_ = saveCmd.RunE(saveCmd, []string{})

	database := db.InitDB()
	var id string
	err := database.QueryRow("SELECT id FROM snapshots LIMIT 1").Scan(&id)
	database.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("unsaved work"), 0644); err != nil {
		t.Fatal(err)
	}
	return id, dir
}

// runRestoreWith drives restoreCmd with a scripted answer on its input stream.
func runRestoreWith(t *testing.T, id, answer string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	restoreCmd.SetOut(&out)
	restoreCmd.SetIn(strings.NewReader(answer))
	t.Cleanup(func() { restoreCmd.SetOut(nil); restoreCmd.SetIn(nil) })
	err := restoreCmd.RunE(restoreCmd, []string{id})
	return out.String(), err
}

// The point of issue #61: declining must leave the working directory untouched.
func TestRestoreCommand_declinedLeavesWorkingDirectoryIntact(t *testing.T) {
	withRestoreYes(t, false)
	id, dir := seedSnapshot(t)

	out, err := runRestoreWith(t, id, "n\n")
	if err != nil {
		t.Fatalf("declining should not be an error, got %v", err)
	}
	content, readErr := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "unsaved work" {
		t.Errorf("declining must not touch the file; got %q", string(content))
	}
	if !strings.Contains(out, "Nothing was deleted") {
		t.Errorf("expected a cancellation message, got %q", out)
	}
}

// A bare Enter is the default, and the default must be "no".
func TestRestoreCommand_emptyAnswerCancels(t *testing.T) {
	withRestoreYes(t, false)
	id, dir := seedSnapshot(t)

	if _, err := runRestoreWith(t, id, "\n"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, _ := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if string(content) != "unsaved work" {
		t.Errorf("a bare Enter must cancel; got %q", string(content))
	}
}

func TestRestoreCommand_confirmedRestores(t *testing.T) {
	withRestoreYes(t, false)
	id, dir := seedSnapshot(t)

	out, err := runRestoreWith(t, id, "y\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, _ := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if string(content) != "saved" {
		t.Errorf("confirming should restore; got %q", string(content))
	}
	// The prompt must name what it is about to delete, or it is not informed consent.
	if !strings.Contains(out, "hello.txt") {
		t.Errorf("prompt should list the entries it will delete, got %q", out)
	}
}

// The prompt must list exactly what RestoreSnapshot deletes — no drift.
func TestRestoreCommand_promptListsExactlyWhatWouldBeDeleted(t *testing.T) {
	withRestoreYes(t, false)
	id, dir := seedSnapshot(t)
	if err := os.WriteFile(filepath.Join(dir, "extra.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	pending, err := snapshot.PendingRemovals()
	if err != nil {
		t.Fatal(err)
	}
	out, err := runRestoreWith(t, id, "n\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range pending {
		if !strings.Contains(out, name) {
			t.Errorf("prompt omitted %q, which restore would have deleted; out=%q", name, out)
		}
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == ".eko" {
			t.Errorf(".eko is preserved by restore and must not be listed as a deletion")
		}
	}
}

// A piped stdin — `eko restore <id> < script` or a CI runner — must fail rather than
// proceed or hang. A pipe is used deliberately: os.DevNull is a *character device* on
// Unix, so it does not exercise the non-terminal branch (see the EOF test below).
func TestRestoreCommand_nonTTYWithoutYesRefuses(t *testing.T) {
	withRestoreYes(t, false)
	id, dir := seedSnapshot(t)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	w.Close() // no writer: a pipe, and not a terminal
	restoreCmd.SetIn(r)
	t.Cleanup(func() { restoreCmd.SetIn(nil) })

	err = restoreCmd.RunE(restoreCmd, []string{id})
	if !errors.Is(err, errRestoreNeedsTTY) {
		t.Fatalf("expected errRestoreNeedsTTY, got %v", err)
	}
	content, _ := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if string(content) != "unsaved work" {
		t.Errorf("a refused restore must not delete anything; got %q", string(content))
	}
}

// /dev/null reports as a character device, so it reaches the prompt and reads EOF.
// That must cancel, never proceed — an empty answer is not consent.
func TestRestoreCommand_devNullReadsAsEOFAndCancels(t *testing.T) {
	withRestoreYes(t, false)
	id, dir := seedSnapshot(t)

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer devNull.Close()
	restoreCmd.SetIn(devNull)
	t.Cleanup(func() { restoreCmd.SetIn(nil) })

	if err := restoreCmd.RunE(restoreCmd, []string{id}); err != nil {
		t.Fatalf("EOF should cancel cleanly, got %v", err)
	}
	content, _ := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if string(content) != "unsaved work" {
		t.Errorf("EOF must not be read as consent; got %q", string(content))
	}
}

// --yes keeps scripting working without a terminal.
func TestRestoreCommand_yesFlagSkipsPromptWithoutTTY(t *testing.T) {
	withRestoreYes(t, true)
	id, dir := seedSnapshot(t)

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer devNull.Close()
	restoreCmd.SetIn(devNull)
	t.Cleanup(func() { restoreCmd.SetIn(nil) })

	if err := restoreCmd.RunE(restoreCmd, []string{id}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, _ := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if string(content) != "saved" {
		t.Errorf("--yes should restore without prompting; got %q", string(content))
	}
}

// A missing snapshot must fail before anything is deleted or prompted.
func TestRestoreCommand_unknownIDFailsBeforePrompting(t *testing.T) {
	withRestoreYes(t, false)
	_, dir := seedSnapshot(t)

	out, err := runRestoreWith(t, "nope1234", "y\n")
	if err == nil {
		t.Fatal("expected an error for an unknown snapshot id")
	}
	if strings.Contains(out, "Continue?") {
		t.Errorf("must not prompt for a snapshot that does not exist, got %q", out)
	}
	content, _ := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if string(content) != "unsaved work" {
		t.Errorf("nothing should have been touched; got %q", string(content))
	}
}

func TestInitCommand_gitWarning(t *testing.T) {
	_ = setupTestDir(t)
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = initCmd.RunE(initCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Tip: A Git repository was detected") {
		t.Errorf("expected output to contain Git tip warning, got: %q", output)
	}
}

func TestSaveCommand_customMessage(t *testing.T) {
	dir := setupTestDir(t)
	_ = initCmd.RunE(initCmd, []string{})

	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set custom save message
	saveMessage = "custom test description"
	defer func() { saveMessage = "snapshot" }() // reset to default

	_ = saveCmd.RunE(saveCmd, []string{})

	database := db.InitDB()
	defer database.Close()

	var message string
	err := database.QueryRow("SELECT message FROM snapshots LIMIT 1").Scan(&message)
	if err != nil {
		t.Fatal(err)
	}
	if message != "custom test description" {
		t.Errorf("expected message 'custom test description', got %q", message)
	}
}

// --- history test helpers ---

// newLegacyProject creates .eko/snapshots and a database using the pre-summary
// schema, so history is exercised against a database written before the
// summary column existed.
func newLegacyProject(t *testing.T) string {
	t.Helper()
	dir := setupTestDir(t)
	if err := os.MkdirAll(filepath.Join(".eko", "snapshots"), 0755); err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("sqlite3", filepath.Join(".eko", "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.Exec(`CREATE TABLE snapshots (
		id TEXT PRIMARY KEY,
		message TEXT,
		path TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatal(err)
	}
	return dir
}

// addSnapshotRow inserts a snapshot row with an explicit recorded path.
func addSnapshotRow(t *testing.T, id, path, createdAt string) {
	t.Helper()
	database, err := sql.Open("sqlite3", filepath.Join(".eko", "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.Exec(
		"INSERT INTO snapshots(id, message, path, created_at) VALUES (?, ?, ?, ?)",
		id, "snapshot", path, createdAt,
	); err != nil {
		t.Fatal(err)
	}
}

// assertNotExist fails when path exists.
func assertNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected %s not to exist, stat error was %v", path, err)
	}
}

// --- eko history --format ------------------------------------------------

// withHistoryFormat sets the history output flags the way cobra would and
// restores the package defaults, including each flag's Changed state, which
// resolveHistoryFormat reads to tell an explicit --format from the default.
func withHistoryFormat(t *testing.T, format string, legacyJSON bool) {
	t.Helper()
	if format != "" {
		if err := historyCmd.Flags().Set("format", format); err != nil {
			t.Fatal(err)
		}
	}
	if legacyJSON {
		if err := historyCmd.Flags().Set("json", "true"); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		jsonOutput, verboseOutput, historyFormat = false, false, historyFormatText
		historyCmd.Flags().Lookup("format").Changed = false
		historyCmd.Flags().Lookup("json").Changed = false
	})
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Stdout = orig
		_ = w.Close()
		_ = r.Close()
	}()

	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = orig

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// newHistoryProject creates a current-schema project with no snapshots.
func newHistoryProject(t *testing.T) string {
	t.Helper()
	dir := setupTestDir(t)
	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatal(err)
	}
	return dir
}

// addHistoryRow inserts a snapshot row with an explicit message and summary.
func addHistoryRow(t *testing.T, id, createdAt, message, summary string) {
	t.Helper()
	database, err := sql.Open("sqlite3", filepath.Join(".eko", "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.Exec(
		"INSERT INTO snapshots(id, message, path, created_at, summary) VALUES (?, ?, ?, ?, ?)",
		id, message, ".eko/snapshots/"+id, createdAt, summary,
	); err != nil {
		t.Fatal(err)
	}
}

func TestResolveHistoryFormat(t *testing.T) {
	tests := []struct {
		name          string
		format        string
		formatChanged bool
		legacyJSON    bool
		want          string
		wantErr       string
	}{
		{name: "default is text", format: historyFormatText, want: historyFormatText},
		{name: "explicit text", format: historyFormatText, formatChanged: true, want: historyFormatText},
		{name: "explicit json", format: historyFormatJSON, formatChanged: true, want: historyFormatJSON},
		{name: "explicit md", format: historyFormatMD, formatChanged: true, want: historyFormatMD},
		{name: "explicit csv", format: historyFormatCSV, formatChanged: true, want: historyFormatCSV},
		{name: "legacy json alone", format: historyFormatText, legacyJSON: true, want: historyFormatJSON},
		{
			name: "legacy json agrees with explicit json", format: historyFormatJSON,
			formatChanged: true, legacyJSON: true, want: historyFormatJSON,
		},
		{
			name: "legacy json conflicts with md", format: historyFormatMD,
			formatChanged: true, legacyJSON: true, wantErr: "--json conflicts with --format md",
		},
		{
			name: "legacy json conflicts with explicit text", format: historyFormatText,
			formatChanged: true, legacyJSON: true, wantErr: "--json conflicts with --format text",
		},
		{
			name: "unknown format", format: "yaml", formatChanged: true,
			wantErr: `unsupported --format "yaml"`,
		},
		{
			// A typo should be reported as a typo even when --json is also set.
			name: "unknown format outranks the conflict", format: "yaml",
			formatChanged: true, legacyJSON: true, wantErr: `unsupported --format "yaml"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveHistoryFormat(tc.format, tc.formatChanged, tc.legacyJSON)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (format %q)", tc.wantErr, got)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("expected format %q, got %q", tc.want, got)
			}
		})
	}
}

func TestEscapeMarkdownCell(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "plain text is unchanged", value: "fix the parser", want: "fix the parser"},
		{name: "pipe is escaped", value: "a | b", want: "a \\| b"},
		{name: "newline becomes a space", value: "line one\nline two", want: "line one line two"},
		{name: "crlf becomes one space", value: "line one\r\nline two", want: "line one line two"},
		{name: "lone carriage return becomes a space", value: "a\rb", want: "a b"},
		{name: "pipes and newlines together", value: "a|b\nc|d", want: "a\\|b c\\|d"},
		{name: "empty stays empty", value: "", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := escapeMarkdownCell(tc.value); got != tc.want {
				t.Errorf("escapeMarkdownCell(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestHistoryCommand_markdownTableAndEscaping(t *testing.T) {
	newHistoryProject(t)
	addHistoryRow(t, "aaaaaaaa", "2026-01-01T00:00:00Z", "adds a|pipe", "summary\nwith a newline")
	withHistoryFormat(t, historyFormatMD, false)

	output := captureStdout(t, func() {
		if err := historyCmd.RunE(historyCmd, []string{}); err != nil {
			t.Fatal(err)
		}
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected header, separator and one row, got %d lines: %q", len(lines), output)
	}
	if lines[0] != "| ID | Created At | Message | Summary |" {
		t.Errorf("unexpected header: %q", lines[0])
	}
	if lines[1] != "| --- | --- | --- | --- |" {
		t.Errorf("unexpected separator: %q", lines[1])
	}
	want := "| aaaaaaaa | 2026-01-01T00:00:00Z | adds a\\|pipe | summary with a newline |"
	if lines[2] != want {
		t.Errorf("expected row %q, got %q", want, lines[2])
	}
}

func TestHistoryCommand_markdownWithNoEntriesStillWritesHeader(t *testing.T) {
	newHistoryProject(t)
	withHistoryFormat(t, historyFormatMD, false)

	output := captureStdout(t, func() {
		if err := historyCmd.RunE(historyCmd, []string{}); err != nil {
			t.Fatal(err)
		}
	})

	want := "| ID | Created At | Message | Summary |\n| --- | --- | --- | --- |\n"
	if output != want {
		t.Errorf("expected a header-only table %q, got %q", want, output)
	}
}

func TestHistoryCommand_csvQuotesAndRoundTrips(t *testing.T) {
	newHistoryProject(t)
	addHistoryRow(t, "aaaaaaaa", "2026-01-01T00:00:00Z", `comma, "quote" and
newline`, "")
	withHistoryFormat(t, historyFormatCSV, false)

	output := captureStdout(t, func() {
		if err := historyCmd.RunE(historyCmd, []string{}); err != nil {
			t.Fatal(err)
		}
	})

	records, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	if err != nil {
		t.Fatalf("expected parseable CSV, got %q: %v", output, err)
	}
	if len(records) != 2 {
		t.Fatalf("expected a header and one row, got %d records: %q", len(records), output)
	}
	wantHeader := []string{"id", "created_at", "message", "summary"}
	if strings.Join(records[0], ",") != strings.Join(wantHeader, ",") {
		t.Errorf("expected header %v, got %v", wantHeader, records[0])
	}
	// The message survives the round trip intact, unlike in Markdown where
	// newlines have to be collapsed to keep one entry on one row.
	wantRow := []string{"aaaaaaaa", "2026-01-01T00:00:00Z", "comma, \"quote\" and\nnewline", ""}
	for i := range wantRow {
		if records[1][i] != wantRow[i] {
			t.Errorf("field %d: expected %q, got %q", i, wantRow[i], records[1][i])
		}
	}
}

func TestHistoryCommand_defaultAndExplicitTextAreByteIdentical(t *testing.T) {
	newHistoryProject(t)
	addHistoryRow(t, "aaaaaaaa", "2026-01-01T00:00:00Z", "first", "a summary")
	addHistoryRow(t, "bbbbbbbb", "2026-01-02T00:00:00Z", "second", "")

	defaultOutput := captureStdout(t, func() {
		if err := historyCmd.RunE(historyCmd, []string{}); err != nil {
			t.Fatal(err)
		}
	})

	withHistoryFormat(t, historyFormatText, false)
	explicitOutput := captureStdout(t, func() {
		if err := historyCmd.RunE(historyCmd, []string{}); err != nil {
			t.Fatal(err)
		}
	})

	if defaultOutput != explicitOutput {
		t.Errorf("--format text changed the default rendering:\ndefault:  %q\nexplicit: %q",
			defaultOutput, explicitOutput)
	}
	if defaultOutput == "" {
		t.Error("expected the text renderer to print something")
	}
}

func TestHistoryCommand_legacyJSONMatchesFormatJSON(t *testing.T) {
	newHistoryProject(t)
	addHistoryRow(t, "aaaaaaaa", "2026-01-01T00:00:00Z", "first", "a summary")

	withHistoryFormat(t, "", true)
	legacyOutput := captureStdout(t, func() {
		if err := historyCmd.RunE(historyCmd, []string{}); err != nil {
			t.Fatal(err)
		}
	})

	jsonOutput = false
	historyCmd.Flags().Lookup("json").Changed = false
	withHistoryFormat(t, historyFormatJSON, false)
	formatOutput := captureStdout(t, func() {
		if err := historyCmd.RunE(historyCmd, []string{}); err != nil {
			t.Fatal(err)
		}
	})

	if legacyOutput != formatOutput {
		t.Errorf("--json and --format json disagree:\n--json:        %q\n--format json: %q",
			legacyOutput, formatOutput)
	}
	var entries []historyEntry
	if err := json.Unmarshal([]byte(formatOutput), &entries); err != nil {
		t.Fatalf("expected valid JSON, got %q: %v", formatOutput, err)
	}
}

func TestHistoryCommand_rejectsUnknownFormatWithoutPrinting(t *testing.T) {
	newHistoryProject(t)
	addHistoryRow(t, "aaaaaaaa", "2026-01-01T00:00:00Z", "first", "")
	withHistoryFormat(t, "yaml", false)

	var err error
	output := captureStdout(t, func() {
		err = historyCmd.RunE(historyCmd, []string{})
	})

	if err == nil {
		t.Fatal("expected an unsupported format to be rejected")
	}
	if !strings.Contains(err.Error(), `unsupported --format "yaml"`) {
		t.Errorf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected no output before the format was rejected, got %q", output)
	}
}

func TestHistoryCommand_rejectsLegacyJSONWithConflictingFormat(t *testing.T) {
	newHistoryProject(t)
	addHistoryRow(t, "aaaaaaaa", "2026-01-01T00:00:00Z", "first", "")
	withHistoryFormat(t, historyFormatCSV, true)

	var err error
	output := captureStdout(t, func() {
		err = historyCmd.RunE(historyCmd, []string{})
	})

	if err == nil {
		t.Fatal("expected --json with --format csv to be rejected")
	}
	if !strings.Contains(err.Error(), "--json conflicts with --format csv") {
		t.Errorf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected no output before the conflict was rejected, got %q", output)
	}
}

func TestHistoryCommand_legacySchemaRendersEveryFormat(t *testing.T) {
	// A database written before the summary column existed falls back to the
	// two-column query, so every renderer must cope with empty message and
	// summary fields rather than assuming the wide schema.
	newLegacyProject(t)
	addSnapshotRow(t, "aaaaaaaa", ".eko/snapshots/aaaaaaaa", "2026-01-01T00:00:00Z")

	for _, format := range []string{historyFormatMD, historyFormatCSV, historyFormatJSON} {
		t.Run(format, func(t *testing.T) {
			withHistoryFormat(t, format, false)
			output := captureStdout(t, func() {
				if err := historyCmd.RunE(historyCmd, []string{}); err != nil {
					t.Fatal(err)
				}
			})
			if !strings.Contains(output, "aaaaaaaa") {
				t.Errorf("expected %s output to contain the snapshot id, got %q", format, output)
			}
		})
	}
}

func TestHistoryCommand_badFormatFailsBeforeTouchingTheDatabase(t *testing.T) {
	// The format is resolved before db.InitDB(), which would otherwise create
	// .eko/db.sqlite. A marker-only .eko proves the ordering: if resolution
	// moved after the open, running with a bad format would leave a database
	// behind for a command that never got as far as reading one.
	setupTestDir(t)
	if err := os.MkdirAll(".eko", 0755); err != nil {
		t.Fatal(err)
	}
	withHistoryFormat(t, "yaml", false)

	if err := historyCmd.RunE(historyCmd, []string{}); err == nil {
		t.Fatal("expected an unsupported format to be rejected")
	}

	for _, name := range []string{"db.sqlite", "db.sqlite-wal", "db.sqlite-shm"} {
		assertNotExist(t, filepath.Join(".eko", name))
	}
}

// --- eko clean ------------------------------------------------------------

// withCleanFlags sets the clean flags for one test and restores the defaults.
func withCleanFlags(t *testing.T, keep int, dryRun bool) {
	t.Helper()
	cleanKeep, cleanDryRun = keep, dryRun
	t.Cleanup(func() { cleanKeep, cleanDryRun = 10, false })
}

// addSnapshot inserts a well-formed snapshot row and creates its directory.
func addSnapshot(t *testing.T, id, createdAt string) {
	t.Helper()
	addSnapshotRow(t, id, ".eko/snapshots/"+id, createdAt)
	dir := filepath.Join(".eko", "snapshots", id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte(id), 0644); err != nil {
		t.Fatal(err)
	}
}

// snapshotIDs returns the snapshot ids still recorded, newest first.
func snapshotIDs(t *testing.T) []string {
	t.Helper()
	database, err := sql.Open("sqlite3", filepath.Join(".eko", "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	rows, err := database.Query("SELECT id FROM snapshots ORDER BY created_at DESC, rowid DESC")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return ids
}

// snapshotColumns returns the column names of the snapshots table.
func snapshotColumns(t *testing.T) []string {
	t.Helper()
	database, err := sql.Open("sqlite3", filepath.Join(".eko", "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	rows, err := database.Query("PRAGMA table_info(snapshots)")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	cols := []string{}
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		cols = append(cols, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return cols
}

func hasColumn(cols []string, name string) bool {
	for _, c := range cols {
		if c == name {
			return true
		}
	}
	return false
}

// hashFile returns the SHA-256 of a file, used to prove byte-for-byte immutability.
func hashFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// assertDirExists fails when path is not an existing directory.
func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("expected %s to exist: %v", path, err)
		return
	}
	if !info.IsDir() {
		t.Errorf("expected %s to be a directory", path)
	}
}

func TestCleanCommand_flagDefaults(t *testing.T) {
	if got := cleanCmd.Flags().Lookup("keep").DefValue; got != "10" {
		t.Errorf("expected --keep to default to 10, got %q", got)
	}
	if got := cleanCmd.Flags().Lookup("dry-run").DefValue; got != "false" {
		t.Errorf("expected --dry-run to default to false, got %q", got)
	}
}

// A missing database must produce an explicit error. Neither open path may
// create the database or its WAL/SHM sidecars before validation.
func TestCleanCommand_missingDatabaseCreatesNothing(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		name := "normal"
		if dryRun {
			name = "dry-run"
		}
		t.Run(name, func(t *testing.T) {
			setupTestDir(t)
			// Marker-only project: .eko exists, the database does not.
			if err := os.MkdirAll(filepath.Join(".eko", "snapshots"), 0755); err != nil {
				t.Fatal(err)
			}
			withCleanFlags(t, 1, dryRun)

			err := cleanCmd.RunE(cleanCmd, []string{})
			if err == nil {
				t.Fatal("expected an explicit missing-database error, got nil")
			}
			if !strings.Contains(err.Error(), "no eko database found") {
				t.Fatalf("expected a missing-database error, got %v", err)
			}

			assertNotExist(t, filepath.Join(".eko", "db.sqlite"))
			assertNotExist(t, filepath.Join(".eko", "db.sqlite-wal"))
			assertNotExist(t, filepath.Join(".eko", "db.sqlite-shm"))
		})
	}
}

// A dry run must not change one byte of the database, even on a legacy schema
// that InitDB would have migrated on open.
func TestCleanCommand_dryRunLeavesLegacyDatabaseByteIdentical(t *testing.T) {
	newLegacyProject(t)
	addSnapshot(t, "aaaaaaaa", "2026-01-01 00:00:00")
	addSnapshot(t, "bbbbbbbb", "2026-01-02 00:00:00")
	addSnapshot(t, "cccccccc", "2026-01-03 00:00:00")

	dbPath := filepath.Join(".eko", "db.sqlite")
	before := hashFile(t, dbPath)

	withCleanFlags(t, 1, true)
	if err := cleanCmd.RunE(cleanCmd, []string{}); err != nil {
		t.Fatal(err)
	}

	if after := hashFile(t, dbPath); after != before {
		t.Errorf("dry run changed the database: %s -> %s", before, after)
	}
	assertNotExist(t, filepath.Join(".eko", "db.sqlite-wal"))
	assertNotExist(t, filepath.Join(".eko", "db.sqlite-shm"))

	for _, id := range []string{"aaaaaaaa", "bbbbbbbb", "cccccccc"} {
		assertDirExists(t, filepath.Join(".eko", "snapshots", id))
	}
	if got := strings.Join(snapshotIDs(t), ","); got != "cccccccc,bbbbbbbb,aaaaaaaa" {
		t.Errorf("dry run changed the snapshot rows, got %v", got)
	}
	if cols := snapshotColumns(t); hasColumn(cols, "summary") {
		t.Errorf("dry run migrated the legacy schema: %v", cols)
	}
}

// A normal run with nothing to remove must also leave a legacy schema alone.
func TestCleanCommand_normalRunDoesNotMigrateLegacySchema(t *testing.T) {
	newLegacyProject(t)
	addSnapshot(t, "aaaaaaaa", "2026-01-01 00:00:00")

	dbPath := filepath.Join(".eko", "db.sqlite")
	before := hashFile(t, dbPath)

	withCleanFlags(t, 10, false)
	if err := cleanCmd.RunE(cleanCmd, []string{}); err != nil {
		t.Fatal(err)
	}

	if after := hashFile(t, dbPath); after != before {
		t.Errorf("a run with no candidates changed the database: %s -> %s", before, after)
	}
	cols := snapshotColumns(t)
	if hasColumn(cols, "summary") {
		t.Errorf("clean added the summary column to a legacy database: %v", cols)
	}
	if strings.Join(cols, ",") != "id,message,path,created_at" {
		t.Errorf("expected the legacy schema to be preserved, got %v", cols)
	}
	assertDirExists(t, filepath.Join(".eko", "snapshots", "aaaaaaaa"))
}

// Snapshots sharing a created_at must still be ordered deterministically, so
// the same keep count always removes the same snapshots.
func TestCleanCommand_deterministicOrder(t *testing.T) {
	newLegacyProject(t)
	for _, id := range []string{"aaaaaaaa", "bbbbbbbb", "cccccccc", "dddddddd"} {
		addSnapshot(t, id, "2026-01-01 00:00:00")
	}

	withCleanFlags(t, 2, false)
	if err := cleanCmd.RunE(cleanCmd, []string{}); err != nil {
		t.Fatal(err)
	}

	if got := strings.Join(snapshotIDs(t), ","); got != "dddddddd,cccccccc" {
		t.Errorf("expected the two newest rows to survive, got %v", got)
	}
	assertDirExists(t, filepath.Join(".eko", "snapshots", "cccccccc"))
	assertDirExists(t, filepath.Join(".eko", "snapshots", "dddddddd"))
	assertNotExist(t, filepath.Join(".eko", "snapshots", "aaaaaaaa"))
	assertNotExist(t, filepath.Join(".eko", "snapshots", "bbbbbbbb"))
}

func TestCleanCommand_keepZeroRemovesEverything(t *testing.T) {
	newLegacyProject(t)
	addSnapshot(t, "aaaaaaaa", "2026-01-01 00:00:00")
	addSnapshot(t, "bbbbbbbb", "2026-01-02 00:00:00")

	withCleanFlags(t, 0, false)
	if err := cleanCmd.RunE(cleanCmd, []string{}); err != nil {
		t.Fatal(err)
	}

	if got := snapshotIDs(t); len(got) != 0 {
		t.Errorf("expected every snapshot row to be removed, got %v", got)
	}
	assertNotExist(t, filepath.Join(".eko", "snapshots", "aaaaaaaa"))
	assertNotExist(t, filepath.Join(".eko", "snapshots", "bbbbbbbb"))
	assertDirExists(t, filepath.Join(".eko", "snapshots"))
}

func TestCleanCommand_rejectsNegativeKeep(t *testing.T) {
	newLegacyProject(t)
	addSnapshot(t, "aaaaaaaa", "2026-01-01 00:00:00")

	withCleanFlags(t, -1, false)
	err := cleanCmd.RunE(cleanCmd, []string{})
	if err == nil {
		t.Fatal("expected a negative --keep to be rejected")
	}
	if !strings.Contains(err.Error(), "--keep must be zero or greater") {
		t.Fatalf("expected a --keep validation error, got %v", err)
	}
	assertDirExists(t, filepath.Join(".eko", "snapshots", "aaaaaaaa"))
	if got := len(snapshotIDs(t)); got != 1 {
		t.Errorf("expected the snapshot row to survive, got %d rows", got)
	}
}

// Every candidate is validated before any of them is removed, so one bad row
// stops the whole run with nothing deleted.
func TestCleanCommand_pathIDMismatchAbortsBeforeAnyDeletion(t *testing.T) {
	newLegacyProject(t)
	addSnapshot(t, "aaaaaaaa", "2026-01-02 00:00:00")
	// Older, so it is validated after the well-formed candidate: its recorded
	// path belongs to a different snapshot id.
	addSnapshotRow(t, "bbbbbbbb", ".eko/snapshots/aaaaaaaa", "2026-01-01 00:00:00")

	withCleanFlags(t, 0, false)
	err := cleanCmd.RunE(cleanCmd, []string{})
	if err == nil {
		t.Fatal("expected the mismatched path to abort the run")
	}
	if !strings.Contains(err.Error(), "is not") {
		t.Fatalf("expected a recorded-path mismatch error, got %v", err)
	}

	assertDirExists(t, filepath.Join(".eko", "snapshots", "aaaaaaaa"))
	if got := len(snapshotIDs(t)); got != 2 {
		t.Errorf("expected both snapshot rows to survive, got %d rows", got)
	}
}

// A row whose directory is already gone is stale, not suspicious: there is
// nothing left to delete on disk, so the run cleans up the row and carries on.
// This is what lets a run interrupted between the disk delete and the row
// delete be finished by the next one.
func TestCleanCommand_missingSnapshotDirectoryIsRecovered(t *testing.T) {
	newLegacyProject(t)
	addSnapshot(t, "aaaaaaaa", "2026-01-02 00:00:00")
	addSnapshotRow(t, "bbbbbbbb", ".eko/snapshots/bbbbbbbb", "2026-01-01 00:00:00")

	withCleanFlags(t, 0, false)
	if err := cleanCmd.RunE(cleanCmd, []string{}); err != nil {
		t.Fatalf("expected the stale row to be recovered, got %v", err)
	}

	if got := len(snapshotIDs(t)); got != 0 {
		t.Errorf("expected every row to be removed, got %d rows", got)
	}
	assertNotExist(t, filepath.Join(".eko", "snapshots", "aaaaaaaa"))
}

// Recovering a missing directory must not widen what clean will delete. A
// dangling symlink is an entry that exists, so it is still refused.
func TestCleanCommand_rejectsDanglingSymlinkSnapshot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs elevated privileges on Windows")
	}

	newLegacyProject(t)
	addSnapshotRow(t, "aaaaaaaa", ".eko/snapshots/aaaaaaaa", "2026-01-01 00:00:00")
	if err := os.Symlink(filepath.Join("..", "..", "gone"), filepath.Join(".eko", "snapshots", "aaaaaaaa")); err != nil {
		t.Fatal(err)
	}

	withCleanFlags(t, 0, false)
	err := cleanCmd.RunE(cleanCmd, []string{})
	if err == nil {
		t.Fatal("expected a dangling symlink to abort the run")
	}
	if !strings.Contains(err.Error(), "cannot resolve") {
		t.Fatalf("expected an unresolvable-path error, got %v", err)
	}
	if got := len(snapshotIDs(t)); got != 1 {
		t.Errorf("expected the row to survive, got %d rows", got)
	}
}

// The end-to-end promise in the deletion comment: a run that removed a
// directory and then failed to delete its row leaves state the next run can
// finish, rather than a project that can never be cleaned again.
func TestCleanCommand_resumesAfterAnInterruptedRun(t *testing.T) {
	newLegacyProject(t)
	addSnapshot(t, "aaaaaaaa", "2026-01-01 00:00:00")
	addSnapshot(t, "bbbbbbbb", "2026-01-02 00:00:00")
	addSnapshot(t, "cccccccc", "2026-01-03 00:00:00")

	// Exactly the state an interrupted run leaves behind: the directory is gone
	// and the row survives.
	if err := os.RemoveAll(filepath.Join(".eko", "snapshots", "bbbbbbbb")); err != nil {
		t.Fatal(err)
	}

	withCleanFlags(t, 1, false)
	if err := cleanCmd.RunE(cleanCmd, []string{}); err != nil {
		t.Fatalf("expected the interrupted run to be resumable, got %v", err)
	}

	if got := strings.Join(snapshotIDs(t), ","); got != "cccccccc" {
		t.Errorf("expected only the kept snapshot to remain, got %v", got)
	}
	assertNotExist(t, filepath.Join(".eko", "snapshots", "aaaaaaaa"))
	assertDirExists(t, filepath.Join(".eko", "snapshots", "cccccccc"))
}

// When the row delete fails the snapshot is already off disk, so the reported
// count has to include it. Reverting the i+1 in clean.go makes this report
// "removed 0 of 2" and fails here.
func TestCleanCommand_rowDeletionFailureCountsTheRemovedSnapshot(t *testing.T) {
	newLegacyProject(t)
	addSnapshot(t, "aaaaaaaa", "2026-01-01 00:00:00") // oldest, removed second
	addSnapshot(t, "bbbbbbbb", "2026-01-02 00:00:00") // removed first
	addSnapshot(t, "cccccccc", "2026-01-03 00:00:00") // newest, kept

	// Block only the first row delete, leaving the disk delete to succeed.
	database, err := sql.Open("sqlite3", filepath.Join(".eko", "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`CREATE TRIGGER block_bbbbbbbb BEFORE DELETE ON snapshots
		WHEN OLD.id = 'bbbbbbbb'
		BEGIN SELECT RAISE(ABORT, 'row delete blocked'); END`); err != nil {
		database.Close()
		t.Fatal(err)
	}
	database.Close()

	withCleanFlags(t, 1, false)
	err = cleanCmd.RunE(cleanCmd, []string{})
	if err == nil {
		t.Fatal("expected the blocked row delete to fail the run")
	}
	if !strings.Contains(err.Error(), "could not delete its database row") {
		t.Fatalf("expected a row-delete error, got %v", err)
	}
	if !strings.Contains(err.Error(), "removed 1 of 2") {
		t.Fatalf("expected the removed snapshot to be counted, got %v", err)
	}

	// The count is honest: the directory really is gone and the row really did
	// survive.
	assertNotExist(t, filepath.Join(".eko", "snapshots", "bbbbbbbb"))
	if got := strings.Join(snapshotIDs(t), ","); got != "cccccccc,bbbbbbbb,aaaaaaaa" {
		t.Errorf("expected every row to survive the blocked delete, got %v", got)
	}
}

func TestCleanCommand_rejectsNonDirectorySnapshot(t *testing.T) {
	newLegacyProject(t)
	addSnapshotRow(t, "aaaaaaaa", ".eko/snapshots/aaaaaaaa", "2026-01-01 00:00:00")
	if err := os.WriteFile(filepath.Join(".eko", "snapshots", "aaaaaaaa"), []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}

	withCleanFlags(t, 0, false)
	err := cleanCmd.RunE(cleanCmd, []string{})
	if err == nil {
		t.Fatal("expected a non-directory snapshot path to abort the run")
	}
	if !strings.Contains(err.Error(), "is not a directory") {
		t.Fatalf("expected a non-directory error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(".eko", "snapshots", "aaaaaaaa")); statErr != nil {
		t.Errorf("expected the file to survive: %v", statErr)
	}
}

// A snapshot directory that is a symlink out of .eko/snapshots must never be
// followed, and a symlink to the snapshots directory itself must be refused.
func TestCleanCommand_rejectsSymlinkedSnapshotPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs elevated privileges on Windows")
	}

	t.Run("escapes the snapshots directory", func(t *testing.T) {
		dir := newLegacyProject(t)
		outside := filepath.Join(dir, "outside")
		if err := os.MkdirAll(outside, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(outside, "keep.txt"), []byte("keep"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(".eko", "snapshots", "aaaaaaaa")); err != nil {
			t.Fatal(err)
		}
		addSnapshotRow(t, "aaaaaaaa", ".eko/snapshots/aaaaaaaa", "2026-01-01 00:00:00")

		withCleanFlags(t, 0, false)
		err := cleanCmd.RunE(cleanCmd, []string{})
		if err == nil {
			t.Fatal("expected an escaping symlink to abort the run")
		}
		if !strings.Contains(err.Error(), "resolves outside") {
			t.Fatalf("expected a containment error, got %v", err)
		}
		assertDirExists(t, outside)
		if _, statErr := os.Stat(filepath.Join(outside, "keep.txt")); statErr != nil {
			t.Errorf("expected the outside file to survive: %v", statErr)
		}
	})

	t.Run("targets the snapshots directory", func(t *testing.T) {
		dir := newLegacyProject(t)
		root := filepath.Join(dir, ".eko", "snapshots")
		if err := os.Symlink(root, filepath.Join(".eko", "snapshots", "aaaaaaaa")); err != nil {
			t.Fatal(err)
		}
		addSnapshotRow(t, "aaaaaaaa", ".eko/snapshots/aaaaaaaa", "2026-01-01 00:00:00")

		withCleanFlags(t, 0, false)
		err := cleanCmd.RunE(cleanCmd, []string{})
		if err == nil {
			t.Fatal("expected a snapshots-root symlink to abort the run")
		}
		if !strings.Contains(err.Error(), "snapshots directory itself") {
			t.Fatalf("expected a snapshots-root error, got %v", err)
		}
		assertDirExists(t, root)
	})

	// An alias that stays inside the snapshots directory passes every
	// containment check above: it resolves to a real direct child of the root.
	// Only the row id ties a candidate to the directory that is about to be
	// removed. Here the alias points at the one snapshot --keep preserves, so
	// without that check clean deletes a kept snapshot off disk and leaves its
	// row behind pointing at nothing.
	t.Run("aliases another snapshot inside the snapshots directory", func(t *testing.T) {
		dir := newLegacyProject(t)
		addSnapshot(t, "cccccccc", "2026-01-03 00:00:00") // newest, kept by --keep 1
		alias := filepath.Join(".eko", "snapshots", "aaaaaaaa")
		if err := os.Symlink(filepath.Join(dir, ".eko", "snapshots", "cccccccc"), alias); err != nil {
			t.Fatal(err)
		}
		// Oldest, so it is the only candidate, and it resolves to the kept one.
		addSnapshotRow(t, "aaaaaaaa", ".eko/snapshots/aaaaaaaa", "2026-01-01 00:00:00")

		dbPath := filepath.Join(".eko", "db.sqlite")
		before := hashFile(t, dbPath)

		withCleanFlags(t, 1, false)
		err := cleanCmd.RunE(cleanCmd, []string{})
		if err == nil {
			t.Fatal("expected an in-root snapshot alias to abort the run")
		}
		if !strings.Contains(err.Error(), "resolves to a different snapshot") {
			t.Fatalf("expected an alias error, got %v", err)
		}

		// Validation completes before the first deletion, so the kept snapshot,
		// the alias itself and every row must be untouched.
		assertDirExists(t, filepath.Join(".eko", "snapshots", "cccccccc"))
		content, readErr := os.ReadFile(filepath.Join(".eko", "snapshots", "cccccccc", "hello.txt"))
		if readErr != nil || string(content) != "cccccccc" {
			t.Errorf("expected the kept snapshot to be intact, got err=%v content=%q", readErr, string(content))
		}
		if _, lstatErr := os.Lstat(alias); lstatErr != nil {
			t.Errorf("expected the alias itself to survive: %v", lstatErr)
		}
		if after := hashFile(t, dbPath); after != before {
			t.Errorf("the rejected run changed the database: %s -> %s", before, after)
		}
		if got := strings.Join(snapshotIDs(t), ","); got != "cccccccc,aaaaaaaa" {
			t.Errorf("expected both snapshot rows to survive, got %v", got)
		}
	})
}

// Deletion is not atomic: the run stops at the first failure, keeps what it
// already removed, and reports exactly how far it got.
func TestCleanCommand_nonAtomicFailureReportsProgress(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions do not block removal on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}

	newLegacyProject(t)
	addSnapshot(t, "aaaaaaaa", "2026-01-01 00:00:00") // oldest, removed last
	addSnapshot(t, "bbbbbbbb", "2026-01-02 00:00:00")
	addSnapshot(t, "cccccccc", "2026-01-03 00:00:00") // newest, kept

	// RemoveAll cannot unlink a file inside a directory it may not write to.
	locked := filepath.Join(".eko", "snapshots", "aaaaaaaa", "locked")
	if err := os.MkdirAll(locked, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locked, "pinned.txt"), []byte("pinned"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0755) })

	withCleanFlags(t, 1, false)
	err := cleanCmd.RunE(cleanCmd, []string{})
	if err == nil {
		t.Fatal("expected the blocked removal to fail the run")
	}
	if !strings.Contains(err.Error(), "removed 1 of 2") {
		t.Fatalf("expected the error to report partial progress, got %v", err)
	}

	assertNotExist(t, filepath.Join(".eko", "snapshots", "bbbbbbbb"))
	assertDirExists(t, filepath.Join(".eko", "snapshots", "aaaaaaaa"))
	if got := strings.Join(snapshotIDs(t), ","); got != "cccccccc,aaaaaaaa" {
		t.Errorf("expected only the removed snapshot's row to be gone, got %v", got)
	}
}
