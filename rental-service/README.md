# 🚲 Rental Service — Bicycle Network

Microservice responsible for managing bicycle rentals within the Bicycle Network System platform. Built with Go + Gin and PostgreSQL, it handles rental creation, finalization, availability validation, and publishes return events to RabbitMQ.

## 🚀 Features

- ✅ Create bicycle rental with automatic availability validation
- ✅ Finalize active rental with duration calculation
- ✅ Query active rental with elapsed time
- ✅ JWT authentication (RS256) on all rental endpoints
- ✅ Internal bicycle availability check via `rentals` table (no cross-service dependency)
- ✅ Publishes `RETURNED` event to RabbitMQ fanout exchange `bike_lifecycle_events`
- ✅ Health and readiness probes (`/health`, `/ready`)
- ✅ PostgreSQL with partial unique indexes to prevent race conditions
- ✅ Dockerized with multi-stage build (~20MB image)
- ✅ Kubernetes manifests for AWS EKS deployment
- ✅ Automated tests with 3-level coverage (repository, service, handler)

## 🛠️ Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.22+ |
| HTTP Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL |
| Messaging | RabbitMQ (`amqp091-go`) |
| JWT | `golang-jwt/jwt` (RS256) |
| Docker | Multi-stage build (alpine) |
| Port | 8002 |

## 🏗️ Architecture

Layered architecture with strict dependency direction:

```
Handlers → Services → Repositories → Database
                  ↘ Messaging (RabbitMQ)
```

## 📁 Project Structure

```
rental-service/
├── cmd/
│   └── server/
│       └── main.go                     # Application entrypoint, Gin setup, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go                   # Environment variable loading
│   ├── database/
│   │   ├── connection.go               # PostgreSQL connection via GORM, migrations
│   │   └── migrations/
│   │       └── init.sql                # SQL schema: rentals table, indexes, constraints
│   ├── dependencies/
│   │   └── auth.go                     # JWT RS256 middleware for Gin
│   ├── handlers/
│   │   └── rentals.go                  # HTTP handlers (create, finalize, get active)
│   ├── messaging/
│   │   ├── rabbitmq.go                 # RabbitMQ connection, lifecycle, publish
│   │   └── events.go                   # RETURNED event definition and publishing
│   ├── models/
│   │   └── rental.go                   # GORM model for rentals table
│   ├── repositories/
│   │   └── rentals.go                  # Data access layer (CRUD queries)
│   ├── schemas/
│   │   └── rentals.go                  # Request/response DTOs
│   └── services/
│       └── rentals.go                  # Business logic, validation, orchestration
├── tests/
│   ├── handlers/
│   │   └── rentals_test.go             # HTTP + auth integration tests (16 tests)
│   ├── services/
│   │   └── rentals_test.go             # Business logic unit tests (11 tests)
│   ├── repositories/
│   │   └── rentals_test.go             # Data access tests with SQL mock (5 tests)
│   └── helpers/
│       └── jwt_helper.go               # RSA key generation and token helpers for tests
├── k8s/
│   ├── deployment.yaml                 # Deployment for EKS (2 replicas, probes, limits)
│   ├── service.yaml                    # ClusterIP Service
│   ├── ingress.yaml                    # ALB Ingress with HTTPS
│   └── migration-job.yaml             # One-time DB migration Job
├── Dockerfile                          # Multi-stage build (alpine, ~20MB)
├── docker-compose.yml                  # Local dev: app + PostgreSQL + RabbitMQ
├── .env.example                        # Environment variable template
├── go.mod
└── go.sum
```

## 📡 Endpoints

| Method | Path | Auth | Description | Success | Error Codes |
|---|---|---|---|---|---|
| `POST` | `/rentals` | User | Create a new rental | `201` | `400`, `401`, `409` |
| `PUT` | `/rentals/:id/finalize` | User | Finalize an active rental | `200` | `400`, `401`, `403`, `404`, `409` |
| `GET` | `/rentals/active` | User | Get current user's active rental | `200` | `401`, `404` |
| `GET` | `/health` | Public | Liveness probe | `200` | — |
| `GET` | `/ready` | Public | Readiness probe (DB + RabbitMQ) | `200` | `503` |

All error responses follow the format:

```json
{
    "detail": "Error description"
}
```

### Business Rules

- A user can only have **one active rental** at a time (FR-27). Attempting a second returns `409`.
- A bicycle can only be rented by **one user** at a time (FR-26). Availability is validated internally via the `rentals` table.
- Only the **owner** of a rental can finalize it. Other users receive `403`.
- Finalizing a rental calculates the **duration** automatically (FR-28) and publishes a `RETURNED` event to RabbitMQ (FR-30).

