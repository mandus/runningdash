# ADR-2: HTTP Requests via Goal Extension

**Status**: Accepted  
**Date**: 2025-01-XX  
**Author**: User

---

## Context
Goal has no built-in HTTP client. To call the Strava API (and serve web endpoints), we need HTTP capabilities. Current workaround is shelling out to `curl`, but this is fragile for production use.

## Decision
Build a **Goal extension in Go** that wraps Go's `net/http` package to provide:
1. **HTTP GET** with headers (for Strava API calls)
2. **HTTP POST** (for future OAuth2 token refresh)
3. **Response parsing** (status code, body, headers)
4. **Error handling** for network failures and HTTP errors

## Alternatives Considered
1. **Continue using `curl` via shell**: Simpler short-term, but:
   - Fragile (depends on curl being installed)
   - Harder to handle errors/headers
   - Less portable
2. **Use Go's `http` package directly via Goal's Go interop**: More complex, less Goal-like

## Consequences
- ✅ Cleaner, more maintainable than shelling to `curl`
- ✅ Reusable for any HTTP needs (Strava API, future integrations)
- ✅ Native Go performance and reliability
- ⚠️ Requires Go knowledge to build and maintain
- ⚠️ Adds build complexity (need Go toolchain for extensions)

## Implementation
The extension will provide these Goal functions:
```
# Make a GET request
http.get(url string, headers map[string]string) -> (status int, body string, error string)

# Make a POST request  
http.post(url string, headers map[string]string, body string) -> (status int, body string, error string)
```

### Go Implementation (extensions/http/http.go)
- Use `net/http` for requests
- Handle timeouts (default: 30 seconds)
- Return structured errors

### Goal Bindings (extensions/http/http.goal)
- Expose Go functions to Goal
- Handle type conversions between Goal and Go

---

## Related
- Required by ADR-1 (Goal language) and ADR-3 (Web serving)
- Blocks: API-1 (Strava API client), API-2 (rate limit handling)
