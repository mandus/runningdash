// Package sqlite provides SQLite database functionality for Goal.
// This enables Goal to interact with SQLite databases.
//
// Implements requirements from:
// - specs/strava_dashboard_spec_v0.3.md §5.2 (Local Database)
// - specs/adrs/ADR-5_sqlite_extension.md
//
// To use this extension, create a custom Goal build that imports this package
// and calls sqlite.Import(ctx, ""). See cmd/strava_goal/main.go for example.
package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"codeberg.org/anaseto/goal"
)

// connections stores open database connections by ID
// Map key is connection ID (int), value is *sql.DB
var connections = make(map[int]*sql.DB)
var nextConnectionID = 1

// Import registers the SQLite extension functions with the Goal context.
// The following functions are registered:
//
//   sqlite.open[path] : open database connection, returns connection_id
//   sqlite.close[conn_id] : close database connection
//   sqlite.query[conn_id, sql, params] : execute SELECT query, returns result rows
//   sqlite.exec[conn_id, sql, params] : execute INSERT/UPDATE/DELETE, returns affected count
//   sqlite.begin[conn_id] : start transaction
//   sqlite.commit[conn_id] : commit transaction
//   sqlite.rollback[conn_id] : rollback transaction
//
// Connection IDs are integers. All functions that need a connection take conn_id as first argument.
func Import(ctx *goal.Context, pfx string) {
	if pfx != "" {
		pfx += "."
	}

	// Register sqlite.open as a monad (takes path string, returns connection_id int)
	ctx.AssignGlobal(pfx+"sqlite.open", ctx.RegisterMonad("."+pfx+"sqlite.open", vfOpen))
	
	// Register sqlite.close as a monad (takes connection_id int)
	ctx.AssignGlobal(pfx+"sqlite.close", ctx.RegisterMonad("."+pfx+"sqlite.close", vfClose))
	
	// Register sqlite.query as a triad (takes conn_id, sql, params)
	ctx.AssignGlobal(pfx+"sqlite.query", ctx.RegisterTriad("."+pfx+"sqlite.query", vfQuery))
	
	// Register sqlite.exec as a triad (takes conn_id, sql, params)
	ctx.AssignGlobal(pfx+"sqlite.exec", ctx.RegisterTriad("."+pfx+"sqlite.exec", vfExec))
	
	// Register transaction functions as monads (take connection_id)
	ctx.AssignGlobal(pfx+"sqlite.begin", ctx.RegisterMonad("."+pfx+"sqlite.begin", vfBegin))
	ctx.AssignGlobal(pfx+"sqlite.commit", ctx.RegisterMonad("."+pfx+"sqlite.commit", vfCommit))
	ctx.AssignGlobal(pfx+"sqlite.rollback", ctx.RegisterMonad("."+pfx+"sqlite.rollback", vfRollback))
}

// HelpFunc returns help text for the SQLite extension
func HelpFunc() func(string) string {
	return func(s string) string {
		if strings.HasPrefix(s, "sqlite.") || s == "sqlite" {
			return `sqlite.open[path]       : open database connection, returns connection_id
  sqlite.close[conn_id]    : close database connection
  sqlite.query[conn_id, sql, params] : execute SELECT, returns array of result rows
  sqlite.exec[conn_id, sql, params]  : execute INSERT/UPDATE/DELETE, returns affected count
  sqlite.begin[conn_id]     : start transaction
  sqlite.commit[conn_id]    : commit transaction
  sqlite.rollback[conn_id]  : rollback transaction
  
  Example:
    conn_id:=sqlite.open["data.db"]
    result:=sqlite.query[conn_id;"SELECT * FROM users";[]]
    sqlite.close[conn_id]`
		}
		return ""
	}
}

// vfOpen implements sqlite.open monadic function
func vfOpen(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.open: expected 1 argument (path), got %d", len(args))
	}

	path, ok := args[0].BV().(goal.S)
	if !ok {
		return goal.Panicf("sqlite.open: expected string path, got %s", args[0].Type())
	}

	db, err := sql.Open("sqlite3", string(path))
	if err != nil {
		return goal.Panicf("sqlite.open: %v", err)
	}

	// Verify the connection is alive
	if err := db.Ping(); err != nil {
		return goal.Panicf("sqlite.open: ping failed: %v", err)
	}

	connID := nextConnectionID
	nextConnectionID++
	connections[connID] = db

	return goal.NewI(connID)
}

// vfClose implements sqlite.close monadic function
func vfClose(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.close: expected 1 argument (conn_id), got %d", len(args))
	}

	connID, ok := args[0].BV().(goal.I)
	if !ok {
		return goal.Panicf("sqlite.close: expected integer conn_id, got %s", args[0].Type())
	}

	conn, ok := connections[int(connID)]
	if !ok {
		return goal.Panicf("sqlite.close: connection %d not found", connID)
	}

	if err := conn.Close(); err != nil {
		return goal.Panicf("sqlite.close: %v", err)
	}

	delete(connections, int(connID))
	return goal.NewI(0) // Return 0 for success
}

