# AGENTS.md: Guide for AI Agents Working on This Project

**Project**: Strava Training Dashboard  
**Methodology**: Spec-Driven Development (SDD)  
**Spec Kit**: GitHub Spec Kit inspired workflow
**Note**: Always use version-agnostic spec references (e.g., `specs/strava_dashboard_spec_*.md`)

---

## 🤖 **For AI Agents: How to Work With This Project**

This project uses **Spec-Driven Development (SDD)**, a methodology from [GitHub Spec Kit](https://github.github.com/spec-kit/). As an AI agent, you **must** follow the workflow described below.

---

## 📚 **Core Principles**

### 1. **Specifications Are the Source of Truth**
- **NEVER** start coding without reading the spec first
- **NEVER** implement features not in the spec
- **ALWAYS** reference spec sections in your work

### 2. **Constitution Guides Decisions**
- The `constitution.md` defines principles, standards, and governance
- If a decision conflicts with the constitution, **flag it for review**

### 3. **ADRs Document Architecture**
- All architecture decisions are in `specs/adrs/`
- **NEVER** make architectural changes without an ADR
- Reference existing ADRs when implementing

### 4. **Tasks Drive Implementation**
- `specs/tasks.md` contains the current implementation roadmap
- **ALWAYS** work from the highest-priority incomplete task
- Update task status (⬜/🟡/✅) as you work

---

## 📖 **Repository Structure**

```
running/
├── specs/                          # ✅ ALWAYS READ FIRST
│   ├── strava_dashboard_spec_*.md  # Main specification (latest version)
│   ├── constitution.md               # Principles & governance
│   ├── tasks.md                      # Current work items
│   └── adrs/                         # Architecture decisions
│       ├── ADR-1_goal_language.md
│       ├── ADR-2_http_extension.md
│       ├── ADR-3_web_serving.md
│       ├── ADR-4_svelte_frontend.md
│       └── ADR-5_sqlite_extension.md
├── goal/                           # Goal source (TODO)
├── extensions/                     # Goal extensions in Go (TODO)
├── frontend/                       # Svelte app (TODO)
├── .gitignore
├── README.md
└── AGENTS.md                       # This file
```

---

## 🔄 **Required Workflow for Agents**

### Branch & PR Workflow (Mandatory for Multi-Agent Development)

**Before any implementation work:**
1. **Create a feature branch** named after the task:
   ```bash
   git checkout -b feature/[TASK-ID]-[brief-description]
   # Example: feature/EXT-1-http-extension
   # Example: feature/STAT-2-weekly-aggregation
   ```

2. **Reference the task** in your prompt to the agent:
   ```
   "Implement task EXT-1 from specs/tasks.md. Work only on this task.
   Branch: feature/EXT-1-http-extension
   When complete, commit and push the branch but DO NOT merge."
   ```

3. **After implementation:**
   - Commit with message referencing task ID (e.g., "[EXT-1] Implement HTTP Goal extension per ADR-2")
   - Push branch to remote
   - **Create a Pull Request** to `main` with:
     - Title: `[TASK-ID] Description` (e.g., `[EXT-1] HTTP Goal extension`)
     - Description: Summary of changes + "Closes task EXT-1"
     - Reference spec sections and ADRs used

4. **Never commit directly to `main`** — all changes must go through PR review

---

### Before Starting Any Work

1. **Read the spec**: `specs/strava_dashboard_spec_*.md` (latest version)
   - Understand the **purpose**, **scope**, **interfaces**, and **acceptance criteria**
   - Note all **business rules** (BR-1 through BR-7)

2. **Read the constitution**: `specs/constitution.md`
   - Understand **principles**, **governance**, and **coding standards**
   - Note the **tech stack**: Goal (primary), Go (extensions), Svelte (frontend)

3. **Review ADRs**: `specs/adrs/`
   - Understand **why** architectural decisions were made
   - Note **consequences** and **tradeoffs**

4. **Check tasks**: `specs/tasks.md`
   - Find the **next incomplete task** (status = ⬜)
   - Understand **dependencies** before starting

### During Implementation

1. **Reference the spec** in all code comments:
   ```
   // Implements BR-3: pace_min_per_km = (moving_time / distance) * 1000 / 60
   // See: specs/strava_dashboard_spec_*.md §6
   ```

2. **Reference ADRs** for architectural decisions:
   ```
   // Uses Goal extension for HTTP per ADR-2
   // See: specs/adrs/ADR-2_http_extension.md
   ```

3. **Update task status** in `specs/tasks.md`:
   - Change ⬜ → 🟡 when starting
   - Change 🟡 → ✅ when complete
   - Add notes if blocked (❌)

4. **Validate against acceptance criteria** (Spec §7):
   - Before marking complete, verify all relevant criteria are met

### After Completing Work

1. **Update the spec** if the implementation revealed gaps:
   - Add clarifications
   - Update open questions
   - Add new acceptance criteria if needed

2. **Document new ADRs** if architectural changes were required

3. **Update tasks.md** with any new tasks discovered

4. **Commit with clear messages** referencing spec/ADR/task IDs:
   ```
   git commit -m "[EXT-1] Implement HTTP Goal extension per ADR-2
   
   - Adds http.get() and http.post() for Strava API calls
   - Handles headers and timeouts
   - See: specs/adrs/ADR-2_http_extension.md
   - Closes task EXT-1
   
   Generated by Mistral Vibe."
   ```

5. **Push and create PR** (do NOT merge directly):
   ```bash
   git push origin feature/EXT-1-http-extension
   # Then create PR on GitHub to main
   ```

---

## 🎯 **Spec-Driven Development (SDD) Methodology**

This project follows SDD as inspired by GitHub Spec Kit:

### Phases

1. **Specification** → Define *what* (COMPLETE for this project)
2. **Constitution** → Define *principles* (COMPLETE)
3. **Architecture** → Define *how* via ADRs (COMPLETE)
4. **Task Breakdown** → Define *work items* (COMPLETE)
5. **Implementation** → Build according to spec (IN PROGRESS)
6. **Validation** → Verify against acceptance criteria

### Key SDD Concepts

| Concept | Location | Agent Action |
|---------|----------|--------------|
| **Spec** | `specs/strava_dashboard_spec_*.md` | Read before any implementation |
| **Constitution** | `specs/constitution.md` | Follow principles and standards |
| **ADRs** | `specs/adrs/` | Reference when making tech decisions |
| **Tasks** | `specs/tasks.md` | Work from this list, update status |
| **Acceptance Criteria** | Spec §7 | Validate implementation against |
| **Business Rules** | Spec §6 | Enforce in code |
| **Interfaces** | Spec §5 | Implement exactly as defined |

---

## ⚠️ **Critical Constraints for This Project**

### Language & Tooling
- **Primary Language**: [Goal](https://codeberg.org/anaseto/goal) (NOT Python, NOT JavaScript)
- **Extensions**: Written in Go (for HTTP, SQLite, JSON capabilities)
- **Frontend**: Svelte (compiles to static files)
- **Database**: SQLite (via Goal extension)

### Do NOT
- ❌ Use Python, Node.js, or other languages for backend logic
- ❌ Shell out to `curl` in production (use HTTP extension per ADR-2)
- ❌ Store tokens in plaintext in code
- ❌ Ignore rate limits (Strava: 100 req/15 min)
- ❌ Modify spec without discussion

### DO
- ✅ Follow Goal's minimal, explicit style
- ✅ Build reusable Goal extensions in Go
- ✅ Reference spec sections in code
- ✅ Log errors clearly (Fail Loud principle)
- ✅ Update tasks.md as you work

---

## 🔍 **How to Find What to Work On**

1. **Check `specs/tasks.md`** for incomplete tasks (⬜)
2. **Filter by priority**: P0 > P1 > P2
3. **Respect dependencies**: Don't start task B if it depends on incomplete task A
4. **Current suggested order**:
   - Phase 0: Foundation (EXT-1, EXT-2, EXT-3, INF-1, INF-2)
   - Phase 1: Data Pipeline (DB-1, CFG-1, API-1, API-2, SYNC-1, SYNC-2)
   - Phase 2: Statistics Engine (STAT-1, STAT-2, STAT-3)
   - Phase 3: Web Backend (EXT-4, API-3, API-4, API-5)
   - Phase 4: Frontend (FE-1, FE-2, FE-3, BUILD-1)
   - Phase 5: Integration (INT-1, INT-2, ERR-1, DOC-1, TEST-1)

---

## 📝 **Agent Prompt Template**

When asking an agent to work on this project, use:

```
You are working on the Strava Training Dashboard project.

RULES:
1. READ specs/strava_dashboard_spec_*.md FIRST
2. READ specs/constitution.md SECOND
3. READ specs/adrs/ for architecture context
4. READ specs/tasks.md for current work items
5. ONLY work on tasks marked ⬜ (not started) or 🟡 (in progress)
6. ALWAYS reference spec sections, ADRs, and task IDs in your work
7. UPDATE specs/tasks.md status as you work
8. VALIDATE against acceptance criteria before marking complete

CONTEXT:
- Language: Goal (https://codeberg.org/anaseto/goal)
- Backend extensions: Go
- Frontend: Svelte
- Database: SQLite
- Methodology: Spec-Driven Development (SDD)

CURRENT TASK: [Insert task ID and description from specs/tasks.md]
```

---

## 📞 **When in Doubt**

If you're unsure about anything:

1. **Re-read the spec** – Most questions are answered there
2. **Check the constitution** – Principles may guide your decision
3. **Look for relevant ADRs** – Architecture decisions are documented
4. **Ask the user** – But only after doing 1-3

**NEVER** make assumptions that contradict the spec or constitution.

---

## 🔗 **Resources**

- [GitHub Spec Kit Documentation](https://github.github.com/spec-kit/)
- [Goal Language](https://codeberg.org/anaseto/goal)
- [Svelte Documentation](https://svelte.dev/docs)
- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [Strava API Documentation](https://developers.strava.com/docs)

---

## 📜 **Changelog for Agents**

| Date | Change | Impact |
|------|--------|--------|
| 2025-01-XX | Initial SDD setup | All agents must follow this workflow |
| 2025-01-XX | Added Goal-specific constraints | Agents must use Goal, not other languages |

---

**Last Updated**: 2025-01-XX  
**Maintained by**: Project Owner  

*This file ensures all AI agents (including future instances of Mistral Vibe) follow the same SDD workflow.*