## 📨 Events (RabbitMQ)

Publishes to the **fanout exchange** `bike_lifecycle_events` (durable):

| Action | Trigger | Message Format |
|---|---|---|
| `RETURNED` | After a rental is finalized | `{"bike_id": "<uuid>", "action": "RETURNED"}` |

Messages are published with `delivery_mode: PERSISTENT` to survive broker restarts (NFR-13).

> **Note:** The `RETURNED` event follows the same format used by the Bicycle Service for `CREATED` and `DELETED` events. Any service bound to the `bike_lifecycle_events` exchange will receive it automatically.

## 🐳 Local Development

### Requirements

- Docker + Docker Compose

### Setup

```bash
cp .env.example .env
docker-compose up --build
```

The API will be available at `http://localhost:8002`.

RabbitMQ management UI: `http://localhost:15673` (guest / guest).

### Run without Docker

Requires Go 1.22+, a running PostgreSQL instance, and a running RabbitMQ instance.

```bash
# Install dependencies
go mod download

# Run the service
go run ./cmd/server/
```

## ⚙️ Environment Variables

| Variable | Description | Default |
|---|---|---|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://user:password@localhost:5433/rental_db` |
| `RABBITMQ_URL` | RabbitMQ connection string | `amqp://guest:guest@localhost:5673/` |
| `JWT_PUBLIC_KEY` | RSA public key in PEM format (must match auth-service) | — (required) |
| `JWT_ALGORITHM` | JWT algorithm | `RS256` |
| `APP_PORT` | Port to expose | `8002` |
| `DISABLE_AUTH` | Skip JWT validation for local development | `false` |
| `RUN_MIGRATIONS` | Execute SQL migrations on startup | `true` |

Copy the template and fill in the values:

```bash
cp .env.example .env
```

## 🔐 JWT Authentication

All `/rentals` endpoints require a valid JWT token in the `Authorization` header:

```
Authorization: Bearer <token>
```

The service validates tokens using **RS256** with the public key provided in `JWT_PUBLIC_KEY`. This matches the pattern used by the Bicycle Service (`bicycle-service/app/dependencies/auth.py`).

| HTTP Code | Condition |
|---|---|
| `401` | Missing header, invalid format, or expired token |
| `403` | Invalid signature or malformed token |

Set `DISABLE_AUTH=true` for local development without the auth-service running.

## ✅ Testing

Run all tests:

```bash
go test ./tests/... -v
```

Run with coverage:

```bash
go test ./tests/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

| Level | File | What is mocked | Tests |
|---|---|---|---|
| Repository | `tests/repositories/rentals_test.go` | Database (go-sqlmock) | 5 |
| Service | `tests/services/rentals_test.go` | Repository + Messaging | 11 |
| Handler | `tests/handlers/rentals_test.go` | Full service (in-memory repo) | 16 |

**Total: 32 tests** covering all endpoints, auth flows, business rules, and error cases.

## ☸️ Production Deployment (AWS EKS)

Kubernetes manifests are located in `k8s/`:

| File | Description |
|---|---|
| `deployment.yaml` | 2 replicas, liveness/readiness probes, resource limits (256m–512m CPU, 512Mi–1Gi RAM) |
| `service.yaml` | Internal ClusterIP service |
| `ingress.yaml` | ALB with HTTPS, routes `/rentals` to this service |
| `migration-job.yaml` | One-time DB migration Job (runs `init.sql`) |

Before applying, replace the following placeholders:

- `<AWS_ACCOUNT_ID>` — your 12-digit AWS account ID
- `<AWS_REGION>` — e.g. `us-east-1`
- `<ACM_CERTIFICATE_ARN>` — ARN of your SSL certificate in AWS Certificate Manager

## 📋 Functional Requirements Coverage

| FR | Description | Implementation |
|---|---|---|
| FR-23 | Create rental | `POST /rentals` → creates rental with status `active` |
| FR-24 | Finalize rental | `PUT /rentals/:id/finalize` → sets status `finalized`, records `end_time` |
| FR-25 | Manage rental status | Status transitions: `active` → `finalized` / `cancelled` |
| FR-26 | Validate bicycle availability | Internal check via `rentals` table (no cross-service call) |
| FR-27 | Validate active rental | Rejects if user already has an active rental (`409`) |
| FR-28 | Calculate rental duration | Computed as `end_time - start_time` on finalization |
| FR-29 | Query active rental | `GET /rentals/active` → returns rental + elapsed duration |
| FR-30 | Publish return event | Publishes `RETURNED` to `bike_lifecycle_events` exchange |
