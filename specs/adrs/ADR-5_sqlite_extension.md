# ADR-5: SQLite Database via Goal Extension

**Status**: Accepted  
**Date**: 2025-01-XX  
**Author**: User

---

## Context
Goal has no database driver. We need persistent storage for:
- Users (Strava tokens)
- Activities (fetched from Strava)
- Pre-aggregated statistics

SQLite is ideal because:
- Zero-configuration (single file)
- Embedded (no server process)
- ACID-compliant
- Supports all our query needs

## Decision
Build a **Goal extension in Go** that wraps `github.com/mattn/go-sqlite3` to provide:
1. **Database connection management** (open/close)
2. **Query execution** (SELECT, INSERT, UPDATE, DELETE)
3. **Parameterized queries** (prevent SQL injection)
4. **Transaction support**

## Alternatives Considered
1. **Use `sqlite3` CLI via shell**:
   - Slow (spawns process per query)
   - Hard to handle results
   - Fragile
2. **Flat files (JSON/CSV)**:
   - No querying capabilities
   - Hard to aggregate (weekly stats)
   - No transactions
3. **PostgreSQL**:
   - Overkill for local single-user app
   - Requires server setup

## Consequences
- ✅ Full SQL support (queries, joins, aggregations)
- ✅ Transaction-safe (important for sync operations)
- ✅ Embedded (portable, easy to deploy)
- ⚠️ Requires Go SQLite driver dependency
- ⚠️ Go extension needed (but straightforward)

## Implementation
The extension will provide these Goal functions:
```
# Open database connection
db.open(path string) -> connection

# Execute query with parameters
connection.query(sql string, params []interface{}) -> (results [][]interface{}, error string)

# Execute statement (INSERT/UPDATE/DELETE)
connection.exec(sql string, params []interface{}) -> (affected int, error string)

# Begin transaction
connection.begin() -> error

# Commit transaction
connection.commit() -> error

# Rollback transaction
connection.rollback() -> error

# Close connection
connection.close() -> error
```

### Go Implementation (extensions/sqlite/sqlite.go)
- Use `github.com/mattn/go-sqlite3`
- Handle connection pooling
- Convert between Goal and Go types
- Return errors clearly

### Example Usage
```goal
conn := db.open("strava.db")

# Insert user
conn.exec(
    "INSERT INTO users (strava_id, api_token) VALUES (?, ?)",
    [12345, "token_xyz"]
)

# Query activities
results := conn.query(
    "SELECT * FROM activities WHERE user_id = ? AND type = 'Run'",
    [1]
)

conn.close()
```

---

## Schema
See main spec (Section 5.2) for table definitions.

---

## Related
- Required by ADR-1 (Goal language)
- Blocks: DB-1 (schema), SYNC-1 (sync logic), STAT-1/2/3 (statistics)
