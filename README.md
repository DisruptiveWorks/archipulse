<div align="center">

# ArchiPulse

**Enterprise Architecture models without vendor lock-in.**
**Publish your ArchiMate models and explore, analyze and navigate them from a self-hosted web platform.**

Built on ArchiMate · Powered by Go · PostgreSQL · Open Source

[![Build](https://img.shields.io/github/actions/workflow/status/DisruptiveWorks/archipulse/ci.yml?branch=main&style=flat-square)](https://github.com/DisruptiveWorks/archipulse/actions)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue?style=flat-square)](./LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24%2B-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17%2B-336791?style=flat-square&logo=postgresql)](https://www.postgresql.org)
[![ArchiMate](https://img.shields.io/badge/ArchiMate-3.2-orange?style=flat-square)](https://www.opengroup.org/archimate-forum)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen?style=flat-square)](./CONTRIBUTING.md)

[Getting Started](#getting-started) · [Features](#features) · [How It Works](#how-it-works) · [Roadmap](#roadmap) · [Contributing](#contributing) · [Support](#support)

</div>

---

> **Note:** ArchiPulse is in early development (pre-alpha). The API and data model are not yet stable. We welcome contributors and early adopters who want to shape the direction of the project.

---

## What is ArchiPulse?

ArchiPulse is an open-source platform for **storing, visualizing, navigating, and analyzing ArchiMate-based Enterprise Architecture models** through a collaborative web platform.

Most EA tools today fall into one of two traps: too academic (OWL ontologies, Protégé, SPARQL) or too proprietary (vendor lock-in, closed formats, expensive licenses). ArchiPulse takes a different approach — it maps the **ArchiMate Open Exchange Format (AOEF) directly to PostgreSQL tables**, making the standard itself the data model.

The result: your architecture is not a static file but **living, collaborative data** — queryable, enrichable, versioned by baseline, and always exportable back to any AOEF-compliant tool.

ArchiPulse works alongside the tools architects already use — **Archi**, **archimate-editor**, or any AOEF-compatible tool. It adds the collaborative repository, the analytical layer, and the enrichment pipeline on top.

---

## Table of Contents

- [Features](#features)
- [How It Works](#how-it-works)
- [Screenshots](#screenshots)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Quick Start](#quick-start)
- [Supported Formats](#supported-formats)
- [Architecture](#architecture)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Support & Sponsorship](#support--sponsorship)
- [License](#license)

---

## Features

**Collaborative Repository**
- AOEF-as-tables: the ArchiMate Open Exchange Format mapped directly to PostgreSQL — no custom metamodel
- Multiple architects edit the same workspace simultaneously — changes visible on refresh
- Optimistic locking prevents silent overwrites — conflicts shown with author and timestamp
- Semantic diff on AOEF upload — review element-by-element what changed and who changed it
- One workspace per baseline (`Q1-2026-AS-IS`, `Q1-2026-TO-BE`, `initiative-payment-modernization`)

**Viewer & Navigation**
- Static viewer — faithful reproduction of ArchiMate views as designed
- EAM views — pre-defined analytical views (capability maps, application landscapes, technology radars) generated from SQL
- Graph explorer — Cytoscape.js with visual filters or direct SQL for ad-hoc queries

**Enrichment Pipeline**
- Connect real-world resource catalogs (AWS, Confluence, Excel, custom sources) to your ArchiMate workspace
- Two-stage ETL: extractors collect raw data, mappers translate it to ArchiMate element types
- Mapping rules execute against the same CRUD API used by the web interface — no special internal paths
- Community-contributed extractor library — one extractor works across all organizations

**Open & Integrable**
- Import and export any workspace as valid AOEF XML or AJX (ArchiMate JSON Exchange)
- Full REST API — every operation available programmatically
- Self-hosted — your data stays in your infrastructure
- Compatible with Archi, archimate-editor, BiZZdesign, Sparx EA, and any AOEF-compliant tool

---

## How It Works

```mermaid
flowchart TD
    Tools["ArchiMate editors\nArchi · archimate-editor · any AOEF tool"]
    -->|"AOEF / AJX upload"| AP["ArchiPulse\nWorkspace Manager"]
    AP --> PG[(PostgreSQL\nAOEF as tables)]

    Src["External sources\nAWS · Confluence · Excel"]
    --> Ext["Extraction Engine\nExtractor → Mapper"]
    --> AP

    PG --> SV["Static viewer\nmodel as designed"]
    PG --> EAM["EAM views\nSQL analytical queries"]
    PG --> GE["Graph explorer\nCytoscape.js"]
    PG -->|"AOEF / AJX export"| Tools
```

1. Architects model in their preferred tool and **upload AOEF or AJX** to ArchiPulse
2. ArchiPulse parses the model and stores it in **PostgreSQL** — one row per element, relationship, and diagram
3. Multiple architects can **edit the workspace directly** via the web interface or API — all changes are immediately visible
4. The **enrichment pipeline** pulls from external sources and maps resources to ArchiMate elements in the workspace
5. The **viewer** renders static diagrams, generates EAM analytical views, and provides an interactive graph explorer
6. Any workspace can be **exported back to AOEF** — importable into any compliant tool at any time

---

## Screenshots

**Application Dashboard** — lifecycle and criticality breakdown across your application portfolio.

![Application Dashboard](docs/screenshots/dashboard.png)

**Application Landscape Map** — applications mapped to business capabilities with overlay filters.

![Application Landscape Map](docs/screenshots/landscape.png)

**Application Dependency Graph** — interactive graph of application integrations with XY Flow.

![Application Dependency Graph](docs/screenshots/dependency-graph.png)

**Capability Tree** — hierarchical capability model with collapsible levels.

![Capability Tree](docs/screenshots/capability-tree.png)

---

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose — recommended
- Or: [Go](https://go.dev/dl/) 1.24+, [Node.js](https://nodejs.org/) 22+, [PostgreSQL](https://www.postgresql.org/download/) 17+

### Docker (recommended)

```bash
git clone https://github.com/DisruptiveWorks/archipulse.git
cd archipulse
docker compose up
```

The web interface will be available at `http://localhost:8080`.

### Manual Installation

```bash
# Clone the repository
git clone https://github.com/DisruptiveWorks/archipulse.git
cd archipulse

# Build the frontend
cd cmd/archipulse/ui && npm install && npm run build && cd ../../..

# Configure environment
cp .env.example .env
# Edit .env — set DATABASE_URL

# Run database migrations
go run ./cmd/archipulse migrate

# Build and run
go build -o archipulse ./cmd/archipulse
./archipulse serve
```

### Quick Start

```bash
# Create a workspace
curl -X POST http://localhost:8080/api/v1/workspaces \
  -H "Content-Type: application/json" \
  -d '{"name": "Q1-2026-AS-IS", "purpose": "as-is"}'

# Import an ArchiMate model
curl -X POST http://localhost:8080/api/v1/workspaces/{id}/import \
  -F "file=@examples/archisurance.xml"

# Open the viewer
open http://localhost:8080
```

ArchiPulse ships with the **ArchiSurance** example model from The Open Group so you can explore the viewer immediately.

---

## Supported Formats

| Format | Import | Export | Notes |
|---|---|---|---|
| ArchiMate Open Exchange Format (AOEF) | ✅ v0.1 | ✅ v0.1 | Official Open Group standard · XSD validated |
| AJX (ArchiMate JSON Exchange) | ✅ v0.1 | ✅ v0.1 | Compact JSON format · Git-friendly |
| CSV | 🔜 v0.3 | ✅ v0.1 | Catalog export for manual workflows |
| Archi native (`.archimate`) | 📋 Backlog | — | Via community contribution |

---

## Architecture

ArchiPulse is built around a single core insight: **the ArchiMate Open Exchange Format already defines what entities exist — map them directly to PostgreSQL tables.**

This means export is a SELECT, import is an INSERT, and collaboration is database-native. No custom metamodel, no graph database, no vendor lock-in.

**Repository structure:**

```
archipulse/
├── cmd/
│   └── archipulse/
│       ├── ui/           # Svelte 5 + Vite 6 frontend
│       │   └── src/      # Components, routes, lib
│       ├── embed.go      # //go:embed ui/dist
│       └── main.go
├── internal/
│   ├── parser/           # AOEF and AJX parsers
│   ├── workspace/        # Workspace manager and CRUD
│   ├── viewer/           # EAM view generation (SQL queries)
│   │   └── views/        # Individual view implementations
│   └── api/              # REST API handlers
├── migrations/           # PostgreSQL migrations
├── examples/             # Sample ArchiMate models (ArchiSurance)
└── tests/                # Integration tests
```

---

## Roadmap

### v0.1 — Foundation ✅
- [x] AOEF and AJX parser with semantic validation
- [x] PostgreSQL schema (AOEF as tables)
- [x] Workspace, element, relationship, diagram CRUD API
- [x] Optimistic locking on all editable resources
- [x] AOEF and AJX export
- [x] CI pipeline and test suite

### v0.2 — Viewer & Navigation ✅
- [x] Embedded SPA frontend (single binary, no runtime deps)
- [x] EAM views: Element Catalogue, Application Catalogue, Application Landscape, Technology Catalogue
- [x] Application Dependency Graph (Cytoscape.js)
- [x] Capability Tree view
- [x] Docker Compose setup

### v0.3 — EAM Views ✅
- [x] Integration Map view (application integration topology)
- [x] Capability Tree rebuilt with Cytoscape dagre LR + tooltips
- [x] Application node sub-type differentiation

### v0.4 — Frontend ✅
- [x] Svelte 5 + Vite 6 component-based frontend
- [x] Cytoscape as npm dependency

### v0.5 — UI & UX ✅
- [x] Corporate Light theme — professional enterprise UI
- [x] Application Dashboard with lifecycle/criticality charts
- [x] Application Dependency Graph rebuilt with XY Flow
- [x] Table View shared route for catalogue drill-downs
- [x] Component-based views architecture

### v0.6 — Analysis _(in progress)_
- [ ] Capability Gap Analysis (coverage heatmap)
- [ ] Technology Stack view (app → infrastructure mapping)
- [ ] Interface Catalogue

### v1.0 — Stable Platform
- [ ] Stable REST API
- [ ] Multi-user authentication and governance levels
- [ ] Helm chart for Kubernetes deployment
- [ ] Full documentation site at archipulse.org

> The roadmap is managed publicly via [GitHub Milestones](https://github.com/DisruptiveWorks/archipulse/milestones). Community input is welcome in [Discussions](https://github.com/DisruptiveWorks/archipulse/discussions).

---

## Contributing

ArchiPulse is in early development and contributions of all kinds are welcome.

Especially impactful at this stage:

- **AOEF/AJX parser** — the Go parser is the first critical piece
- **PostgreSQL schema** — migrations for the AOEF-as-tables schema
- **Extractors** — connectors for data sources your organization uses (AWS, Azure, Jira, Confluence, ServiceNow...)
- **EAM view queries** — SQL queries that generate meaningful analytical views
- **Web frontend** — Cytoscape.js graph explorer and static viewer

Please read [CONTRIBUTING.md](./CONTRIBUTING.md) to get started.

Good entry points: [`good first issue`](https://github.com/DisruptiveWorks/archipulse/labels/good%20first%20issue) · [`help wanted`](https://github.com/DisruptiveWorks/archipulse/labels/help%20wanted) · [`extractor`](https://github.com/DisruptiveWorks/archipulse/labels/extractor)

---

## Support & Sponsorship

ArchiPulse is developed and maintained by [Disruptive Works](https://github.com/DisruptiveWorks) and released free and open source under the Apache 2.0 license.

Ways to support the project:

- **Star the repository** — helps with visibility
- **Report issues and suggest features** — your feedback shapes the roadmap
- **Contribute code, documentation, or extractors** — see [CONTRIBUTING.md](./CONTRIBUTING.md)
- **Sponsor** — if your organization wants to support sustained development, reach out at [archipulse.org](https://archipulse.org) or open a [Discussion](https://github.com/DisruptiveWorks/archipulse/discussions)

---

## License

ArchiPulse is licensed under the [Apache License 2.0](./LICENSE).

ArchiMate® is a registered trademark of The Open Group. ArchiPulse is an independent project and is not affiliated with or endorsed by The Open Group.

---

<div align="center">
  <sub>Built with care by <a href="https://github.com/DisruptiveWorks">Disruptive Works</a> · <a href="https://archipulse.org">archipulse.org</a></sub>
</div>
