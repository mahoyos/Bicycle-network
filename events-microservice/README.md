# Events Microservice

A FastAPI microservice for managing events and registrations using SQLAlchemy + PostgreSQL. It includes CRUD for `Event` and registration endpoints.

## 📁 Project Structure

- `Dockerfile` - container image build.
- `docker-compose.yml` - defines services: `postgres`, `app`.
- `init.sql` - Postgres DB initialization.
- `requirements.txt` - Python dependencies.
- `app/`:
  - `main.py` - application entrypoint and global exception handler.
  - `database.py` - SQLAlchemy engine/session & dependency.
  - `models.py` - SQLAlchemy data models (`Events`, `Registrations`).
  - `schemas.py` - Pydantic schemas.
  - `crud.py` - DB CRUD operations.
  - `routers/events.py` - API endpoints.

## 🔧 Requirements

- Python 3.10+
- PostgreSQL 15+
- Docker + docker-compose (recommended)

## 🛠 Local Setup (without Docker)

1. Create virtual environment:

```bash
python -m venv .venv
source .venv/bin/activate
```

2. Install dependencies:

```bash
pip install --upgrade pip
pip install -r requirements.txt
```

3. Create database:

```sql
CREATE DATABASE "Events";
```

4. Optionally create `.env`:

```text
DATABASE_URL=postgresql://postgres:password@localhost:5432/Events
```

5. Run:

```bash
uvicorn app.main:app --reload
```

6. Open docs:

- Swagger: `http://127.0.0.1:8000/docs`
- ReDoc: `http://127.0.0.1:8000/redoc`

## 🐳 Docker Setup

1. Set the DATABASE_URL environment variable or create a `.env` file in the project root:

```text
DATABASE_URL=postgresql://postgres:password@postgres:5432/Events
```

2. Run:

```bash
docker compose up --build
```

Then access `http://127.0.0.1:8000` and API docs.

## 🧩 Database Initialization

`init.sql` currently contains:

```sql
CREATE DATABASE "Events";
CREATE DATABASE "Registrations";
```

The app uses `Events`; tables are created automatically at startup.

## 🚀 API Endpoints

### Events

- `GET /` - welcome message
- `GET /events/` - get events
  - Optional query params: `skip`, `limit`, `name`, `type`, `date`, `description`
- `POST /events/` - create event
- `GET /events/{event_id}` - read event by ID
- `PUT /events/{event_id}` - update event
- `DELETE /events/{event_id}` - delete event

### Registrations

- `POST /events/{event_id}/registrations` - register user
- `DELETE /events/{event_id}/registrations/{user_id}` - unregister user
- `GET /users/{user_id}/registrations` - list user registrations

## 🧪 Tests

Tests are in `tests/` and run with:

```bash
pytest -q
```

Test coverage includes:
- create/read/update/delete of events
- registration lifecycle
- existing/non-existent event registration error handling

## 📝 Suggested improvements

- Add Alembic migrations for schema management.
- Add authentication/authorization (JWT/OAuth2).
- Add logging/monitoring (Sentry, Prometheus).
- Add CI pipeline with lint and tests.
