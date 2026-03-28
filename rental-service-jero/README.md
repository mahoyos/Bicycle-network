# Rental Service

Microservice for managing bicycle rentals in the Red Bicicletas platform. Built with **Go + Gin**, it handles the full rental lifecycle: creating, finalizing, cancelling, and querying rentals. Includes **admin endpoints** for listing all rentals and cancelling any rental.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.22 |
| Framework | Gin |
| ORM | GORM + PostgreSQL 15 |
| Messaging | RabbitMQ (amqp091-go) |
| Auth | JWT HS256 (golang-jwt/v5) |
| Container | Docker multi-stage Alpine |

## Architecture

```
                           ┌─────────────────────┐
                           │   Auth Service       │
                           │   (generates JWT)    │
                           └─────────┬───────────┘
                                     │ JWT token
                                     ▼
┌──────────────────────────────────────────────────────────────┐
│                     RENTAL SERVICE                           │
│                                                              │
│  ┌──────────┐   ┌──────────────┐   ┌──────────────────┐     │
│  │ Handlers │──>│ Service      │──>│ Repositories     │     │
│  │ (HTTP)   │   │ (Logic)      │   │ (GORM/Postgres)  │     │
│  └──────────┘   └──────────────┘   └──────────────────┘     │
│                                                              │
└──────────────────────────┬───────────────────────────────────┘
                           │ consumes
                           ▼
                ┌─────────────────────┐
                │ bike_lifecycle_     │
                │ events (FANOUT)     │
                │                     │
                │ CREATED / DELETED   │
                │ from Bicycle Service│
                └─────────────────────┘
```

The service follows a **layered architecture**: Handler -> Service -> Repository -> Database, with a RabbitMQ consumer for tracking bike existence.

### Event-Sourced Bike Availability

Instead of making HTTP calls to the Bicycle Service to check if a bike exists, the Rental Service **consumes `bike_lifecycle_events`** (the same FANOUT exchange used by the Geolocation Service) and maintains a local `known_bikes` table. When a bike is created or deleted in the Bicycle Service, the Rental Service receives the event and updates its local registry.

A bike is considered **available** when:
1. It exists in the `known_bikes` table (received a `CREATED` event)
2. No active rental exists for that `bicycle_id` in the `rentals` table

The Rental Service is **self-sufficient** for availability tracking — no other service needs to know about rental status.

## API Endpoints

| Method | Path | Auth | Role | Description |
|--------|------|------|------|-------------|
| `POST` | `/rentals` | JWT | any | Create a new rental |
| `PATCH` | `/rentals/:id/finalize` | JWT | any | Finalize an active rental |
| `GET` | `/rentals/active` | JWT | any | Get the user's active rental |
| `GET` | `/rentals` | JWT | admin | List all rentals (paginated) |
| `PATCH` | `/rentals/:id/cancel` | JWT | admin | Cancel any active rental |
| `GET` | `/health` | No | — | Liveness probe |
| `GET` | `/ready` | No | — | Readiness probe (DB + RabbitMQ) |

### POST /rentals

Creates a rental associating the authenticated user with an available bicycle.

**Request:**
```json
{
  "bicycle_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response (201):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "user_id": "770e8400-e29b-41d4-a716-446655440002",
  "bicycle_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "active",
  "start_time": "2026-03-27T10:00:00Z",
  "end_time": null,
  "duration_seconds": null
}
```

**Errors:**
- `404` — Bicycle not found in local registry
- `409` — Bicycle is already rented / User already has an active rental

### PATCH /rentals/:id/finalize

Finalizes an active rental. Records the end time and calculates duration in seconds.

**Response (200):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "user_id": "770e8400-e29b-41d4-a716-446655440002",
  "bicycle_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "finalized",
  "start_time": "2026-03-27T10:00:00Z",
  "end_time": "2026-03-27T11:30:00Z",
  "duration_seconds": 5400
}
```

**Errors:**
- `403` — Rental does not belong to this user
- `404` — Rental not found
- `409` — Rental is not active (already finalized or cancelled)

### GET /rentals/active

Returns the authenticated user's current active rental, or `404` if none exists.

### GET /rentals (Admin only)

Lists all rentals with pagination. Requires `admin` role.

**Query parameters:**
- `limit` — Number of results (default: `20`, max: `100`)
- `offset` — Number of results to skip (default: `0`)

**Response (200):**
```json
[
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "770e8400-e29b-41d4-a716-446655440002",
    "bicycle_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "finalized",
    "start_time": "2026-03-27T10:00:00Z",
    "end_time": "2026-03-27T11:30:00Z",
    "duration_seconds": 5400
  }
]
```

### PATCH /rentals/:id/cancel (Admin only)

Cancels any active rental. Sets status to `cancelled` and records end time. Requires `admin` role.

**Response (200):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "status": "cancelled",
  "end_time": "2026-03-27T12:00:00Z"
}
```