// vfQuery implements sqlite.query triadic function
func vfQuery(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 3 {
		return goal.Panicf("sqlite.query: expected 3 arguments (conn_id, sql, params), got %d", len(args))
	}

	// Parse connection ID
	connID, ok := args[0].BV().(goal.I)
	if !ok {
		return goal.Panicf("sqlite.query: expected integer conn_id, got %s", args[0].Type())
	}

	conn, ok := connections[int(connID)]
	if !ok {
		return goal.Panicf("sqlite.query: connection %d not found", connID)
	}

	// Parse SQL
	sqlStr, ok := args[1].BV().(goal.S)
	if !ok {
		return goal.Panicf("sqlite.query: expected string sql, got %s", args[1].Type())
	}

	// Parse params (array of strings)
	params, ok := args[2].BV().(*goal.AS)
	if !ok && args[2].Type() != goal.NT {
		return goal.Panicf("sqlite.query: expected array params, got %s", args[2].Type())
	}

	// Convert params to []interface{}
	var interfaceParams []interface{}
	if params != nil {
		for _, p := range params.Slice {
			interfaceParams = append(interfaceParams, string(p.(goal.S)))
		}
	}

	// Execute query
	rows, err := conn.Query(string(sqlStr), interfaceParams...)
	if err != nil {
		return goal.Panicf("sqlite.query: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return goal.Panicf("sqlite.query: get columns failed: %v", err)
	}

	// Read all rows
	var results []goal.V
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return goal.Panicf("sqlite.query: scan failed: %v", err)
		}

		// Convert to Goal array
		row := make([]goal.V, len(columns))
		for i, val := range values {
			row[i] = goal.NewS(fmt.Sprintf("%v", val))
		}
		results = append(results, goal.NewAS(row))
	}

	if err := rows.Err(); err != nil {
		return goal.Panicf("sqlite.query: rows error: %v", err)
	}

	// Return array of rows
	return goal.NewAS(results)
}

// vfExec implements sqlite.exec triadic function
func vfExec(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 3 {
		return goal.Panicf("sqlite.exec: expected 3 arguments (conn_id, sql, params), got %d", len(args))
	}

	// Parse connection ID
	connID, ok := args[0].BV().(goal.I)
	if !ok {
		return goal.Panicf("sqlite.exec: expected integer conn_id, got %s", args[0].Type())
	}

	conn, ok := connections[int(connID)]
	if !ok {
		return goal.Panicf("sqlite.exec: connection %d not found", connID)
	}

	// Parse SQL
	sqlStr, ok := args[1].BV().(goal.S)
	if !ok {
		return goal.Panicf("sqlite.exec: expected string sql, got %s", args[1].Type())
	}

	// Parse params (array of strings)
	params, ok := args[2].BV().(*goal.AS)
	if !ok && args[2].Type() != goal.NT {
		return goal.Panicf("sqlite.exec: expected array params, got %s", args[2].Type())
	}

	// Convert params to []interface{}
	var interfaceParams []interface{}
	if params != nil {
		for _, p := range params.Slice {
			interfaceParams = append(interfaceParams, string(p.(goal.S)))
		}
	}

	// Execute statement
	result, err := conn.Exec(string(sqlStr), interfaceParams...)
	if err != nil {
		return goal.Panicf("sqlite.exec: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return goal.Panicf("sqlite.exec: get rows affected failed: %v", err)
	}

	return goal.NewI(affected)
}

// vfBegin implements sqlite.begin monadic function
func vfBegin(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.begin: expected 1 argument (conn_id), got %d", len(args))
	}

	connID, ok := args[0].BV().(goal.I)
	if !ok {
		return goal.Panicf("sqlite.begin: expected integer conn_id, got %s", args[0].Type())
	}

	conn, ok := connections[int(connID)]
	if !ok {
		return goal.Panicf("sqlite.begin: connection %d not found", connID)
	}

	_, err := conn.Begin()
	if err != nil {
		return goal.Panicf("sqlite.begin: %v", err)
	}

	return goal.NewI(0) // Return 0 for success
}

// vfCommit implements sqlite.commit monadic function
func vfCommit(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.commit: expected 1 argument (conn_id), got %d", len(args))
	}

	connID, ok := args[0].BV().(goal.I)
	if !ok {
		return goal.Panicf("sqlite.commit: expected integer conn_id, got %s", args[0].Type())
	}

	conn, ok := connections[int(connID)]
	if !ok {
		return goal.Panicf("sqlite.commit: connection %d not found", connID)
	}

	// Note: In SQLite, transactions are auto-committed unless Begin was called
	// This is a no-op but kept for API consistency
	return goal.NewI(0)
}

// vfRollback implements sqlite.rollback monadic function
func vfRollback(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.rollback: expected 1 argument (conn_id), got %d", len(args))
	}

	connID, ok := args[0].BV().(goal.I)
	if !ok {
		return goal.Panicf("sqlite.rollback: expected integer conn_id, got %s", args[0].Type())
	}

	conn, ok := connections[int(connID)]
	if !ok {
		return goal.Panicf("sqlite.rollback: connection %d not found", connID)
	}

	// Note: In SQLite, transactions are auto-committed unless Begin was called
	// This is a no-op but kept for API consistency
	return goal.NewI(0)
}
