# Specification: Strava Training Dashboard

**Status**: Draft  
**Author**: mandus  
**Date**: 2026-05-14  
**Version**: 0.3  

---

## 1. Overview
- **Purpose**: Build a local system that syncs Strava training activities for a single user, calculates basic running statistics (weekly km, pace), and displays them in a dashboard. This is the **minimum viable foundation** for future expansions (trainer plans, advanced stats).
- **Stakeholders**: End user (runner), Developer, Strava (API provider)
- **Dependencies**: Strava API access, API token for target user

---

## 2. Scope

### In Scope
✅ **Core**
- Authenticate with Strava API using a pre-configured user token
- Fetch all historical training activities for that user
- Store activities in a local database
- Calculate weekly kilometers run
- Calculate average pace per activity (in min/km)
- Display a dashboard with:
  - Weekly km (aggregated)
  - Pace trends (per activity)
  - Basic progress visualization (e.g., weekly bar chart)

✅ **Technical**
- Database schema for activities
- API client for Strava with pagination and rate limit handling
- Sync mechanism (manual trigger for v1)
- Basic dashboard UI (web interface)

✅ **Error Handling**
- Log all API errors
- Implement retry logic with exponential backoff (max 3 retries)
- Detect and handle Strava activity deletions by flagging records as `is_deleted=TRUE`

### Out of Scope
❌ Trainer plan generation
❌ Multi-user support
❌ Advanced statistics (VO2 max, heart rate zones, etc.)
❌ Real-time sync (polling only)
❌ Mobile app or native UI
❌ **Progress goals**: User-specified targets (e.g., "run 50km this month") will be added via Web UI in a future iteration

---

## 3. Non-Functional Requirements
- **Privacy**: Store Strava token securely (encrypted or env vars)
- **Performance**: Sync < 10k activities in < 5 minutes
- **Reliability**: Handle API rate limits (Strava: 100 req/15 min) with backoff
- **Data Retention**: Store all historical activities; no auto-deletion
- **Accuracy**: 
  - Pace (min/km) calculated as `(moving_time / distance) * 1000 / 60`
  - Distance displayed in km (stored in meters, converted on display)
- **Portability**: Local DB (SQLite) for easy setup; migratable to PostgreSQL later
- **Token Management**: System must support long-lived Strava refresh tokens with initial OAuth2 setup
- **Testing**: All code must have automated tests with minimum 80% coverage; tests must pass in CI before merge

---

## 4. Domain Model

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│   StravaUser    │       │ TrainingActivity│       │    Statistic    │
├─────────────────┤       ├─────────────────┤       ├─────────────────┤
│ id (PK)         │◄──────│ id (PK)         │       │ id (PK)         │
│ strava_id       │       │ user_id (FK)    │       │ user_id (FK)    │
│ api_token       │       │ strava_id       │       │ week_start      │
│ refresh_token   │       │ type            │       │ metric_name     │
│ token_expires   │       │ name            │       │ value           │
└─────────────────┘       │ distance        │       │ created_at      │
                          │ moving_time     │       └─────────────────┘
                          │ start_date      │
                          │ average_speed   │
                          │ elevation_gain  │
                          │ is_deleted      │
                          └─────────────────┘
