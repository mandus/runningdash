# ADR-4: Frontend with Svelte

**Status**: Accepted  
**Date**: 2025-01-XX  
**Author**: User

---

## Context
Goal cannot render dynamic web UIs. We need a modern frontend framework that:
- Is lightweight
- Compiles to static files (easy to serve)
- Handles data visualization well
- Is easy to integrate with a backend API

## Decision
Use **Svelte** for the frontend with:
- **Vite** as the build tool (fast, modern)
- **svelte-chartjs** for data visualization (charts for weekly km, pace trends)
- **TypeScript** for type safety

## Alternatives Considered
1. **Vanilla JavaScript + HTML/CSS**:
   - No build step, but more boilerplate
   - Harder to maintain for complex UIs
2. **React**:
   - More popular, but heavier and more complex
3. **Alpine.js**:
   - Very lightweight, but less structured
4. **Vue**:
   - Good middle ground, but Svelte is simpler

## Consequences
- ✅ Compiles to static JS/CSS/HTML (easy to serve from Goal backend)
- ✅ Excellent for data visualization
- ✅ Minimal runtime overhead
- ✅ Good developer experience
- ⚠️ Requires npm for dependency management
- ⚠️ Separate tech stack from backend (Goal/Go)

## Implementation
Frontend structure:
```
frontend/
├── src/
│   ├── lib/
│   │   ├── Dashboard.svelte      # Main view with charts
│   │   ├── types.ts             # TypeScript interfaces
│   │   └── api.ts               # API client for /api/* endpoints
│   ├── App.svelte                # Root component
│   └── main.ts                   # Entry point
├── public/                       # Static assets
├── package.json
├── vite.config.ts
└── tsconfig.json
```

### Build Process
```bash
# Install dependencies
npm install

# Develop
npm run dev

# Build for production (outputs to /static)
npm run build -- --outDir ../static
```

### Communication with Backend
- Frontend makes fetch calls to `/api/*` endpoints
- Backend (Goal + Go) serves from `/static/` and `/api/*`
- No direct Goal ⇄ Svelte integration (clean separation)

---

## Related
- Depends on ADR-3 (Web serving) to serve the static files
- Blocks: FE-1, FE-2, FE-3, FE-4 (all frontend tasks)
