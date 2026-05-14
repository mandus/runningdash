# ADR-1: Goal as Primary Language

**Status**: Accepted  
**Date**: 2025-01-XX  
**Author**: User

---

## Context
The project must use Goal (https://codeberg.org/anaseto/goal) as the primary programming language, per user requirement. Goal is a bare-bones language built in Go, with minimal standard library support.

## Decision
Use **Goal** for all core backend logic:
- Sync logic (fetching Strava activities)
- Statistics calculations (pace, weekly distance)
- Database interactions
- Business rule enforcement

## Alternatives Considered
1. **Python/Flask**: More mature ecosystem, but violates user requirement
2. **Go (pure)**: More capable for HTTP/web, but user wants Goal
3. **Shell scripts + curl**: Too fragile for production

## Consequences
- ✅ Aligns with user's tech stack and requirements
- ⚠️ No built-in HTTP client → requires workarounds (curl or extensions)
- ⚠️ Limited standard library → may need custom extensions for JSON, SQLite, etc.
- ⚠️ Smaller community → fewer resources for troubleshooting

## Related
- Requires ADR-2 (HTTP extension) to handle API calls
- Requires ADR-5 (SQLite extension) for database access
- Requires ADR-3 (Web serving) for HTTP server

---

## Implementation Notes
- Goal code will live in `/goal/` directory
- Extensions will be Go packages in `/extensions/`
- Goal files use `.goal` extension
