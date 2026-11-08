# Go Production API (DDD Architecture)

[![Go Version](https://img.shields.io/badge/Go-1.27.3+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Supported-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![Architecture](https://img.shields.io/badge/Architecture-DDD-orange)](https://en.wikipedia.org/wiki/Domain-driven_design)
[![Monitoring](https://img.shields.io/badge/Monitoring-Prometheus_%26_Grafana-E6522C?style=flat&logo=prometheus)](https://prometheus.io/)

A production-grade, robust, and clean RESTful API built in Go using **Domain-Driven Design (DDD)** and the **Ports & Adapters (Hexagonal)** architectural pattern. Features full containerization with **Docker & Docker Compose** and real-time observability using **Prometheus & Grafana**.

---

## Table of Contents

- [Architectural Overview (DDD)](#architectural-overview-ddd)
- [Project Directory Layout](#project-directory-layout)
- [Prerequisites](#prerequisites)
- [Environment Configuration (`.env`)](#environment-configuration-env)
- [Database Migrations](#database-migrations)
- [Running the Project](#running-the-project)
  - [Method 1: Local Go Run](#method-1-local-go-run)
  - [Method 2: Docker Compose (Full Stack with Monitoring)](#method-2-docker-compose-full-stack-with-monitoring)
  - [Method 3: Standalone Docker Container](#method-3-standalone-docker-container)
- [Observability & Monitoring (Prometheus + Grafana)](#observability--monitoring-prometheus--grafana)
  - [Scraped Metrics](#scraped-metrics)
  - [Prometheus Dashboard (`:9090`)](#prometheus-dashboard-9090)
  - [Grafana Visualization (`:3000`)](#grafana-visualization-3000)
- [API Endpoints & Documentation](#api-endpoints--documentation)
- [Automated Testing](#automated-testing)
- [Makefile Cheatsheet](#makefile-cheatsheet)

---

## Architectural Overview (DDD)

This project strictly follows **Domain-Driven Design (DDD)** principles to keep business logic independent, testable, and completely decoupled from databases and HTTP frameworks.

```
┌────────────────────────────────────────────────────────────────────────┐
│  1. Presentation Layer (internal/server, internal/httpx)               │
│     - HTTP Handlers (Primary / Driving Adapters)                       │
│     - Middleware (RequestID, Prometheus Metrics)                       │
│     - JSON Serialization & Domain Error to HTTP Status Mapping         │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ Calls Use Cases
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│  2. Application Layer (internal/application/listing)                   │
│     - Application Service (Use Cases: Create, GetByID, List, Delete)   │
│     - DTOs (Data Transfer Objects: CreateListingInput, ListingOutput)  │
└───────────────────┬────────────────────────────────────▲───────────────┘
                    │ Coordinates                        │ Implements Port
                    ▼                                    │
┌────────────────────────────────────────┐   ┌───────────┴───────────────┐
│  3. Domain Layer (Pure Business Core)  │   │ 4. Infrastructure Layer   │
│     (internal/domain/listing)          │   │    (internal/infrastruc-  │
│     - Entity & Aggregate Root (Listing)│   │     ture/postgres)        │
│     - Value Objects (Price)            │   │    - PostgreSQL Adapter   │
│     - Domain Invariants & Rules        │   │    - SQL Queries          │
│     - Repository Port (Interface)      │   │    - Entity Reconstitution│
│     - Domain Errors                    │   │                           │
└────────────────────────────────────────┘   └───────────────────────────┘
```

### Key DDD Patterns Used

1. **Entity & Aggregate Root ([`Listing`](internal/domain/listing/listing.go))**:
   - Represents the core business object with an explicit identity (`ID`).
   - Private fields protect internal state from unauthorized external mutation (*prevents Anemic Domain Model*).
   - Invariants are verified at creation via `NewListing(...)` and mutation via `UpdateDetails(...)`.
   - `Reconstitute(...)` restores entities from persistence without re-triggering new-entity rules.
2. **Value Object ([`Price`](internal/domain/listing/price.go))**:
   - An immutable object defined entirely by its value with no independent identity.
   - Enforces the non-negative price rule (`amount >= 0`).
   - Implements equality by value (`Equals(other)`).
3. **Repository Port ([`Repository`](internal/domain/listing/repository.go))**:
   - Outbound interface defined inside the **Domain** layer.
   - Adheres to the **Dependency Inversion Principle (DIP)**: high-level business logic does not depend on database implementations.
4. **Secondary Adapter ([`ListingRepository`](internal/infrastructure/postgres/listing_repository.go))**:
   - Implements the Domain's Repository interface using PostgreSQL and raw SQL queries.
5. **Primary Adapter ([`ListingHandler`](internal/server/listing_handler.go))**:
   - Handles incoming HTTP requests, decodes JSON into DTOs, and translates domain errors to HTTP status codes.
6. **Composition Root ([`main.go`](cmd/api/main.go))**:
   - Single point where dependency injection wires all layers together.

---

## Project Directory Layout

```
.
├── cmd/
│   ├── api/                     # API entry point & Composition Root
│   │   └── main.go
│   └── migrate/                 # Database migration CLI tool
│       └── main.go
├── deploy/                      # Observability & deployment configs
│   ├── grafana/
│   │   └── provisioning/
│   │       └── datasources/     # Pre-configured Prometheus datasource
│   │           └── datasource.yml
│   └── prometheus/
│       └── prometheus.yml       # Prometheus scrape configuration
├── internal/
│   ├── application/             # Layer 2: Application Use Cases & DTOs
│   │   └── listing/
│   │       ├── dto.go
│   │       ├── service.go
│   │       └── service_test.go  # In-memory use case unit tests
│   ├── config/                  # Environment variable configuration loader
│   │   └── config.go
│   ├── db/                      # PostgreSQL connection pool initializer
│   │   └── db.go
│   ├── domain/                  # Layer 3: Core Domain Logic (Pure Go)
│   │   └── listing/
│   │       ├── errors.go        # Domain-specific errors
│   │       ├── listing.go       # Aggregate Root & Entity
│   │       ├── listing_test.go  # Pure domain unit tests
│   │       ├── price.go         # Value Object
│   │       └── repository.go    # Outbound Repository Port
│   ├── httpx/                   # HTTP response & error mapping helpers
│   │   └── response.go
│   ├── infrastructure/          # Layer 4: Infrastructure & DB Adapters
│   │   └── postgres/
│   │       └── listing_repository.go
│   ├── middleware/              # HTTP Middlewares (RequestId, Prometheus Metrics)
│   │   ├── metrics.go
│   │   └── requestId.go
│   └── server/                  # Layer 1: HTTP Presentation / Transport
│       ├── listing_handler.go
│       ├── listing_handler_test.go
│       └── router.go
├── migrations/                  # SQL migration files (golang-migrate)
│   ├── 000001_create_listings_table.up.sql
│   └── 000001_create_listings_table.down.sql
├── seeds/                       # Sample SQL seed data
│   └── listings.sql
├── .dockerignore
├── .env.example                 # Template for environment variables
├── docker-compose.yml           # Multi-service stack (API + Prometheus + Grafana)
├── Dockerfile                   # Multi-stage production container build (11MB)
├── go.mod
├── go.sum
└── Makefile                     # Build, run, test, and docker shortcuts
```

---

## Prerequisites

- [Go](https://go.dev/doc/install) (1.24+ or 1.27+)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/)
- [PostgreSQL](https://www.postgresql.org/) (Local PostgreSQL or cloud providers like [Neon](https://neon.tech/), AWS RDS, Supabase)
- [Make](https://www.gnu.org/software/make/) (optional, but recommended)

---

## Environment Configuration (`.env`)

Copy the example environment file to `.env`:

```bash
cp .env.example .env
```

Edit `.env` with your settings:

```env
# Application Port
PORT=8080

# Environment Mode (development | staging | production)
ENV=development

# Optional Host Port Mapping for Docker Compose (if host port 8080 is in use)
HOST_PORT=8080

# PostgreSQL Connection String
DATABASE_URL=postgresql://user:password@host:5432/dbname?sslmode=require
```

---

## Database Migrations

Apply database migrations to create the required tables:

```bash
# Run pending migrations up
make migrate-up
# Or directly with Go:
go run ./cmd/migrate up

# Roll back the last migration step
make migrate-down
# Or directly with Go:
go run ./cmd/migrate down
```

---

## Running the Project

### Method 1: Local Go Run

```bash
# 1. Run automated tests to verify everything passes
make test

# 2. Build and start the binary
make run
```
The server will start listening at `http://localhost:8080`.

---

### Method 2: Docker Compose (Full Stack with Monitoring)

Run the API alongside **Prometheus** and **Grafana** with a single command:

```bash
# Build and start all services in detached mode
docker compose up --build -d
```

Check the status of running containers:
```bash
docker compose ps
```

| Service | Host URL | Description |
|---|---|---|
| **API** | `http://localhost:8080` | Go REST API |
| **Prometheus** | `http://localhost:9090` | Metrics Scraper & Dashboard |
| **Grafana** | `http://localhost:3000` | Real-time Observability UI (Login: `admin` / `admin`) |

To view real-time logs:
```bash
docker compose logs -f api
```

To stop all services:
```bash
docker compose down
```

> **Tip:** If your local machine already uses port `8080`, change the host port easily without modifying the code:
> ```bash
> HOST_PORT=8082 docker compose up -d
> ```

---

### Method 3: Standalone Docker Container

If you just want to run the API container independently:

```bash
# Build the production Docker image (only ~11MB)
make docker-build

# Run container with environment file mounted
make docker-run
```

---

## Observability & Monitoring (Prometheus + Grafana)

The API includes built-in Prometheus instrumentation tracking throughput, latencies, error rates, and Go runtime stats.

### Scraped Metrics

The `/metrics` endpoint exposes metrics formatted for Prometheus:

- `http_requests_total{method, path, status}`: Total request count.
- `http_request_duration_seconds{method, path, status}`: Latency histogram buckets.
- Standard Go runtime metrics:
  - `go_goroutines`: Current active goroutines.
  - `go_memstats_alloc_bytes`: Current memory allocation.
  - `go_gc_duration_seconds`: Garbage collection pause times.

Test the raw metrics endpoint:
```bash
curl http://localhost:8080/metrics
```

### Prometheus Dashboard (`:9090`)

Open `http://localhost:9090` in your browser. Under **Status > Targets**, you should see `go-prod-api (1/1 up)`.

Useful **PromQL** queries to try in the Expression console:

```promql
# Request rate per second over the last 1 minute
rate(http_requests_total[1m])

# 95th percentile request latency in seconds
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# Total error responses (4xx and 5xx)
sum(rate(http_requests_total{status=~"[45].."}[5m]))

# Memory allocated by Go runtime
go_memstats_alloc_bytes
```

### Grafana Visualization (`:3000`)

1. Open `http://localhost:3000`.
2. Login with:
   - **Username**: `admin`
   - **Password**: `admin`
3. The Prometheus data source is **automatically pre-configured** via [`deploy/grafana/provisioning/datasources/datasource.yml`](deploy/grafana/provisioning/datasources/datasource.yml).
4. Click **Explore** or create a **New Dashboard** to visualize request rates and latency graphs.

---

## API Endpoints & Documentation

| Method | Endpoint | Description | Status Code |
|---|---|---|---|
| `GET` | `/healthz` | Liveness / health probe | `200 OK` |
| `GET` | `/metrics` | Prometheus metrics endpoint | `200 OK` |
| `POST` | `/listings` | Create a new listing | `201 Created` |
| `GET` | `/listings` | List recent listings (up to 100) | `200 OK` |
| `GET` | `/listings/{id}` | Get listing by ID | `200 OK` / `404 Not Found` |
| `DELETE` | `/listings/{id}` | Delete listing by ID | `204 No Content` / `404 Not Found` |

### Example API Requests

#### 1. Health Check
```bash
curl -i http://localhost:8080/healthz
```
*Response:*
```http
HTTP/1.1 200 OK
OK
```

#### 2. Create Listing (`POST /listings`)
```bash
curl -i -X POST http://localhost:8080/listings \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Modern Studio Apartment",
    "description": "Fully furnished studio near metro station",
    "price": 25000,
    "city": "Dhaka"
  }'
```
*Response:*
```json
{
  "id": "c1f7b093-41bb-457f-b883-fa537b02c89e",
  "title": "Modern Studio Apartment",
  "createdAt": "2026-11-08T15:30:00Z"
}
```

#### 3. List All Listings (`GET /listings`)
```bash
curl -i http://localhost:8080/listings
```

#### 4. Get Listing by ID (`GET /listings/{id}`)
```bash
curl -i http://localhost:8080/listings/c1f7b093-41bb-457f-b883-fa537b02c89e
```

#### 5. Delete Listing (`DELETE /listings/{id}`)
```bash
curl -i -X DELETE http://localhost:8080/listings/c1f7b093-41bb-457f-b883-fa537b02c89e
```

---

## Automated Testing

Thanks to DDD, each layer can be tested in isolation:
- **Domain Layer**: Tests business invariants and Value Object validation in < 2ms without any database or mock framework.
- **Application Layer**: Tests business use cases using an in-memory repository fake.
- **Presentation Layer**: End-to-end HTTP tests using `net/http/httptest`.

Run all tests:
```bash
make test
# Or directly:
go test -v ./...
```

---

## Makefile Cheatsheet

```bash
make build                 # Compile Go binary into bin/api
make run                   # Build and run the API locally
make test                  # Execute all unit and integration tests
make migrate-up            # Run database migrations up
make migrate-down          # Roll back latest migration
make docker-build          # Build production Docker image
make docker-run            # Run standalone container with .env
make docker-compose-up     # Launch API, Prometheus & Grafana stack
make docker-compose-down   # Tear down Docker Compose containers
make docker-compose-logs   # Stream logs from Docker Compose services
```

