# TelemetryGo

A full-stack telemetry dashboard platform for ingesting, storing, querying, and visualizing operational telemetry data from software services. Built with a focus on clean architecture, scalability, and real-time capabilities.

**Author:** Caique Guimarães de Oliveira
**License:** MIT
**Year:** 2026

---

## Table of Contents

- [Overview](#overview)
- [System Design &amp; Architecture](#system-design--architecture)
- [Technology Stack](#technology-stack)
- [Getting Started](#getting-started)
- [API Reference](#api-reference)
- [Security](#security)
- [Testing](#testing)

---

## Overview

TelemetryGo enables development teams to collect and monitor operational telemetry data — **events** (deployments, errors, threshold alerts) and **metrics** (CPU usage, memory, latency, error rates) — from their microservices through a unified dashboard.

The system follows a **polyglot persistence** strategy, using the right database for each data type, and implements **real-time streaming** via Server-Sent Events backed by Redis Pub/Sub.

---

## System Design & Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Client (Browser)                     │
│              Next.js + React 19 + Zustand               │
└─────────────────────────┬───────────────────────────────┘
                          │ HTTP / SSE
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   Go Backend API                        │
│              Clean Architecture Layers                  │
│  ┌──────────┐ ┌───────────┐ ┌────────────┐ ┌────────┐   │
│  │ Domain   │→│Application│→│Controllers │→│ Routes │   │
│  │ Entities │ │Use Cases  │ │  (Gin)     │ │        │   │
│  └──────────┘ └───────────┘ └────────────┘ └────────┘   │
│                       │                                 │
│  ┌────────────────────┴──────────────────────────────┐  │
│  │              Infrastructure Layer                 │  │
│  │  Repositories · Auth · Messaging · Middleware     │  │
│  └──────┬─────────────┬──────────────┬───────────────┘  │
└─────────┼─────────────┼──────────────┼──────────────────┘
          │             │              │
          ▼             ▼              ▼
   ┌────────────┐ ┌──────────┐ ┌────────────┐
   │ PostgreSQL │ │Cassandra │ │   Redis    │
   │  (Users)   │ │(Metrics) │ │ (Pub/Sub)  │
   └────────────┘ └──────────┘ └────────────┘
```

### Architectural Decisions

#### Clean Architecture

The backend is organized into four distinct layers with strict dependency flow (outer layers depend on inner layers, never the reverse):

1. **Domain Layer** — Pure business entities and value objects with zero external dependencies. Repository interfaces (ports) define contracts without implementation details.
2. **Application Layer** — Use cases orchestrate domain logic. Data Transfer Objects handle input/output boundaries. No framework or infrastructure code leaks here.
3. **Presentation Layer** — HTTP controllers translate between HTTP requests and use-case invocations.
4. **Infrastructure Layer** — Concrete implementations: database connectors, JWT service, Redis publisher, middleware.

**Why this matters:** This separation enables swapping entire infrastructure components (e.g., PostgreSQL → SQLite) without modifying business logic. It also makes unit testing straightforward since use cases can be tested with in-memory fakes.

#### Polyglot Persistence

Each data type is stored in a database optimized for its access pattern:

| Database                   | Data                                 | Why                                                                                                  |
| -------------------------- | ------------------------------------ | ---------------------------------------------------------------------------------------------------- |
| **PostgreSQL**       | User accounts, credentials, API keys | Strong consistency, ACID transactions, relational integrity, unique constraints                      |
| **Apache Cassandra** | Events, Metrics (time-series)        | Write-optimized LSM-tree storage, linear horizontal scaling, ideal append-only time-series workloads |
| **Redis**            | Pub/Sub message channels             | Sub-millisecond latency for real-time event streaming to connected clients                           |

**Why not a single database?** Using PostgreSQL for everything would create write contention when high-volume telemetry data streams in. Cassandra's append-only write path handles ingestion throughput that would degrade a relational database. Redis fills the gap for real-time pub/sub that neither relational nor wide-column stores provide natively.

#### Real-Time Streaming via SSE + Redis Pub/Sub

When telemetry data is ingested, it is simultaneously:

1. Persisted to Cassandra (durable storage)
2. Published to a Redis Pub/Sub channel (fan-out)

Dashboard clients subscribe to Server-Sent Events endpoints, which read from Redis channels. This decouples ingestion from delivery — the API doesn't hold connections open while waiting for clients.

**Why SSE over WebSockets?** SSE is simpler, works over standard HTTP, auto-reconnects natively, and telemetry dashboards are primarily read-only from server to client — making unidirectional streaming sufficient.

#### Graceful Degradation Pattern

The bootstrap layer checks environment variables and falls back to in-memory implementations when external services are unavailable:

- No `DATABASE_URL` → In-memory user store (map-based)
- No `CASSANDRA_HOSTS` → In-memory event/metric store
- No `REDIS_URL` → No-op publisher (streaming silently disabled)

This allows the entire backend to run with **zero external dependencies** during development, while seamlessly upgrading to production infrastructure when configured.

#### Multi-Tenant Data Isolation

Every data query is scoped to `user_id`, injected into the request context by authentication middleware. Users cannot access each other's telemetry data — enforced at the repository query level, not just at the API layer.

---

## Technology Stack

### Backend

| Technology             | Justification                                                                                                                                                                               |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Go 1.27**      | Compiled language with excellent concurrency primitives (goroutines, channels). Ideal for high-throughput ingestion services. Statically typed, fast cold starts, single binary deployment. |
| **Gin**          | High-performance HTTP framework with middleware support, routing groups, and minimal overhead. Battle-tested in production environments.                                                    |
| **GORM**         | Provides auto-migration, connection pooling, and type-safe query building for PostgreSQL. Reduces boilerplate while maintaining SQL access when needed.                                     |
| **gocql/gocqlx** | Official Apache Cassandra driver with query builder. Native CQL support for partition-key queries.                                                                                          |
| **go-redis**     | Redis client with native Pub/Sub support. Handles connection pooling and reconnection automatically.                                                                                        |
| **golang-jwt**   | Standard library for JWT creation and verification. HS256 signing for stateless authentication.                                                                                             |
| **Argon2id**     | Memory-hard password hashing function (winner of the Password Hashing Competition). Resistant to GPU/ASIC brute-force attacks. Parameters: 64MB memory, 1 iteration, 4 parallelism.         |
| **google/uuid**  | RFC 4122 compliant UUID v4 generation for unique entity identifiers.                                                                                                                        |

### Frontend

| Technology                      | Justification                                                                                                                                            |
| ------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Next.js 16**            | React framework with App Router, server components, and built-in API proxying. The proxy layer (`rewrites`) eliminates CORS issues during development. |
| **React 19**              | Component-based UI library with concurrent features. Latest stable release with performance improvements.                                                |
| **TypeScript**            | Static type checking prevents runtime type errors. Shared type definitions between components improve maintainability.                                   |
| **Zustand**               | Minimal, unopinionated state management. No providers/boilerplate needed. Supports middleware for persistence and devtools.                              |
| **React Hook Form + Zod** | Form state management with schema-based validation. Zod schemas serve as single source of truth for both validation and TypeScript types.                |
| **Tailwind CSS v4**       | Utility-first CSS with zero runtime overhead. Rapid prototyping with consistent design tokens.                                                           |
| **shadcn/ui**             | Copy-published components (not a dependency) built on Radix UI primitives. Full control over styling and behavior.                                       |
| **Axios**                 | HTTP client with interceptor support. Enables automatic token refresh on 401 responses without manual handling in every request.                         |

### Infrastructure

| Technology                        | Justification                                                                                                                |
| --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| **Docker + Docker Compose** | Consistent development environments. Reproducible builds, dependency management, and service orchestration in a single file. |
| **PostgreSQL 16**           | Industry-standard RDBMS. ACID compliance for user data integrity. Alpine image for minimal footprint.                        |
| **Apache Cassandra 5.0**    | Distributed wide-column store. Write-optimized for time-series data. Linear scalability through partitioning.                |
| **Redis 7**                 | In-memory data store with Pub/Sub, persistence (AOF), and LRU eviction. Used as message broker for real-time streaming.      |

### Why Not [Alternative]?

| Alternative               | Why Not                                                                                                                             |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| MongoDB                   | Cassandra's partition-key model is more efficient for time-series queries with user-scoped access patterns.                         |
| RabbitMQ/Kafka            | Redis Pub/Sub is sufficient for the fan-out pattern here and adds no additional infrastructure. Would be preferred at higher scale. |
| Express/Fastify (Node.js) | Go's goroutine model provides better resource efficiency for concurrent ingestion workloads with lower memory footprint.            |
| Django/Rails              | Framework overhead and Python/Ruby performance would bottleneck the high-throughput ingestion path.                                 |

---

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Git

### Quick Start

```bash
git clone https://github.com/your-username/TelemetryGo.git
cd TelemetryGo
docker compose up --build
```

The application will be available at:

- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:8080

### Environment Variables

Copy `.env.example` to `.env` and configure:

```env
# Database
DATABASE_URL=postgres://user:password@database2:5432/telemetrygo?sslmode=disable

# Cassandra
CASSANDRA_HOSTS=database1:9042

# Redis
REDIS_URL=redis:6379

# JWT
JWT_SECRET=your-256-bit-secret

# Backend URL (for frontend proxy)
BACKEND_URL=http://localhost:8080
```

### Development without Docker

**Backend (Go only):**

```bash
cd server
go run src/main.go
# Runs with in-memory storage — no external dependencies needed
```

**Frontend (Bun only):**

```bash
cd web
bun install
bun run dev
```

---

## API Reference

### Authentication

| Endpoint                 | Method | Auth   | Description                               |
| ------------------------ | ------ | ------ | ----------------------------------------- |
| `/api/v1/users`        | POST   | None   | Register (returns JWT + API key)          |
| `/api/v1/login`        | POST   | None   | Login (returns JWT + sets refresh cookie) |
| `/api/v1/auth/refresh` | POST   | Cookie | Refresh access token                      |
| `/api/v1/auth/logout`  | POST   | None   | Clear refresh cookie                      |
| `/api/v1/me`           | GET    | Bearer | Get current user claims                   |

### Telemetry Ingestion (API Key Auth)

| Endpoint            | Method | Auth      | Description          |
| ------------------- | ------ | --------- | -------------------- |
| `/api/v1/events`  | POST   | X-API-Key | Batch ingest events  |
| `/api/v1/metrics` | POST   | X-API-Key | Batch ingest metrics |

### Dashboard Queries (JWT Auth)

| Endpoint                   | Method | Auth   | Description                 |
| -------------------------- | ------ | ------ | --------------------------- |
| `/api/v1/events`         | GET    | Bearer | List all events for user    |
| `/api/v1/events/stream`  | GET    | Bearer | SSE real-time event stream  |
| `/api/v1/metrics`        | GET    | Bearer | List all metrics for user   |
| `/api/v1/metrics/stream` | GET    | Bearer | SSE real-time metric stream |

### Example Ingestion Payload

```json
// POST /api/v1/events
[
  {
    "type": "deploy",
    "service": "payment-api",
    "message": "Deployed v2.3.1 to production",
    "severity": "info",
    "timestamp": "2026-09-10T14:30:00Z"
  }
]

// POST /api/v1/metrics
[
  {
    "name": "cpu_usage",
    "service": "payment-api",
    "value": "87.5",
    "unit": "%",
    "status": "warn",
    "timestamp": "2026-09-10T14:30:00Z"
  }
]
```

---

## Security

| Layer                        | Mechanism                                                                                       |
| ---------------------------- | ----------------------------------------------------------------------------------------------- |
| **Password Storage**   | Argon2id with 64MB memory, constant-time comparison                                             |
| **Session Management** | Short-lived JWT access tokens (10 min) + long-lived refresh tokens (7 days) in HttpOnly cookies |
| **Token Transport**    | `SameSite=Lax` cookies prevent CSRF; Bearer tokens for API access                             |
| **API Keys**           | UUID-based keys with`tg_` prefix, unique per user, returned only at creation                  |
| **Data Isolation**     | Multi-tenant queries scoped to authenticated`user_id`                                         |
| **Secrets Management** | Environment variables (`.env` files, gitignored)                                              |
| **Transport Security** | Docker network segmentation — databases on private network only                                |

---

## Testing

### Unit Tests

Pure Go standard library tests (`testing` package) covering:

- Password hashing and validation (Argon2id)
- JWT token generation and verification
- Auth middleware (Bearer token and API key validation)
- User creation and login use cases

### Integration Tests

Full HTTP-level integration tests using `httptest`:

- **Auth flow** — Register → Login → Access /me → Refresh token → Logout
- **Event flow** — Ingest → List → Tenant isolation verification
- **Metric flow** — Ingest → List → Tenant isolation verification

All tests use in-memory repositories — no external dependencies required.

```bash
cd server
go test ./...
```

---
