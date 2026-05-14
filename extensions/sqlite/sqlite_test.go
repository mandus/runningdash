package sqlite_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/anaseto/goal"
	"github.com/mandus/runningdash/extensions/sqlite"
)

func TestSQLiteOpenClose(t *testing.T) {
	ctx := goal.NewContext()
	sqlite.Import(ctx, "")

	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Test open
	x, err := ctx.Eval(fmt.Sprintf(`sqlite.open["%s"]`, dbPath))
	if err != nil {
		t.Fatalf("sqlite.open failed: %v", err)
	}

	connID, ok := x.BV().(goal.I)
	if !ok {
		t.Fatalf("Expected integer connection ID, got %s", x.Type())
	}

	if connID == 0 {
		t.Fatal("Expected non-zero connection ID")
	}

	// Test close
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.close[%d]`, connID))
	if err != nil {
		t.Fatalf("sqlite.close failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Database file was not created")
	}

	// Clean up
	os.Remove(dbPath)
}

func TestSQLiteExec(t *testing.T) {
	ctx := goal.NewContext()
	sqlite.Import(ctx, "")

	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open connection
	x, err := ctx.Eval(fmt.Sprintf(`sqlite.open["%s"]`, dbPath))
	if err != nil {
		t.Fatalf("sqlite.open failed: %v", err)
	}
	connID := x.BV().(goal.I)

	// Create table
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%d;"CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)";[]]`, connID))
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	// Insert data
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%d;"INSERT INTO test (name) VALUES (?)";["test_value"]]`, connID))
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	// Close connection
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.close[%d]`, connID))
	if err != nil {
		t.Fatalf("sqlite.close failed: %v", err)
	}

	// Clean up
	os.Remove(dbPath)
}

func TestSQLiteQuery(t *testing.T) {
	ctx := goal.NewContext()
	sqlite.Import(ctx, "")

	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open connection
	x, err := ctx.Eval(fmt.Sprintf(`sqlite.open["%s"]`, dbPath))
	if err != nil {
		t.Fatalf("sqlite.open failed: %v", err)
	}
	connID := x.BV().(goal.I)

	// Create table and insert data
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%d;"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)";[]]`, connID))
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%d;"INSERT INTO users (name) VALUES (?)";["Alice"]]`, connID))
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%d;"INSERT INTO users (name) VALUES (?)";["Bob"]]`, connID))
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	// Query data
	x, err = ctx.Eval(fmt.Sprintf(`sqlite.query[%d;"SELECT * FROM users ORDER BY id";[]]`, connID))
	if err != nil {
		t.Fatalf("sqlite.query failed: %v", err)
	}

	// Check result type
	if x.Type() != goal.AT {
		t.Fatalf("Expected array of rows, got %s", x.Type())
	}

	rows, ok := x.BV().(*goal.AS)
	if !ok {
		t.Fatalf("Expected array, got %s", x.Type())
	}

	if rows.Len() != 2 {
		t.Fatalf("Expected 2 rows, got %d", rows.Len())
	}

	// Close connection
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.close[%d]`, connID))
	if err != nil {
		t.Fatalf("sqlite.close failed: %v", err)
	}

	// Clean up
	os.Remove(dbPath)
}

func TestSQLiteInvalidConnection(t *testing.T) {
	ctx := goal.NewContext()
	sqlite.Import(ctx, "")

	// Try to use invalid connection ID
	_, err := ctx.Eval(`sqlite.query[999;"SELECT 1";[]]`)
	if err == nil {
		t.Fatal("Expected error for invalid connection ID")
	}

	if !fmt.Sprintf("%v", err).Contains("connection 999 not found") {
		t.Fatalf("Expected 'connection not found' error, got: %v", err)
	}
}
