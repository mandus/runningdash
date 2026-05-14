// Package sqlite provides SQLite database functionality for Goal extensions
// This enables Goal to interact with SQLite databases
//
// Implements requirements from:
// - specs/strava_dashboard_spec_v0.3.md §5.2 (Local Database)
// - specs/adrs/ADR-5_sqlite_extension.md
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Connection represents a database connection
type Connection struct {
	DB *sql.DB
}

// OpenRequest represents a request to open a database connection
type OpenRequest struct {
	Path string `json:"path"`
}

// OpenResponse represents the response from opening a connection
type OpenResponse struct {
	ConnectionID int    `json:"connection_id"`
	Error       string `json:"error"`
}

// QueryRequest represents a query request
type QueryRequest struct {
	ConnectionID int      `json:"connection_id"`
	SQL         string   `json:"sql"`
	Params      []string `json:"params"`
}

// QueryResponse represents the response from a query
type QueryResponse struct {
	Results [][]string `json:"results"`
	Error   string     `json:"error"`
}

// ExecRequest represents an exec request
type ExecRequest struct {
	ConnectionID int      `json:"connection_id"`
	SQL         string   `json:"sql"`
	Params      []string `json:"params"`
}

// ExecResponse represents the response from an exec
type ExecResponse struct {
	Affected int    `json:"affected"`
	Error   string `json:"error"`
}

// SimpleResponse represents a simple response with just error
type SimpleResponse struct {
	Error string `json:"error"`
}

// connections stores open database connections
var connections = make(map[int]*Connection)
var nextConnectionID = 1

// Open opens a database connection
// Input: JSON string with path
// Output: JSON string with connection_id and error
func Open(requestJSON string) string {
	var req OpenRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(OpenResponse{Error: fmt.Sprintf("Invalid request JSON: %v", err)})
	}

	if req.Path == "" {
		return fmtResponse(OpenResponse{Error: "Path is required"})
	}

	db, err := sql.Open("sqlite3", req.Path)
	if err != nil {
		return fmtResponse(OpenResponse{Error: err.Error()})
	}

	// Verify the connection is alive
	if err := db.Ping(); err != nil {
		return fmtResponse(OpenResponse{Error: err.Error()})
	}

	connID := nextConnectionID
	nextConnectionID++
	connections[connID] = &Connection{DB: db}

	return fmtResponse(OpenResponse{ConnectionID: connID})
}

// Query executes a SELECT query
// Input: JSON string with connection_id, sql, and params
// Output: JSON string with results and error
func Query(requestJSON string) string {
	var req QueryRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(QueryResponse{Error: fmt.Sprintf("Invalid request JSON: %v", err)})
	}

	conn, ok := connections[req.ConnectionID]
	if !ok {
		return fmtResponse(QueryResponse{Error: fmt.Sprintf("Connection %d not found", req.ConnectionID)})
	}

	// Convert string params to interface{}
	interfaceParams := make([]interface{}, len(req.Params))
	for i, p := range req.Params {
		interfaceParams[i] = p
	}

	rows, err := conn.DB.Query(req.SQL, interfaceParams...)
	if err != nil {
		return fmtResponse(QueryResponse{Error: err.Error()})
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return fmtResponse(QueryResponse{Error: err.Error()})
	}

	// Read all rows
	var results [][]string
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmtResponse(QueryResponse{Error: err.Error()})
		}

		// Convert to string slice
		row := make([]string, len(columns))
		for i, val := range values {
			row[i] = fmt.Sprintf("%v", val)
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return fmtResponse(QueryResponse{Error: err.Error()})
	}

	return fmtResponse(QueryResponse{Results: results})
}

