package sqlite_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/anaseto/goal"
	"github.com/mandus/runningdash/extensions/sqlite"
)

func newCtx(t *testing.T) *goal.Context {
	t.Helper()
	ctx := goal.NewContext()
	sqlite.Import(ctx, "")
	return ctx
}

func getInt(t *testing.T, v goal.V, name string) int64 {
	t.Helper()
	if !v.IsI() {
		t.Fatalf("%s: expected integer, got %s", name, v.Type())
	}
	return v.I()
}

func TestSQLiteOpenClose(t *testing.T) {
	ctx := newCtx(t)

	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Test open
	x, err := ctx.Eval(fmt.Sprintf(`sqlite.open["%s"]`, dbPath))
	if err != nil {
		t.Fatalf("sqlite.open failed: %v", err)
	}

	if !x.IsI() {
		t.Fatalf("Expected integer connection ID, got %s", x.Type())
	}
	connID := x.I()

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
	ctx := newCtx(t)

	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open connection
	x, err := ctx.Eval(fmt.Sprintf(`sqlite.open["%s"]`, dbPath))
	if err != nil {
		t.Fatalf("sqlite.open failed: %v", err)
	}
	connID := getInt(t, x, "sqlite.open")

	// Create table using dict request without params
	createReq := fmt.Sprintf(`("conn_id";"sql")!(%d;"CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")`, connID)
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%s]`, createReq))
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	// Insert data using dict request with params (string, not array - Goal treats single values as strings)
	insertReq := fmt.Sprintf(`("conn_id";"sql";"params")!(%d;"INSERT INTO test (name) VALUES (?)";"test_value")`, connID)
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%s]`, insertReq))
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
	ctx := newCtx(t)

	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open connection
	x, err := ctx.Eval(fmt.Sprintf(`sqlite.open["%s"]`, dbPath))
	if err != nil {
		t.Fatalf("sqlite.open failed: %v", err)
	}
	connID := getInt(t, x, "sqlite.open")

	// Create table and insert data
	createReq := fmt.Sprintf(`("conn_id";"sql")!(%d;"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")`, connID)
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%s]`, createReq))
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	// Insert data using dict request with params (strings, not arrays)
	insertReq1 := fmt.Sprintf(`("conn_id";"sql";"params")!(%d;"INSERT INTO users (name) VALUES (?)";"Alice")`, connID)
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%s]`, insertReq1))
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	insertReq2 := fmt.Sprintf(`("conn_id";"sql";"params")!(%d;"INSERT INTO users (name) VALUES (?)";"Bob")`, connID)
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%s]`, insertReq2))
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	// Query data using dict request
	queryReq := fmt.Sprintf(`("conn_id";"sql")!(%d;"SELECT * FROM users ORDER BY id")`, connID)
	x, err = ctx.Eval(fmt.Sprintf(`sqlite.query[%s]`, queryReq))
	if err != nil {
		t.Fatalf("sqlite.query failed: %v", err)
	}

	// Check result type - should be array type ("A" in Goal for array of values)
	if x.Type() != "A" {
		t.Fatalf("Expected array of rows, got %s", x.Type())
	}

	rows, ok := x.BV().(*goal.AV)
	if !ok {
		t.Fatalf("Expected *goal.AV, got %s", x.Type())
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

func TestSQLiteQueryWithParams(t *testing.T) {
	ctx := newCtx(t)

	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open connection
	x, err := ctx.Eval(fmt.Sprintf(`sqlite.open["%s"]`, dbPath))
	if err != nil {
		t.Fatalf("sqlite.open failed: %v", err)
	}
	connID := getInt(t, x, "sqlite.open")

	// Create table and insert data
	createReq := fmt.Sprintf(`("conn_id";"sql")!(%d;"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")`, connID)
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%s]`, createReq))
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	// Insert data - use string directly (Goal treats single strings as single-element arrays in SQLite extension)
	insertReq := fmt.Sprintf(`("conn_id";"sql";"params")!(%d;"INSERT INTO users (name) VALUES (?)";"Alice")`, connID)
	_, err = ctx.Eval(fmt.Sprintf(`sqlite.exec[%s]`, insertReq))
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	// Query with parameters - use string directly
	queryReq := fmt.Sprintf(`sqlite.query[("conn_id";"sql";"params")!(%d;"SELECT name FROM users WHERE name = ?";"Alice")]`, connID)
	x, err = ctx.Eval(queryReq)
	if err != nil {
		t.Fatalf("sqlite.query failed: %v", err)
	}

	rows, ok := x.BV().(*goal.AV)
	if !ok {
		t.Fatalf("Expected array of values (AV), got %s", x.Type())
	}

	if rows.Len() != 1 {
		t.Fatalf("Expected 1 row, got %d", rows.Len())
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
	ctx := newCtx(t)

	// Try to use invalid connection ID in exec
	_, err := ctx.Eval(`sqlite.exec[("conn_id";"sql")!(999;"SELECT 1")]`)
	if err == nil {
		t.Fatal("Expected error for invalid connection ID")
	}

	if !strings.Contains(fmt.Sprintf("%v", err), "connection 999 not found") {
		t.Fatalf("Expected 'connection not found' error, got: %v", err)
	}
}
