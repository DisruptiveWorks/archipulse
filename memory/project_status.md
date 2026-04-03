---
name: ArchiPulse project status
description: Current version, completed milestones, and next up on the roadmap
type: project
---

As of 2026-04-02, main includes full Application Dashboard (merged via PR #10).

## Completed (merged to main)

- v0.1: AOEF/AJX parser, PostgreSQL schema, CRUD API, optimistic locking, export, CI
- v0.2: Embedded SPA, EAM views, Cytoscape dependency graph, Docker Compose
- v0.3: Integration Map view, Capability Tree (Cytoscape dagre LR), app sub-type differentiation
- v0.4: Svelte 5 + Vite 6 frontend (replaced vanilla JS SPA)
- `element_properties` table (migration 006): source-tracked key/value properties per element, AOEF parser updated, GET /elements/:id returns properties grouped by source
- Tailwind CSS v4 + shadcn-svelte: full component migration (Button, Badge, Dialog, Input, Label, Separator, Table), Node 24, postgres:17 in CI
- `examples/archisurance-extended.xml`: full ArchiSurance AOEF model (40 apps, 5L1/20L2 capabilities, technology + motivation layers, intentional coverage gaps)
- Application Dashboard: all-property donut grid + capability filter (layerchart@next ArcChart); backend at `GET /views/application-dashboard/stats?capability=<name>`

## Application Dashboard — shipped (main, 2026-04-02)
- All-properties donut grid (lifecycle_status, deployment_model, criticality, vendor, business_owner, user_count)
- Capability filter dropdown (fixes: AOEF uses xsi:type="Realization" not "RealizationRelationship")
- Left app list panel (w-48 card, scrollable, filtered by capability)
- Hover tooltip per legend row showing app list for that slice value
- App highlight: clicking an app in left panel rings its slice across all donuts
- Sidebar active state fix: uses `$: loc = $location` directly in template (not prop from parent)

## Rich Catalogue Views — shipped (main, 2026-04-02)
- Application Catalogue: sortable table with property columns (vendor, lifecycle, criticality, deployment, owner, user_count), colour badges, column visibility toggle, global search, CSV export (PR #12)
- Technology Catalogue: sortable table with category badges (Node/SystemSoftware/TechnologyService), Hosted Applications chips via Assignment relationships, search, CSV export
- Both catalogues moved to dedicated "Catalogues" sidebar section (layer: 'catalogue' in views.js)
- Backend: `GET /views/application-catalogue/entries` and `GET /views/technology-catalogue/entries`

## Application Landscape Map — shipped (main, 2026-04-02)
- Replaces plain table with L1/L2 capability grid + coloured app chips (PR #11)
- Overlay selector (lifecycle_status, criticality, deployment_model, vendor…) re-colours chips; legend bar shows distinct values
- Hover tooltip per chip shows name, type, all properties; L2 rows with no apps show gap-analysis "No applications"
- Backend: `GET /views/application-landscape/map` → `ApplicationLandscapeMap()` in `internal/viewer/views/application_landscape_map.go`
- Routing: `views.js` `map: true` flag → `ViewRouter` uses `$:` reactive redirect (not `onMount`) so it works from any prior view
- Reactivity fix: `chipColor(app, overlay)` passes overlay explicitly so Svelte re-evaluates on overlay change

## Current stack

- **Backend**: Go 1.24, chi v5, lib/pq, PostgreSQL 17
- **Frontend**: Svelte 5, Vite 6, Tailwind CSS v4 (@tailwindcss/vite), shadcn-svelte, layerchart@next, Cytoscape.js + dagre
- **UI components**: shadcn-svelte copy-paste in `cmd/archipulse/ui/src/lib/components/ui/`
- **CSS tokens**: `--brand` (orange), `--bg`, `--surface`, `--surface2`, `--text`, `--text-muted` mapped to shadcn vars
- **Infra**: Docker Compose (postgres:17-alpine + multi-stage build node:24-alpine + golang:1.24-alpine)
- **Local dev**: `docker compose up db -d` → `go run ./cmd/archipulse serve` → `cd cmd/archipulse/ui && npm run dev` (Vite proxy to :8080)

## Key CSS gotcha (fixed in #8)

In Tailwind v4, unlayered CSS beats `@layer utilities`. All layout/reset styles must be in `@layer base`, otherwise they strip Tailwind utility classes (padding, margin, etc.) from shadcn components.
