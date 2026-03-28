# Red Bicicletas - Bicycle Network

Microservices platform for managing a bicycle rental network. Built as part of the **Advanced Software Architecture** course at EAFIT University.

## Architecture

```
                          +-----------------+
                          |  Auth Service   |
                          |  (JWT issuer)   |
                          +--------+--------+
                                   | JWT tokens
                     +-------------+-------------+
                     |             |              |
                     v             v              v
             +-------+---+ +------+------+ +-----+--------+
             |  Bicycle  | |   Rental    | |    Events    |
             |  Service  | |   Service   | |    Service   |
             +-----+-----+ +------+------+ +--------------+
                   |               |
                   | publishes     | consumes
                   v               v
           +-------+---------------+-------+
           |  RabbitMQ: bike_lifecycle_events |
           |          (FANOUT)                |
           +-------+--------------------------+
                   |
                   | consumes
                   v
           +-------+--------+
           |  Geolocation   |
           |    Service     |
           +----------------+
```

## Services

| Service | Tech Stack | Port | Description |
|---------|-----------|------|-------------|
| **auth-service** | Python, FastAPI, PostgreSQL | 8000 | Authentication, JWT, Google OAuth, password recovery |
| **bicycle-service** | Python, FastAPI, PostgreSQL, RabbitMQ | 8000 | Bicycle registry CRUD, publishes lifecycle events |
| **rental-service** | Go, Gin, PostgreSQL, RabbitMQ | 8080 | Rental lifecycle (create, finalize, cancel), event-sourced bike availability |
| **events-microservice** | Python, FastAPI, PostgreSQL | 8000 | Event management and user registrations |
| **geolocation** | Python, FastAPI, MongoDB, RabbitMQ | 8000 | Real-time bike location tracking |

## Communication

- **Synchronous:** REST/HTTP between clients and services, JWT-based authentication
- **Asynchronous:** RabbitMQ FANOUT exchange (`bike_lifecycle_events`) consumed by the Rental and Geolocation services

### Event flow

1. Bicycle Service creates/deletes a bike and publishes `{"bike_id": "uuid", "action": "CREATED|DELETED"}` to the `bike_lifecycle_events` exchange
2. Rental Service consumes the event and updates its local `known_bikes` table (event-sourced availability)
3. Geolocation Service consumes the event and tracks/removes the bike for location updates

## Tech Overview

| Concern | Technology |
|---------|-----------|
| Languages | Python 3.11+, Go 1.22 |
| Web Frameworks | FastAPI, Gin |
| Databases | PostgreSQL 15/16, MongoDB |
| Messaging | RabbitMQ (FANOUT exchanges) |
| Auth | JWT HS256 (shared secret), Google OAuth 2.0 |
| Containers | Docker, Docker Compose |
| Production | Kubernetes (AWS EKS) |

## Getting Started

Each service has its own `docker-compose.yml` and can be started independently:

```bash
# Auth Service
cd auth-service && docker-compose up --build -d

# Bicycle Service
cd bicycle-service && docker-compose up --build -d

# Rental Service
cd rental-service && docker-compose up --build -d

# Events Service
cd events-microservice && docker-compose up --build -d

# Geolocation Service
cd geolocation && docker-compose up --build -d
```

Each service directory contains a `.env.example` (or equivalent) with the required environment variables.

## Port Map

| Service | App | PostgreSQL | MongoDB | RabbitMQ (AMQP) | RabbitMQ (UI) |
|---------|-----|-----------|---------|-----------------|--------------|
| auth-service | 8000 | 5432 | - | - | - |
| bicycle-service | 8000 | 5432 | - | 5672 | 15672 |
| rental-service | 8080 | 5433 | - | 5673 | 15673 |
| events-microservice | 8000 | 5432 | - | - | - |
| geolocation | 8000 | - | 27017 | 5672 | 15672 |

> The rental service uses offset ports (5433, 5673, 15673) to avoid conflicts when running alongside the bicycle service.

## Testing

```bash
# Auth Service (Python)
cd auth-service && pytest

# Bicycle Service (Python)
cd bicycle-service && pytest

# Rental Service (Go)
cd rental-service && go test ./... -cover

# Events Service (Python)
cd events-microservice && pytest
```

## Project Structure

```
Bicycle-network/
  auth-service/              # Authentication & authorization
  bicycle-service/           # Bicycle registry management
  rental-service/            # Rental lifecycle management
  events-microservice/       # Event & registration management
  geolocation/               # Real-time bike location tracking
```

Each service follows a layered architecture (Router/Handler -> Service -> Repository -> Database) and can be deployed independently.
