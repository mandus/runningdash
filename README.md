# Strava Training Dashboard

A local system that syncs Strava training activities, calculates running statistics (weekly km, pace), and displays them in a web dashboard.

## 📋 Project Status

- **Status**: Specification Complete (v0.2)
- **Primary Language**: [Goal](https://codeberg.org/anaseto/goal) (with Go extensions)
- **Frontend**: [Svelte](https://svelte.dev/)
- **Database**: SQLite

---

## 🗂️ Repository Structure

```
running/
├── specs/                          # Spec-Driven Development documents
│   ├── strava_dashboard_spec_v0.2.md  # Main project specification
│   ├── constitution.md               # Project principles & governance
│   ├── tasks.md                      # Implementation task breakdown
│   └── adrs/                         # Architecture Decision Records
│       ├── ADR-1_goal_language.md
│       ├── ADR-2_http_extension.md
│       ├── ADR-3_web_serving.md
│       ├── ADR-4_svelte_frontend.md
│       └── ADR-5_sqlite_extension.md
├── goal/                           # Goal source files (TODO)
├── extensions/                     # Goal extensions in Go (TODO)
├── frontend/                       # Svelte app (TODO)
├── static/                         # Svelte build output (TODO)
├── .gitignore
├── README.md
└── .env.example                    # Environment template (TODO)
```

---

## 🚀 Getting Started

### Prerequisites

1. **Goal Language**: Install from [codeberg.org/anaseto/goal](https://codeberg.org/anaseto/goal)
2. **Go**: Required for building Goal extensions (v1.20+ recommended)
3. **Node.js**: Required for Svelte frontend (v18+ recommended)

### Quick Start (Once Implemented)

```bash
# Clone the repo
git clone <repo-url>
cd running

# Set up environment
cp .env.example .env
# Edit .env with your Strava tokens

# Install Svelte dependencies
cd frontend
npm install
cd ..

# Build Svelte for production
cd frontend
npm run build -- --outDir ../static
cd ..

# Run the application
goal run main.goal

# Access dashboard at http://localhost:8080
```

---

## 📖 Documentation

| Document | Purpose |
|----------|---------|
| [specs/strava_dashboard_spec_v0.2.md](specs/strava_dashboard_spec_v0.2.md) | Full project specification |
| [specs/constitution.md](specs/constitution.md) | Principles, governance, standards |
| [specs/tasks.md](specs/tasks.md) | Implementation task breakdown |
| [specs/adrs/](specs/adrs/) | Architecture decisions |

---

## 🎯 Implementation Phases

See [specs/tasks.md](specs/tasks.md) for detailed task breakdown.

### Phase 0: Foundation (P0)
- Goal extensions for HTTP, SQLite, JSON
- Development environment setup

### Phase 1: Data Pipeline
- Strava API client
- Database schema and sync logic

### Phase 2: Statistics Engine
- Pace and distance calculations
- Weekly aggregation

### Phase 3: Web Backend
- HTTP server extension
- API endpoints

### Phase 4: Frontend
- Svelte dashboard with charts

### Phase 5: Integration & Polish
- End-to-end testing
- Documentation
- Error handling

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Strava Training Dashboard               │
├─────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐   │
│  │   Strava    │    │    SQLite    │    │   Goal Core      │   │
│  │    API      │◄───┤  Database    │◄───┤  (sync, stats)  │   │
│  └─────────────┘    └─────────────┘    └─────────────────┘   │
│          ▲                  ▲                     ▲          │
│          │                  │                     │          │
│  ┌───────┴───────┐  ┌───────┴───────┐  ┌───────┴───────┐   │
│  │ HTTP Ext     │  │ SQLite Ext   │  │ JSON Ext      │   │
│  │ (Go + Goal)  │  │ (Go + Goal)  │  │ (Go + Goal)   │   │
│  └───────┬───────┘  └───────┬───────┘  └─────────────────┘   │
│          │                  │                                 │
│          └──────────┬───────┘                                 │
│                     │                                         │
│              ┌──────▼───────┐                                  │
│              │  HTTPS Ext    │                                  │
│              │ (Go + Goal)  │────┐                               │
│              └──────┬───────┘    │                               │
│                     │            │                               │
│           ┌─────────┴────────┐  │                               │
│           │   Goal Backend    │  │                               │
│           │   (main.goal)     │──┘                               │
│           └─────────┬────────┘                                  │
│                     │                                            │
│              ┌──────▼───────┐                                  │
│              │   Svelte      │                                  │
│              │  Frontend    │                                  │
│              │ (static files)│                                  │
│              └──────────────┘                                  │
│                                                                  │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 Spec-Driven Development

This project follows **Spec-Driven Development (SDD)** methodology:

1. **Specifications** are the source of truth
2. All features start with a spec update
3. Implementation must match acceptance criteria
4. Architecture decisions are documented as ADRs

### Workflow

```
Spec Update → Task Breakdown → Implementation → Validation → Merge
```

---

## 📝 Contributing

1. **Before coding**: Update the spec and create/approve ADRs for architectural changes
2. **Task tracking**: Update [specs/tasks.md](specs/tasks.md) with progress
3. **Code reviews**: Reference spec sections and ADRs in PR descriptions

---

## 🤝 Stakeholders

- **End User**: Runner who wants to track training stats
- **Developer**: You (implementing the system)
- **Strava**: API provider

---

## 📄 License

Not specified yet. Add LICENSE file when ready.
