# Implementation Tasks: Strava Training Dashboard

**Project**: Strava Training Dashboard (Goal + Svelte)  
**Spec Version**: v0.2  
**Last Updated**: 2025-01-XX

---

## 🎯 **Overview**

This document breaks down the implementation of the Strava Training Dashboard into actionable tasks, organized by phase and priority. All tasks reference the main [specification](../strava_dashboard_spec_v0.2.md) and [constitution](../constitution.md).

---

## 📊 **Task Summary by Phase**

| Phase | Description | Tasks | Priority |
|-------|-------------|-------|----------|
| Phase 0 | Foundation (P0 Extensions) | 5 | P0 |
| Phase 1 | Data Pipeline | 6 | P0 |
| Phase 2 | Statistics Engine | 3 | P0 |
| Phase 3 | Web Backend | 4 | P1 |
| Phase 4 | Frontend (Svelte) | 4 | P1 |
| Phase 5 | Integration & Polish | 5 | P1/P2 |

---

## 🔧 **Phase 0: Foundation (P0 Extensions)**
*Must complete before core development. Blocks all other phases.*

| ID | Task | Description | Spec Ref | Dependencies | Effort | Status |
|----|------|-------------|----------|--------------|--------|--------|
| EXT-1 | Create `http` Goal extension (Go) | Wrap Go's `net/http` for GET/POST requests with headers | ADR-2 | None | Medium | ⬜ |
| EXT-2 | Create `sqlite` Goal extension (Go) | Wrap `github.com/mattn/go-sqlite3` for DB operations | ADR-5 | None | Medium | ⬜ |
| EXT-3 | Create `json` Goal extension (Go) | Basic JSON parsing for Strava API responses | ADR-2 | None | Small | ⬜ |
| INF-1 | Set up Goal development environment | Install Goal, verify toolchain, test extensions | None | None | Small | ⬜ |
| INF-2 | Initialize Svelte frontend | `npm create vite@latest frontend -- --template svelte` | ADR-4 | None | Small | ⬜ |

---

## 🔄 **Phase 1: Data Pipeline**
*Sync Strava activities → Store in SQLite*

| ID | Task | Description | Spec Ref | Dependencies | Effort | Status |
|----|------|-------------|----------|--------------|--------|--------|
| DB-1 | Define SQLite schema | Create `users`, `activities`, `stats` tables (Section 5.2) | Spec §5.2 | EXT-2 | Small | ⬜ |
| CFG-1 | Environment config for tokens | `.env` with `STRAVA_TOKEN`, `STRAVA_REFRESH_TOKEN` | Spec §5.1 | None | Small | ⬜ |
| API-1 | Implement Strava API client in Goal | `GET /athlete/activities` with pagination | Spec §5.1 | EXT-1, EXT-3 | Medium | ⬜ |
| API-2 | Handle rate limits & retries | Exponential backoff on 429 errors (BR-7) | Spec §5.1 | API-1 | Small | ⬜ |
| SYNC-1 | Implement sync logic in Goal | Fetch all historical activities, store in DB | Spec §7 | API-1, DB-1, CFG-1 | Medium | ⬜ |
| SYNC-2 | Handle duplicates & deletions | `is_deleted` flag, update existing records (BR-4, BR-6) | Spec §7 | SYNC-1 | Small | ⬜ |

---

## 📊 **Phase 2: Statistics Engine**
*Calculate weekly km, pace, etc.*

| ID | Task | Description | Spec Ref | Dependencies | Effort | Status |
|----|------|-------------|----------|--------------|--------|--------|
| STAT-1 | Implement pace calculation | `pace_min_per_km = (moving_time / distance) * 1000 / 60` (BR-3) | Spec §6 | SYNC-1 | Small | ⬜ |
| STAT-2 | Implement weekly aggregation | Sum distance, avg pace per week for Run activities (BR-2) | Spec §6 | STAT-1 | Medium | ⬜ |
| STAT-3 | Pre-aggregate stats into `stats` table | Insert weekly metrics for dashboard | Spec §5.2 | STAT-2 | Small | ⬜ |

---

## 🌐 **Phase 3: Web Backend**
*Serve API + static files*

