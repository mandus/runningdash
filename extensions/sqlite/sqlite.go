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
// Map key is connection ID (int64), value is *sql.DB
var connections = make(map[int64]*sql.DB)
var nextConnectionID int64 = 1

// Import registers the SQLite extension functions with the Goal context.
// The following functions are registered:
//
//	sqlite.open[path] : open database connection, returns connection_id
//	sqlite.close[conn_id] : close database connection
//	sqlite.query[request] : execute SELECT query, returns result rows
//	sqlite.exec[request] : execute INSERT/UPDATE/DELETE, returns affected count
//	sqlite.begin[conn_id] : start transaction
//	sqlite.commit[conn_id] : commit transaction
//	sqlite.rollback[conn_id] : rollback transaction
//
// For query and exec, request is a dict with keys:
//   conn_id (int): connection ID
//   sql (string): SQL query
//   params (array of strings, optional): query parameters
//
// Connection IDs are integers. All functions that need a connection take conn_id as a parameter.
func Import(ctx *goal.Context, pfx string) {
	if pfx != "" {
		pfx += "."
	}

	// Register sqlite.open as a monad (takes path string, returns connection_id int)
	ctx.AssignGlobal(pfx+"sqlite.open", ctx.RegisterMonad("."+pfx+"sqlite.open", vfOpen))

	// Register sqlite.close as a monad (takes connection_id int)
	ctx.AssignGlobal(pfx+"sqlite.close", ctx.RegisterMonad("."+pfx+"sqlite.close", vfClose))

	// Register sqlite.query as a monad (takes dict with conn_id, sql, params)
	ctx.AssignGlobal(pfx+"sqlite.query", ctx.RegisterMonad("."+pfx+"sqlite.query", vfQuery))

	// Register sqlite.exec as a monad (takes dict with conn_id, sql, params)
	ctx.AssignGlobal(pfx+"sqlite.exec", ctx.RegisterMonad("."+pfx+"sqlite.exec", vfExec))

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
  sqlite.query[request]    : execute SELECT, request is dict with conn_id, sql, params
  sqlite.exec[request]     : execute INSERT/UPDATE/DELETE, request is dict with conn_id, sql, params
  sqlite.begin[conn_id]     : start transaction
  sqlite.commit[conn_id]    : commit transaction
  sqlite.rollback[conn_id]  : rollback transaction
  
  Example:
    conn_id:=sqlite.open["data.db"]
    result:=sqlite.query[(,"conn_id")!,$conn_id;(,"sql")!,"SELECT * FROM users";(,"params")!,[]]
    sqlite.close[conn_id]`
		}
		return ""
	}
}

// getInt extracts an int64 from a Goal value, panicking if not an integer
func getInt(v goal.V, name string) int64 {
	if !v.IsI() {
		goal.Panicf("%s: expected integer, got %s", name, v.Type())
	}
	return v.I()
}

// getString extracts a string from a Goal value, panicking if not a string
func getString(v goal.V, name string) string {
	s, ok := v.BV().(goal.S)
	if !ok {
		goal.Panicf("%s: expected string, got %s", name, v.Type())
	}
	return string(s)
}

// vfOpen implements sqlite.open monadic function
func vfOpen(_ *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.open: expected 1 argument (path), got %d", len(args))
	}

	path := getString(args[0], "sqlite.open")

	db, err := sql.Open("sqlite3", path)
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
func vfClose(_ *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.close: expected 1 argument (conn_id), got %d", len(args))
	}

	connID := getInt(args[0], "sqlite.close")

	conn, ok := connections[connID]
	if !ok {
		return goal.Panicf("sqlite.close: connection %d not found", connID)
	}

	if err := conn.Close(); err != nil {
		return goal.Panicf("sqlite.close: %v", err)
	}

	delete(connections, connID)
	return goal.NewI(0) // Return 0 for success
}

// vfQuery implements sqlite.query monadic function
func vfQuery(_ *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.query: expected 1 argument (request dict), got %d", len(args))
	}

	request, ok := args[0].BV().(*goal.D)
	if !ok {
		return goal.Panicf("sqlite.query: expected dict argument, got %s", args[0].Type())
	}

	// Parse connection ID
	connIDVal, ok := request.GetS("conn_id")
	if !ok {
		return goal.Panicf("sqlite.query: missing required field: conn_id")
	}
	connID := getInt(connIDVal, "sqlite.query conn_id")

	conn, ok := connections[connID]
	if !ok {
		return goal.Panicf("sqlite.query: connection %d not found", connID)
	}

	// Parse SQL
	sqlVal, ok := request.GetS("sql")
	if !ok {
		return goal.Panicf("sqlite.query: missing required field: sql")
	}
	sqlStr := getString(sqlVal, "sqlite.query sql")

	// Parse params (array of strings, optional)
	// Goal note: single strings are passed as-is; arrays are space-separated values.
	// In Goal, "a b c" creates an array of strings, while "a" is just a string.
	// Handle both cases: if params is a string, treat it as a single-element array.
	var interfaceParams []interface{}
	if paramsVal, ok := request.GetS("params"); ok {
		if params, ok := paramsVal.BV().(*goal.AS); ok {
			// Array of strings
			for i := 0; i < params.Len(); i++ {
				p := params.At(i)
				if pS, ok := p.BV().(goal.S); ok {
					interfaceParams = append(interfaceParams, string(pS))
				}
			}
		} else if params, ok := paramsVal.BV().(*goal.AV); ok {
			// Array of values - extract strings
			for i := 0; i < params.Len(); i++ {
				p := params.At(i)
				if pS, ok := p.BV().(goal.S); ok {
					interfaceParams = append(interfaceParams, string(pS))
				}
			}
		} else if paramsS, ok := paramsVal.BV().(goal.S); ok {
			// Single string - treat as single-element array
			interfaceParams = append(interfaceParams, string(paramsS))
		}
	}

	// Execute query
	rows, err := conn.Query(sqlStr, interfaceParams...)
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

		// Convert to Goal array of strings
		rowStrings := make([]string, len(columns))
		for i, val := range values {
			rowStrings[i] = fmt.Sprintf("%v", val)
		}
		results = append(results, goal.NewAS(rowStrings))
	}

	if err := rows.Err(); err != nil {
		return goal.Panicf("sqlite.query: rows error: %v", err)
	}

	// Return array of rows (array of arrays of strings)
	return goal.NewAV(results)
}

// vfExec implements sqlite.exec monadic function
func vfExec(_ *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.exec: expected 1 argument (request dict), got %d", len(args))
	}

	request, ok := args[0].BV().(*goal.D)
	if !ok {
		return goal.Panicf("sqlite.exec: expected dict argument, got %s", args[0].Type())
	}

	// Parse connection ID
	connIDVal, ok := request.GetS("conn_id")
	if !ok {
		return goal.Panicf("sqlite.exec: missing required field: conn_id")
	}
	connID := getInt(connIDVal, "sqlite.exec conn_id")

	conn, ok := connections[connID]
	if !ok {
		return goal.Panicf("sqlite.exec: connection %d not found", connID)
	}

	// Parse SQL
	sqlVal, ok := request.GetS("sql")
	if !ok {
		return goal.Panicf("sqlite.exec: missing required field: sql")
	}
	sqlStr := getString(sqlVal, "sqlite.exec sql")

	// Parse params (array of strings, optional)
	// Goal note: single strings are passed as-is; arrays are space-separated values.
	// Handle both cases: if params is a string, treat it as a single-element array.
	var interfaceParams []interface{}
	if paramsVal, ok := request.GetS("params"); ok {
		if params, ok := paramsVal.BV().(*goal.AS); ok {
			// Array of strings
			for i := 0; i < params.Len(); i++ {
				p := params.At(i)
				if pS, ok := p.BV().(goal.S); ok {
					interfaceParams = append(interfaceParams, string(pS))
				}
			}
		} else if params, ok := paramsVal.BV().(*goal.AV); ok {
			// Array of values - extract strings
			for i := 0; i < params.Len(); i++ {
				p := params.At(i)
				if pS, ok := p.BV().(goal.S); ok {
					interfaceParams = append(interfaceParams, string(pS))
				}
			}
		} else if paramsS, ok := paramsVal.BV().(goal.S); ok {
			// Single string - treat as single-element array
			interfaceParams = append(interfaceParams, string(paramsS))
		}
	}

	// Execute statement
	result, err := conn.Exec(sqlStr, interfaceParams...)
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
func vfBegin(_ *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.begin: expected 1 argument (conn_id), got %d", len(args))
	}

	connID := getInt(args[0], "sqlite.begin")

	conn, ok := connections[connID]
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
func vfCommit(_ *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.commit: expected 1 argument (conn_id), got %d", len(args))
	}

	connID := getInt(args[0], "sqlite.commit")

	if _, ok := connections[connID]; !ok {
		return goal.Panicf("sqlite.commit: connection %d not found", connID)
	}

	// Note: In SQLite, transactions are auto-committed unless Begin was called
	// This is a no-op but kept for API consistency
	return goal.NewI(0)
}

// vfRollback implements sqlite.rollback monadic function
func vfRollback(_ *goal.Context, args []goal.V) goal.V {
	if len(args) != 1 {
		return goal.Panicf("sqlite.rollback: expected 1 argument (conn_id), got %d", len(args))
	}

	connID := getInt(args[0], "sqlite.rollback")

	if _, ok := connections[connID]; !ok {
		return goal.Panicf("sqlite.rollback: connection %d not found", connID)
	}

	// Note: In SQLite, transactions are auto-committed unless Begin was called
	// This is a no-op but kept for API consistency
	return goal.NewI(0)
}
