---
name: ArchiPulse technical decisions
description: Stack details, conventions, local dev config, and architectural decisions
type: project
---

## Local dev

- Runs via `docker compose up --build` — postgres + app in containers
- Migrations run automatically via `entrypoint.sh` on startup
- Frontend at http://localhost:8080

## Conventions

- Conventional Commits: `feat(scope):`, `fix(scope):`, `docs(scope):` etc.
- Branch naming: `feat/`, `fix/`, `docs/`, `extractor/`, `view/`
- One PR per concern; squash merge to main
- `go fmt` + `golangci-lint` enforced in CI

## Architecture decisions

- AOEF mapped directly to PostgreSQL tables — no custom metamodel
- `element_properties`: no UNIQUE constraint — same key can exist from multiple sources (model + extractors)
- `source` field on properties = extractor name or 'model'; `collected_at` nullable (null for model source)
- GET /elements/:id returns `{ element, properties: { "model": [...], "aws-lambda": [...] } }` — grouped by source (Option C: separate lanes, no forced reconciliation)
- EAM views are SQL queries in `internal/viewer/views/`, exposed via `/api/v1/workspaces/:id/views/:name`
- Frontend is an embedded SPA (`//go:embed all:ui/dist`) — single binary, no runtime deps

## Svelte 5 reactivity gotchas

- Functions called in templates (`{fn(arg)}`) don't create reactive dependencies on variables read inside the function body — only on args visible in the template expression. Fix: pass reactive variables explicitly as args.
- `onMount` does NOT re-fire when the same component receives new props (same route pattern, different params). Use `$: if (dep) { ... }` reactive blocks for side effects that must re-run on prop changes.

## shadcn-svelte setup

- Tailwind v4, CSS-first config (no tailwind.config.js)
- `@tailwindcss/vite` plugin in vite.config.js
- `components.json` in `cmd/archipulse/ui/`
- Add components: `npx shadcn-svelte@latest add <component>` from `cmd/archipulse/ui/`
- CSS token names: `--brand` (not `--accent`), `--text-muted` (not `--muted`) to avoid conflict with shadcn reserved names
- All layout/reset CSS must be inside `@layer base` — otherwise beats Tailwind utilities