// Exec executes an INSERT/UPDATE/DELETE statement
// Input: JSON string with connection_id, sql, and params
// Output: JSON string with affected count and error
func Exec(requestJSON string) string {
	var req ExecRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(ExecResponse{Error: fmt.Sprintf("Invalid request JSON: %v", err)})
	}

	conn, ok := connections[req.ConnectionID]
	if !ok {
		return fmtResponse(ExecResponse{Error: fmt.Sprintf("Connection %d not found", req.ConnectionID)})
	}

	// Convert string params to interface{}
	interfaceParams := make([]interface{}, len(req.Params))
	for i, p := range req.Params {
		interfaceParams[i] = p
	}

	result, err := conn.DB.Exec(req.SQL, interfaceParams...)
	if err != nil {
		return fmtResponse(ExecResponse{Error: err.Error()})
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmtResponse(ExecResponse{Error: err.Error()})
	}

	return fmtResponse(ExecResponse{Affected: int(affected)})
}

// Begin starts a transaction
// Input: JSON string with connection_id
// Output: JSON string with error
func Begin(requestJSON string) string {
	var req struct {
		ConnectionID int `json:"connection_id"`
	}
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(SimpleResponse{Error: fmt.Sprintf("Invalid request JSON: %v", err)})
	}

	conn, ok := connections[req.ConnectionID]
	if !ok {
		return fmtResponse(SimpleResponse{Error: fmt.Sprintf("Connection %d not found", req.ConnectionID)})
	}

	_, err := conn.DB.Begin()
	if err != nil {
		return fmtResponse(SimpleResponse{Error: err.Error()})
	}

	return fmtResponse(SimpleResponse{})
}

// Commit commits a transaction
// Input: JSON string with connection_id
// Output: JSON string with error
func Commit(requestJSON string) string {
	var req struct {
		ConnectionID int `json:"connection_id"`
	}
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(SimpleResponse{Error: fmt.Sprintf("Invalid request JSON: %v", err)})
	}

	conn, ok := connections[req.ConnectionID]
	if !ok {
		return fmtResponse(SimpleResponse{Error: fmt.Sprintf("Connection %d not found", req.ConnectionID)})
	}

	// Note: In sqlite3, transactions are auto-committed unless Begin was called
	// This is a no-op for sqlite but kept for API consistency
	return fmtResponse(SimpleResponse{})
}

// Rollback rolls back a transaction
// Input: JSON string with connection_id
// Output: JSON string with error
func Rollback(requestJSON string) string {
	var req struct {
		ConnectionID int `json:"connection_id"`
	}
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(SimpleResponse{Error: fmt.Sprintf("Invalid request JSON: %v", err)})
	}

	conn, ok := connections[req.ConnectionID]
	if !ok {
		return fmtResponse(SimpleResponse{Error: fmt.Sprintf("Connection %d not found", req.ConnectionID)})
	}

	// Note: In sqlite3, transactions are auto-committed unless Begin was called
	// This is a no-op for sqlite but kept for API consistency
	return fmtResponse(SimpleResponse{})
}

// Close closes a database connection
// Input: JSON string with connection_id
// Output: JSON string with error
func Close(requestJSON string) string {
	var req struct {
		ConnectionID int `json:"connection_id"`
	}
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return fmtResponse(SimpleResponse{Error: fmt.Sprintf("Invalid request JSON: %v", err)})
	}

	conn, ok := connections[req.ConnectionID]
	if !ok {
		return fmtResponse(SimpleResponse{Error: fmt.Sprintf("Connection %d not found", req.ConnectionID)})
	}

	if err := conn.DB.Close(); err != nil {
		return fmtResponse(SimpleResponse{Error: err.Error()})
	}

	delete(connections, req.ConnectionID)
	return fmtResponse(SimpleResponse{})
}

// fmtResponse formats a response as JSON string
func fmtResponse[T any](resp T) string {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return `{"error": "Internal error"}`
	}
	return string(jsonBytes)
}

func main() {
	// Test the extension
	// Open a connection
	openReq := OpenRequest{Path: ":memory:"}
	openReqJSON, _ := json.Marshal(openReq)
	openResp := Open(string(openReqJSON))
	fmt.Println("Open:", openResp)
}