**Errors:**
- `404` — Rental not found
- `409` — Rental is not active

## Authentication & Authorization

The service validates JWT tokens generated by the Auth Service:
- **Algorithm:** HS256 (shared secret via `JWT_SECRET` env var)
- **Token payload:** `{"sub": "user-uuid", "role": "user", "type": "access", "exp": timestamp}`
- The `sub` claim is extracted as the `user_id` for all operations
- The `role` claim is used for admin authorization (defaults to `"user"` if missing)
- Only `access` tokens are accepted (refresh tokens are rejected)

### Role-Based Access

The `RequireAdmin` middleware protects admin-only endpoints. It checks that `role == "admin"` in the JWT claims and returns `403 Forbidden` otherwise.

## Database Schema

```sql
-- Table: known_bikes (event-sourced from bike_lifecycle_events)
id         UUID PRIMARY KEY
created_at TIMESTAMPTZ

-- Table: rentals
id               UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id          UUID NOT NULL
bicycle_id       UUID NOT NULL
status           VARCHAR(20) CHECK (active | finalized | cancelled)
start_time       TIMESTAMPTZ NOT NULL
end_time         TIMESTAMPTZ
duration_seconds INTEGER
created_at       TIMESTAMPTZ
updated_at       TIMESTAMPTZ
```

**Race condition protection:** Partial unique indexes enforce at the database level that:
- A user can have at most one active rental (`idx_rentals_active_user`)
- A bike can be in at most one active rental (`idx_rentals_active_bike`)

## RabbitMQ Integration

| Direction | Exchange | Type | Queue | Message |
|-----------|----------|------|-------|---------|
| Consumes | `bike_lifecycle_events` | FANOUT | `rental_bike_lifecycle` | `{"bike_id": "uuid", "action": "CREATED\|DELETED"}` |

The service **only consumes** — it does not publish events. Bike availability is tracked entirely within the rental service's own database.

## Getting Started

### Prerequisites

- Docker and Docker Compose

### Run with Docker

```bash
# Start all services (rental-service + PostgreSQL + RabbitMQ)
docker-compose up --build

# Or in detached mode
docker-compose up --build -d
```

The service will be available at `http://localhost:8080`.

RabbitMQ Management UI is available at `http://localhost:15673` (guest/guest).

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://user:password@postgres:5432/rentals_db?sslmode=disable` | PostgreSQL connection string |
| `RABBITMQ_URL` | `amqp://guest:guest@rabbitmq:5672/` | RabbitMQ connection string |
| `JWT_SECRET` | — | Shared secret with Auth Service for HS256 JWT validation |
| `SERVER_PORT` | `8080` | HTTP server port |
| `RUN_MIGRATIONS` | `true` | Run SQL migrations on startup |

Copy `.env.example` to `.env` and adjust as needed.

### Local Development (without Docker)

```bash
# Install dependencies
go mod download

# Run the service (requires PostgreSQL and RabbitMQ running locally)
go run ./cmd/server

# Build binary
go build -o bin/rental-service ./cmd/server
```

## Testing

```bash
# Run all tests
make test

# Run tests with coverage report
make coverage

# Run tests with verbose output
go test ./... -v -cover
```

## Project Structure

```
rental-service-jero/
  cmd/server/main.go               # Entry point, dependency wiring, graceful shutdown
  internal/
    config/config.go                # Environment variable loading
    models/
      rental.go                     # Rental GORM model + status constants
      known_bike.go                 # KnownBike model (event-sourced)
    repository/
      rental_repository.go          # Rental CRUD interface + GORM implementation
      bike_repository.go            # KnownBike interface + GORM implementation
    service/
      rental_service.go             # Business logic (create, finalize, cancel, list)
    handler/
      rental_handler.go             # HTTP handlers for rental endpoints
      health_handler.go             # Health and readiness probes
    middleware/
      auth.go                       # JWT validation + RequireAdmin middleware
    messaging/
      rabbitmq.go                   # RabbitMQ connection manager
      consumer.go                   # bike_lifecycle_events consumer
    router/
      router.go                     # Gin route registration (user + admin groups)
  migrations/
    000001_init.sql                 # Database schema
  Dockerfile                        # Multi-stage build (golang:1.22-alpine -> alpine:3.19)
  docker-compose.yml                # App + PostgreSQL + RabbitMQ
  Makefile                          # Build, test, docker commands
```

## Docker

The Dockerfile uses a multi-stage build:
1. **Build stage:** `golang:1.22-alpine` compiles the binary with `CGO_ENABLED=0`
2. **Runtime stage:** `alpine:3.19` runs the binary (~15MB final image)

### Ports

| Service | Host Port | Container Port |
|---------|-----------|---------------|
| Rental Service | 8080 | 8080 |
| PostgreSQL | 5433 | 5432 |
| RabbitMQ AMQP | 5673 | 5672 |
| RabbitMQ Management | 15673 | 15672 |
