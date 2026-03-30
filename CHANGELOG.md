# Changelog

All notable changes to ArchiPulse will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
ArchiPulse uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Planned
- Capability Gap Analysis view
- Technology Stack view

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

[Unreleased]: https://github.com/DisruptiveWorks/archipulse/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/DisruptiveWorks/archipulse/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/DisruptiveWorks/archipulse/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/DisruptiveWorks/archipulse/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/DisruptiveWorks/archipulse/releases/tag/v0.1.0