| ID | Task | Description | Spec Ref | Dependencies | Effort | Status |
|----|------|-------------|----------|--------------|--------|--------|
| EXT-4 | Create `httpserver` Goal extension (Go) | Wrap Go's `net/http` for serving routes | ADR-3 | EXT-1 | Medium | ⬜ |
| API-3 | Implement `/api/sync` endpoint | Trigger sync via HTTP | Spec §5.3 | EXT-4, SYNC-1 | Small | ⬜ |
| API-4 | Implement `/api/stats` endpoint | Return JSON: `{weekly_distance_km, avg_pace, activity_count}` | Spec §5.3 | EXT-4, STAT-2 | Small | ⬜ |
| API-5 | Serve static Svelte files | Route `/` → `static/index.html` | Spec §5.3 | EXT-4 | Small | ⬜ |

---

## 🎨 **Phase 4: Frontend (Svelte)**
*Display dashboard*

| ID | Task | Description | Spec Ref | Dependencies | Effort | Status |
|----|------|-------------|----------|--------------|--------|--------|
| FE-1 | Set up Svelte project | Install `svelte-chartjs` for charts | ADR-4 | INF-2 | Small | ⬜ |
| FE-2 | Create dashboard page | Fetch `/api/stats`, display weekly km bar chart | Spec §5.3 | FE-1, API-4 | Medium | ⬜ |
| FE-3 | Add pace trend visualization | Line chart for avg pace over time | Spec §5.3 | FE-2 | Small | ⬜ |
| BUILD-1 | Configure Vite build output | `vite build --outDir ../static` | ADR-4 | FE-1 | Small | ⬜ |

---

## 🔗 **Phase 5: Integration & Polish**

| ID | Task | Description | Spec Ref | Dependencies | Effort | Status |
|----|------|-------------|----------|--------------|--------|--------|
| INT-1 | End-to-end test | Manual test: sync → stats → dashboard | Spec §7 | All P0, P1 | Medium | ⬜ |
| INT-2 | Add CLI sync command | `goal run sync.goal` for manual sync | Spec §2 | SYNC-1 | Small | ⬜ |
| ERR-1 | Improve error handling | User-friendly errors in UI and logs | Spec §7 | API-2, SYNC-1 | Small | ⬜ |
| DOC-1 | Write README | Setup, usage, local dev instructions | None | All | Small | ⬜ |
| TEST-1 | Add Goal script tests | Validate calculations with sample data | Spec §6 | STAT-2 | Medium | ⬜ |

---

## 🗺️ **Dependency Graph**

```
Phase 0 (EXT-1, EXT-2, EXT-3, INF-1, INF-2)
    ↓
Phase 1 (DB-1, CFG-1, API-1, API-2, SYNC-1, SYNC-2)
    ↓
Phase 2 (STAT-1, STAT-2, STAT-3)
    ↓
Phase 3 (EXT-4, API-3, API-4, API-5)
    ↓
Phase 4 (FE-1, FE-2, FE-3, BUILD-1)
    ↓
Phase 5 (INT-1, INT-2, ERR-1, DOC-1, TEST-1)
```

---

## 🎯 **Suggested Order of Attack**

### Option A: Sequential (Recommended for Solo Dev)
1. **Complete Phase 0** (extensions) – Unblocks everything
2. **Complete Phase 1** (data pipeline) – Core functionality
3. **Complete Phase 2** (stats) – Business logic
4. **Complete Phase 3** (backend) – API layer
5. **Complete Phase 4** (frontend) – UI
6. **Complete Phase 5** (polish) – Production readiness

### Option B: Parallel (Faster with Multiple Devs)
- **Track 1 (Backend)**: Phase 0 → Phase 1 → Phase 2 → Phase 3
- **Track 2 (Frontend)**: Phase 0 (INF-2) → Phase 4 (with mock data) → Integrate with Phase 3

---

## 📋 **Task Status Legend**
- ⬜ = Not Started
- 🟡 = In Progress
- ✅ = Complete
- ❌ = Blocked

---

## 📁 **File Structure Reference**

See [constitution.md](../constitution.md) for the complete project structure.

---

## 🔗 **Related Documents**
- [Project Specification](../strava_dashboard_spec_v0.2.md)
- [Constitution](../constitution.md)
- [ADR-1: Goal Language](../adrs/ADR-1_goal_language.md)
- [ADR-2: HTTP Extension](../adrs/ADR-2_http_extension.md)
- [ADR-3: Web Serving](../adrs/ADR-3_web_serving.md)
- [ADR-4: Svelte Frontend](../adrs/ADR-4_svelte_frontend.md)
- [ADR-5: SQLite Extension](../adrs/ADR-5_sqlite_extension.md)
