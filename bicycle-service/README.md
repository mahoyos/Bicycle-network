# Bicycle Service

Microservice responsible for managing the bicycle registry within the Bicycle Network System platform. Exposes a REST API with role-based access control (JWT RS256).

## Tech Stack

- **Python 3.11** + **FastAPI**
- **PostgreSQL** via asyncpg (raw SQL, no ORM)
- **RabbitMQ** via aio-pika (fanout exchange `bike_lifecycle_events`)
- **Docker** + **Docker Compose** (local development)
- **Kubernetes** (production on AWS EKS)

## Architecture

Layered architecture with strict dependency direction:

```
Router → Service → Repository → Database
```

| Layer | Responsibility |
|---|---|
| `app/routers/` | HTTP handling, input validation, auth via JWT |
| `app/services/` | Business logic, RabbitMQ event publishing |
| `app/repositories/` | Raw SQL queries via asyncpg |
| `app/schemas/` | Pydantic v2 request/response models |
| `app/dependencies/` | JWT auth dependencies (`get_current_user`, `require_admin`) |
| `app/messaging/` | RabbitMQ connection and event publishing |
| `app/database/` | asyncpg connection pool and migrations |

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/bikes` | Any role | List bikes (paginated, filterable by type) |
| `POST` | `/bikes` | Admin | Create a bike |
| `GET` | `/bikes/{id}` | Any role | Get a bike by ID |
| `PUT` | `/bikes/{id}` | Admin | Partially update a bike |
| `DELETE` | `/bikes/{id}` | Admin | Soft delete a bike |
| `GET` | `/health` | Public | Liveness check |
| `GET` | `/ready` | Public | Readiness check (DB + RabbitMQ) |

### Query parameters for `GET /bikes`

| Parameter | Type | Default | Description |
|---|---|---|---|
| `type` | string | — | Comma-separated filter: `Cross`, `Mountain Bike`, `Route` |
| `page` | int | 1 | Page number |
| `limit` | int | 10 | Results per page (max 100) |

## Events

Publishes to the fanout exchange `bike_lifecycle_events` on RabbitMQ:

| Action | Trigger |
|---|---|
| `CREATED` | After a bike is successfully inserted in the DB |
| `DELETED` | After a bike is successfully soft-deleted in the DB |

Message format:
```json
{"bike_id": "uuid", "action": "CREATED"}
```

## Local Development

### Requirements

- Docker + Docker Compose

### Setup

```bash
cp .env.example .env
docker-compose up --build
```

The API will be available at `http://localhost:8000`.

RabbitMQ management UI: `http://localhost:15672` (guest / guest)

### Environment Variables

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `RABBITMQ_URL` | RabbitMQ connection string |
| `JWT_PUBLIC_KEY` | RSA public key in PEM format |
| `JWT_ALGORITHM` | JWT algorithm (RS256) |
| `APP_PORT` | Port to expose (default 8000) |
| `DISABLE_AUTH` | Set to `true` to skip JWT validation (local only) |
| `RUN_MIGRATIONS` | Set to `false` when migrations run via Kubernetes Job |

## Testing

```bash
pip install -r requirements.txt
python -m pytest tests/ -v
```

Tests use `pytest-asyncio` with `httpx` and mock at the service/repository layer — no real database needed.

## Production Deployment (AWS EKS)

Kubernetes manifests are located in `k8s/`:

| File | Purpose |
|---|---|
| `deployment.yaml` | 2 replicas, liveness/readiness probes, secrets from AWS Secrets Manager |
| `service.yaml` | Internal ClusterIP service |
| `ingress.yaml` | ALB with HTTPS, routes `/bikes` to this service |
| `migration-job.yaml` | One-time DB migration Job, runs before each deploy |

Before applying, replace the following placeholders in the `k8s/` files:

- `<AWS_ACCOUNT_ID>` — your 12-digit AWS account ID
- `<AWS_REGION>` — e.g. `us-east-1`
- `<ACM_CERTIFICATE_ARN>` — ARN of your SSL certificate in AWS Certificate Manager
