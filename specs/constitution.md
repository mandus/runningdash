# Constitution: Strava Training Dashboard

---

## We Value
1. **Minimal Dependencies** – Prefer Goal's bare-bones approach; only extend when necessary
2. **Pragmatism** – Shell out to `curl` or build Goal extensions as needed, but document tradeoffs
3. **User Privacy** – Strava tokens and activity data are sensitive; handle with care
4. **Maintainability** – Extensions must be reusable and well-documented
5. **Progressive Enhancement** – Start with `curl` for API calls, but build proper Goal extensions for production
6. **Testable Code** – All non-trivial logic must have automated tests; if it's hard to test, refactor it

---

## Principles

| Principle | Description | Rationale |
|-----------|-------------|-----------|
| **Goal-First** | Use Goal for core logic; extend only when stdlib is insufficient | Aligns with project constraints |
| **No Magic** | All external calls (curl, Go extensions) must be explicit | Goal's philosophy; avoids hidden complexity |
| **Separation of Concerns** | Goal handles backend (sync, stats); Svelte handles frontend | Clean boundaries between systems |
| **Document Extensions** | Any Goal extension must include usage examples and limitations | Ensures maintainability |
| **Fail Loud** | Errors from `curl` or Go extensions must propagate clearly | Easier debugging |
| **Testable Code** | All non-trivial logic must have automated tests; if it's hard to test, refactor it | Ensures reliability and maintainability |

---

## Governance

### Decision Making
- **Language/Tool Decisions**: User (project owner) has final say on Goal vs. extensions
- **Scope Changes**: Must update spec *before* implementation (critical for Goal's constraints)
- **Extension Approval**: New Goal extensions require a design doc (what, why, alternatives)

### Change Process
1. **Spec Update**: Modify spec with Goal-specific implementation notes
2. **Extension Design**: For any new Goal capability, write a short design doc
3. **Review**: User approves all extensions
4. **Implementation**: Build extension or use `curl` as stopgap

### Review Requirements
- **All PRs** must reference the relevant spec section
- **Database schema changes** require migration script
- **New statistics** require spec update with calculation formula
- **No Merge Without Tests**: PRs that add or modify code must include tests; exceptions require explicit approval and must be documented in the PR description

---

## Coding Standards

| Area | Standard |
|------|----------|
| **Language** | **Goal** (primary), Go (for extensions), JavaScript/TypeScript (Svelte frontend) |
| **Style** | Follow Goal's idioms (minimal, explicit); Svelte follows its [style guide](https://svelte.dev/docs/fundamentals) |
| **Testing** | **Required**: Unit tests for all calculation logic (stats, pace); Integration tests for API sync; E2E tests for dashboard. Minimum 80% coverage for backend and frontend. |
| **Test Location** | `tests/` directory, mirroring source structure (e.g., `tests/sync_test.goal`, `tests/stats.test.ts`) |
| **Test Framework** | Goal: Custom test scripts per ADR; Svelte: `vitest` with `@testing-library/svelte` |
| **Test Naming** | `[module]_test.goal` or `[module].test.ts` |
| **CI Integration** | All tests must pass in CI before merge |
| **Error Handling** | Goal: Return error codes + messages; never silent failures. Svelte: Display user-friendly errors |
| **Configuration** | Environment variables (`.env` file) for tokens; Goal reads via shell or extension |
| **Dependencies** | Goal: Prefer stdlib; Go extensions only for HTTP/web; Svelte: Minimal npm packages |

---

## Architecture Decisions

All ADRs are documented separately in `specs/adrs/`.

---

## Enforcement
- **Pre-commit hooks**: Run linting/formatting if available for Goal
- **CI/CD**: GitHub Actions to run tests on PRs (when set up)
- **Spec Validation**: Before merging, verify code matches spec acceptance criteria
- **ADR Compliance**: All architecture changes must reference an ADR

---

## File Structure
```
/strava_dashboard
├── specs/
│   ├── strava_dashboard_spec_v0.3.md    # Project specification
│   ├── constitution.md                  # This file
│   └── adrs/                            # Architecture Decision Records
│       ├── ADR-1_goal_language.md
│       ├── ADR-2_http_extension.md
│       ├── ADR-3_web_serving.md
│       ├── ADR-4_svelte_frontend.md
│       └── ADR-5_sqlite_extension.md
├── goal/                               # Goal source files
│   ├── sync.goal                       # Fetch & store activities
│   ├── stats.goal                      # Calculations
│   └── db.goal                         # Database interactions
├── tests/                              # Test files
│   ├── sync_test.goal                  # Sync logic tests
│   ├── stats_test.goal                 # Calculation tests
│   └── integration/                    # Integration tests
├── extensions/                         # Goal extensions (Go packages)
│   ├── http/                           # HTTP client/server
│   │   ├── http.go                     # Go implementation
│   │   └── http.goal                   # Goal bindings
│   ├── sqlite/                         # SQLite wrapper
│   │   ├── sqlite.go
│   │   └── sqlite.goal
│   └── json/                           # JSON parsing
│       ├── json.go
│       └── json.goal
├── frontend/                           # Svelte app
│   ├── src/
│   │   ├── lib/
│   │   │   └── Dashboard.svelte       # Main stats view
│   │   └── App.svelte
│   ├── package.json
│   └── vite.config.js
├── static/                             # Svelte build output
├── .env.example                        # Environment variables template
└── README.md
```
