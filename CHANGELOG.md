# Changelog

All notable changes to ArchiPulse will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
ArchiPulse uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Planned
- Capability Gap Analysis view (coverage heatmap)
- Interface Catalogue
- User Management UI (invitations, password reset, roles)

---

## [0.6.0] — 2026-06-23

### Added — AI-native via MCP
- MCP server (`archipulse mcp`) — 25+ tools so AI agents can create, query and export complete architecture models in natural language. Compatible with Claude Desktop and any MCP-compatible client.
- Model Editor — two-tab UI to create and edit elements and relationships directly from the web.
- Structural validation — enforces ArchiMate 3.2 relationship rules at write time.

### Added — Collaboration & governance
- Workspace RBAC — owner/member/viewer roles enforced on all scoped endpoints.
- Workspace settings UI with owner-only danger zone (delete workspace).
- Audit log — per-workspace event history with async event bus.
- Snapshots — point-in-time baselines of model state.
- OIDC role mapping for multi-user auth via JWT + OIDC.

### Added — Model lifecycle
- Import preview — semantic diff (added/changed/removed) before applying changes.
- Full AOEF round-trip — properties, multilang, viewpoints, model metadata, XSD-compliant section order.
- Saved views — bookmark filtered analytical views; update saved view from current filters.

### Added — New EAM views
- Application Landscape Map — L1/L2 capability grid with coloured app chips.
- Capability Landscape with detail panel.
- Process–Application Usage matrix.
- Technology Stack matrix.

### Added — API & tooling
- CLI (`archipulse login/context/workspace/element/relationship/diagram/import/export`) with named contexts via `~/.archipulse/config.yaml`.
- Pagination on all list endpoints.
- `GET /health` with version injected via `ldflags`.

### Changed
- README and website repositioned around AI-native EA and AOEF interoperability.
- Enriched example models (archimetal-extended, archisurance-extended).
- Workspace access middleware — consistent `{wsID}` path parameter.

### Removed
- AJX (JSON) interchange format — OEF (XML) is now the single interchange format.

### Upgrade notes
- Migrate to OEF (XML) if you were using AJX (JSON).
- Apply database migrations before starting v0.6 (RBAC tables, snapshots, audit log).

---

## [0.5.0] — 2026-04-05

### Added
- Corporate Light theme — professional light-mode UI replacing the dark navy theme
- Application Dashboard view with lifecycle/criticality donut charts and summary KPIs
- Table View as a shared route powering multiple catalogue drill-downs
- Application Dependency Graph rebuilt with @xyflow/svelte (XY Flow) — pan/zoom, custom nodes, lifecycle colors, hover tooltips

### Changed
- AppNode and CapabilityNode: pastel backgrounds with dark same-hue text, reduced shadow
- Dependency Graph and Capability Tree: light canvas, controls, and minimap
- Application Landscape Map: light L1 headers, chips visually tied to legend via color-tinted backgrounds
- All catalogue badge colors use inline `style=` to avoid Tailwind v4 build-time class purging
- Sidebar active item uses white card on slate background instead of blue tint
- Design tokens: `#f8fafc` background, `#ffffff` surface, `#2563eb` brand blue, `#cbd5e1` borders
- Frontend structure: ViewRouter-rendered views moved to `components/views/`

### Removed
- Integration Map view (superseded by Application Dependency Graph)

---

## [0.4.0] — 2026-03-30

### Added
- Svelte 5 + Vite 6 frontend replacing single-file vanilla JS SPA
- Component-based architecture: Nav, Sidebar, Home, WorkspaceOverview, TableView, GraphView, CapabilityTree, ViewRouter
- Cytoscape moved from CDN to npm dependency
- Node 22 build stage in Dockerfile and CI

### Changed
- `cmd/archipulse/web/` replaced by `cmd/archipulse/ui/` (Svelte project)
- `embed.go` points to `ui/dist/` instead of `web/`
- `frontend.go` serves `/assets/*` instead of `/static/*`

---

## [0.3.0] — 2026-03-29

### Added
- Integration Map view — application integration topology with components, services and data objects; edges colored by relationship type (Serving, Access, Flow, Triggering)
- Capability Tree rebuilt with Cytoscape.js + dagre LR layout — rectangular nodes, left-to-right hierarchy, zoom/pan, hover tooltips
- Application node sub-type differentiation: Component (solid), Service (dashed), Function (muted), Interface (teal)
- Backend filters Capability Tree to only `Capability` type elements

### Removed
- App↔Business Matrix view

---

## [0.2.0] — 2026-03-28

### Added
- Embedded SPA frontend (`//go:embed`) — single binary with no runtime dependencies
- Sidebar layout with views grouped by ArchiMate layer
- Workspace overview with element counts by layer
- EAM views: Element Catalogue, Application Catalogue, Application Landscape, Technology Catalogue, Capability Tree, Application Dependency Graph
- Application Dependency Graph with Cytoscape.js
- Docker Compose setup (postgres:17-alpine + app with healthcheck)
- Multi-stage Dockerfile (golang:1.24-alpine → alpine:3.21)
- ArchiPulse branding: orange hexagon logo, Trebuchet MS wordmark

---

## [0.1.0] — 2026-03-15

### Added
- Initial project structure
- PostgreSQL schema — AOEF as tables (workspaces, elements, relationships, diagrams)
- AOEF (XML) and AJX (JSON) parser with semantic validation
- Workspace CRUD API
- Element, relationship, and diagram CRUD API with optimistic locking
- AOEF and AJX export
- EAM viewer engine (`internal/viewer`) with SQL-based analytical views
- CI pipeline (Go build, gofmt, go vet, tests)

---

[Unreleased]: https://github.com/DisruptiveWorks/archipulse/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/DisruptiveWorks/archipulse/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/DisruptiveWorks/archipulse/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/DisruptiveWorks/archipulse/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/DisruptiveWorks/archipulse/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/DisruptiveWorks/archipulse/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/DisruptiveWorks/archipulse/releases/tag/v0.1.0
