# ADR-3: Web Serving via Goal + Go HTTP Server

**Status**: Accepted  
**Date**: 2025-01-XX  
**Author**: User

---

## Context
Goal has no web server capabilities. We need to:
1. Serve API endpoints for the dashboard (`/api/sync`, `/api/stats`)
2. Serve static files for the Svelte frontend

## Decision
Build a **Goal extension in Go** that:
1. Starts an HTTP server using Go's `net/http`
2. Exposes Goal functions as HTTP endpoints
3. Serves static files from a directory (for Svelte build output)

## Alternatives Considered
1. **Use `caddy` or `nginx` as reverse proxy**:
   - Adds infrastructure complexity
   - Separates Goal logic from HTTP layer
2. **Shell out to a Python/Node micro-service**:
   - Violates Goal-first principle
   - Adds language complexity

## Consequences
- ✅ Keeps entire backend in Goal/Go ecosystem
- ✅ Tight integration between Goal logic and HTTP layer
- ✅ Single process to run (Goal + extensions)
- ⚠️ Requires careful design for Goal ↔ Go communication
- ⚠️ HTTP server runs in Go, not Goal (but this is acceptable)

## Implementation
The extension will provide:
```
# Start HTTP server on port, with route handlers
http.serve(port int, routes map[string]func) -> error

# Register a route handler
http.route(method string, path string, handler func) -> void

# Serve static files from directory
http.serve_static(path string, dir string) -> void
```

### Example Usage
```goal
# In Goal code
route_handler := func(request) {
    # Call sync logic
    result := sync.activities()
    return http.response(200, result)
}

http.route("GET", "/api/sync", route_handler)
http.serve_static("/", "./static")
http.serve(8080, routes)
```

### Go Implementation (extensions/httpserver/httpserver.go)
- Use `net/http` to create server
- Map Goal functions to HTTP handlers
- Handle request/response conversion

---

## Related
- Required by ADR-1 (Goal language)
- Depends on ADR-2 (HTTP extension) for client-side needs
- Blocks: API-3, API-4, API-5 (all web endpoints)