```

**Activity Types**: Run, Ride, Swim, etc. (store all, but **only Run activities** are used for running-specific stats in v1)

---

## 5. Interfaces

### 5.1 Strava API
- **Endpoint**: `GET /athlete/activities`
- **Auth**: Bearer token (user-provided) + refresh token support
- **Pagination**: Handle `?page` and `?per_page` (max 200 per request)
- **Fields**: `id`, `type`, `name`, `distance`, `moving_time`, `start_date`, `average_speed`, `elevation_gain`
- **Rate Limiting**: 100 requests/15 min; implement backoff with max 3 retries

### 5.2 Local Database (SQLite)
**Table: `users`**
```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  strava_id INTEGER UNIQUE,
  api_token TEXT NOT NULL,
  refresh_token TEXT,
  token_expires TIMESTAMP
);
```

**Table: `activities`**
```sql
CREATE TABLE activities (
  id INTEGER PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  strava_id INTEGER UNIQUE,
  type TEXT NOT NULL,
  name TEXT,
  distance REAL,       -- meters
  moving_time INTEGER, -- seconds
  start_date TIMESTAMP,
  average_speed REAL,  -- m/s
  elevation_gain REAL, -- meters
  is_deleted BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Table: `stats`** *(pre-aggregated for dashboard)*
```sql
CREATE TABLE stats (
  id INTEGER PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  week_start DATE NOT NULL,
  metric_name TEXT NOT NULL,  -- 'weekly_distance_km', 'avg_pace_min_per_km', 'activity_count'
  value REAL,
  UNIQUE(user_id, week_start, metric_name)
);
```

### 5.3 Dashboard (Web UI)
- **Output**: Simple web interface (HTML + lightweight framework)
- **Metrics**:
  - Weekly distance (km) for last 12 weeks
  - Average pace (min/km) per activity
  - Activity count per week
- **Time Range**: Last 12 weeks (configurable)
- **Future**: User-specified progress goals (e.g., target km per week) via UI

---

## 6. Business Rules

| ID | Rule | Given | When | Then |
|----|------|-------|------|------|
| BR-1 | Only fetch for authorized user | User has valid token | Sync triggered | Fetch activities for that user only |
| BR-2 | Filter activity types | Activity type is "Run" | On stats calculation | Include in running-specific stats; store all activity types |
| BR-3 | Calculate pace | Activity has `distance` > 0 and `moving_time` > 0 | On stats update | `pace_min_per_km = (moving_time / distance) * 1000 / 60` |
| BR-4 | Handle duplicates | Activity `strava_id` already exists | On fetch | Skip insert, update existing record if data differs, log warning |
| BR-5 | Unit conversion | Activity distance in meters | On display | `distance_km = distance / 1000` |
| BR-6 | Handle deletions | Strava activity no longer exists in API | On sync | Set `is_deleted=TRUE` in DB, exclude from stats |
| BR-7 | Rate limit handling | API returns 429 status | On request | Wait and retry with exponential backoff (max 3 attempts) |

---

## 7. Acceptance Criteria

### Sync & Storage
- [ ] System accepts a Strava API token and refresh token via config/environment
- [ ] System fetches all historical activities for the user with pagination
- [ ] Activities are stored in the local database with all required fields
- [ ] Duplicate activities (same `strava_id`) are updated, not duplicated
- [ ] Deleted Strava activities are flagged as `is_deleted=TRUE`
- [ ] API rate limits are respected with automatic backoff

### Statistics
- [ ] Weekly running distance (km) is calculated correctly for Run activities only
- [ ] Average pace (min/km) per activity is calculated using BR-3 formula
- [ ] Stats are pre-aggregated in the `stats` table for the dashboard
- [ ] Distance is converted from meters to km for display

### Dashboard
- [ ] Dashboard displays weekly km for the last 12 weeks
- [ ] Dashboard displays average pace per activity
- [ ] Dashboard shows total activity count per week
- [ ] Dashboard is accessible via a web browser

### Error Handling
- [ ] API errors are logged with timestamps
- [ ] Failed requests are retried up to 3 times with backoff
- [ ] Token expiration triggers refresh flow (if refresh token present)

### Testing
- [ ] All calculation functions (pace, weekly distance) have unit tests
- [ ] API sync logic has integration tests verifying data flow from Strava → DB
- [ ] Dashboard rendering has E2E tests verifying UI displays correct stats
- [ ] Minimum 80% code coverage for backend (Goal) and frontend (Svelte)
- [ ] All tests pass in CI before merge

---

## 8. Open Questions
1. Should the sync be **manual only** for v1, or include a **basic scheduler**? *(Lean toward manual for v1)*
2. Should the Web UI use a **framework** (e.g., Flask equivalent in Goal) or serve **static HTML + JS**?
3. Should we store **raw JSON** from Strava alongside parsed fields for debugging?
4. What is the **minimum supported Goal version**?

---

## 9. Changelog
- 2025-01-XX: Initial draft (v0.1)
- 2025-01-XX: Applied fixes from review (v0.2)
  - Corrected pace calculation formula
  - Added unit conversion rule
  - Added error handling and token management NFRs
  - Clarified "progress" scope (user-specified goals via future Web UI)
  - Added `is_deleted` flag to activities table
  - Added `refresh_token` to users table
- 2025-01-XX: Added testing requirements (v0.3)
  - Added "Testable Code" principle to constitution
  - Added "No Merge Without Tests" governance rule
  - Updated Coding Standards with specific test requirements
  - Added Testing section to acceptance criteria
  - Added Testing to Non-Functional Requirements
  - Added tests/ directory to file structure

---

**Next Steps**:
1. Review v0.3 for any remaining issues
2. Begin implementation with Phase 0 (Foundation)
3. Update constitution and spec as implementation reveals gaps
