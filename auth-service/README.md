# 🔐 Auth Service — Red Bicicletas

Authentication microservice for the **Red Bicicletas** platform. Built with **FastAPI** and **PostgreSQL**, it handles user registration, login, password recovery, Google OAuth 2.0, and JWT token management (Access + Refresh Token).

---

## 🚀 Features

- ✅ **User registration and authentication** with bcrypt
- ✅ **JWT** — Access Token (15–60 min) + Refresh Token (7 days)
- ✅ **Google OAuth 2.0** login
- ✅ **Password recovery** via email (Gmail SMTP)
- ✅ **Account lockout** after 5 consecutive failed login attempts (15 min)
- ✅ **Refresh token revocation** (secure logout)
- ✅ **Rate limiting** — max 10 requests/min per IP on auth endpoints
- ✅ **Current user endpoint** (`/auth/me`)
- ✅ **PostgreSQL** for user and token persistence
- ✅ **Dockerized** with Docker Compose
- ✅ **Automated tests** with 90% coverage

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────┐
│                   API Gateway                        │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│              Auth Microservice                       │
│  ┌────────────────────────────────────────────┐     │
│  │              API Layer                      │     │
│  │  (register, login, refresh, logout,        │     │
│  │   password-recovery, google OAuth, me)     │     │
│  └─────────────────┬───────────────────────────┘     │
│                    │                                  │
│  ┌─────────────────▼───────────────────────────┐     │
│  │           Service Layer                     │     │
│  │  (AuthService, OAuthService)                │     │
│  └─────────────────┬───────────────────────────┘     │
│                    │                                  │
│  ┌─────────────────▼───────────────────────────┐     │
│  │            Core Layer                       │     │
│  │  (config, database, security/JWT, limiter)  │     │
│  └─────────────┬───────────────────────────────┘     │
└────────────────┼──────────────────────────────────────┘
                 │
          ┌──────▼──────┐
          │  PostgreSQL │
          │   (Users,   │
          │   Tokens)   │
          └─────────────┘
```

---

## 🛠️ Tech Stack

| Technology | Purpose |
|---|---|
| FastAPI | Web framework |
| PostgreSQL | Relational database |
| SQLAlchemy | ORM |
| bcrypt | Password hashing |
| python-jose | JWT generation and validation |
| Authlib | Google OAuth 2.0 |
| fastapi-mail | Email sending (Gmail SMTP) |
| slowapi | Rate limiting per IP |
| Docker + Docker Compose | Containerization |
| pytest + pytest-cov | Testing and coverage |

---

## 📁 Project Structure

```
auth-service/
├── app/
│   ├── api/
│   │   └── auth.py             # Endpoints
│   ├── core/
│   │   ├── config.py           # Environment variables
│   │   ├── database.py         # PostgreSQL connection
│   │   ├── dependencies.py     # JWT auth dependency
│   │   ├── email.py            # Gmail SMTP module
│   │   ├── limiter.py          # Rate limiting
│   │   └── security.py         # Password hashing and JWT
│   ├── models/
│   │   └── user.py             # Database models
│   ├── schemas/
│   │   └── auth.py             # Pydantic schemas
│   ├── services/
│   │   ├── auth_service.py     # Business logic
│   │   └── oauth_service.py    # Google OAuth configuration
│   └── main.py
├── tests/
│   ├── conftest.py             # Test fixtures and database setup
│   ├── test_register.py        # Registration tests
│   ├── test_login.py           # Login and lockout tests
│   ├── test_tokens.py          # JWT token tests
│   ├── test_password.py        # Password recovery tests
│   └── test_me.py              # /auth/me endpoint tests
├── conftest.py                 # Root conftest for Python path
├── pytest.ini                  # pytest configuration
├── Dockerfile
├── docker-compose.yml
├── requirements.txt
└── .env
```

---

## 📡 Endpoints

| Method | Route | Description | Auth required |
|---|---|---|---|
| `POST` | `/auth/register` | Register a new user | No |
| `POST` | `/auth/login` | Login with email and password | No |
| `POST` | `/auth/refresh` | Get a new access token | No |
| `POST` | `/auth/logout` | Revoke refresh token | No |
| `POST` | `/auth/password-recovery` | Request password recovery via email | No |
| `POST` | `/auth/password-reset` | Reset password with recovery token | No |
| `GET` | `/auth/google/login` | Redirect to Google consent screen | No |
| `GET` | `/auth/google/callback` | Google OAuth callback | No |
| `GET` | `/auth/me` | Get authenticated user profile | ✅ Yes |
| `GET` | `/health` | Health check | No |

Interactive documentation available at `http://localhost:8000/docs`.

---

## 🔐 JWT Flow

```
Successful login
  → Server generates Access Token (30 min) + Refresh Token (7 days)
  → Client includes Access Token on every request:
    Authorization: Bearer {access_token}
  → When Access Token expires, use Refresh Token:
    POST /auth/refresh { "refresh_token": "..." }
  → On logout, Refresh Token is revoked in the database
```

### JWT Structure

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "type": "access",
  "exp": 1704123456
}
```

---

## 🌐 Google OAuth Flow

```
User visits GET /auth/google/login
  → Redirected to Google consent screen
  → User approves permissions
  → Google redirects to GET /auth/google/callback
  → System creates account if email does not exist
  → Returns Access Token + Refresh Token
```

---

## 🛡️ Security

- Passwords hashed with **bcrypt** (min 8 chars, max 72 chars).
- Tokens signed with **HS256**.
- Accounts locked for **15 minutes** after 5 consecutive failed login attempts.
- Refresh tokens stored in database with revocation support.
- Recovery token valid for **1 hour**, invalidated after use.
- Rate limiting: **10 requests/min per IP** on login and register endpoints.
- **5 requests/min per IP** on password recovery endpoint.
- All communications recommended over **HTTPS + TLS 1.2** or higher in production.

---

## ✅ Testing

Run all tests with coverage:

```bash
pytest -v --cov=app --cov-report=term-missing
```

Current coverage: **90%** (NFR-10 requires ≥ 60%)

| Test file | Coverage |
|---|---|
| test_register.py | Registration, duplicate email, weak password |
| test_login.py | Login, wrong password, lockout, counter reset |
| test_tokens.py | Refresh, logout, token revocation |
| test_password.py | Recovery email, reset, token invalidation |
| test_me.py | Authenticated profile, missing/invalid token |

---

## ⚙️ Configuration

Copy the example file and fill in the values:

```bash
cp .env.example .env
```

```dotenv
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres123
POSTGRES_DB=auth_db

DATABASE_URL=postgresql://postgres:postgres123@postgres:5432/auth_db

# Generate with: openssl rand -hex 32
SECRET_KEY=

ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=30
REFRESH_TOKEN_EXPIRE_DAYS=7
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_MINUTES=15

# Google OAuth — get from console.cloud.google.com
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URI=http://localhost:8000/auth/google/callback

# Gmail SMTP — use an App Password, not your regular Gmail password
MAIL_USERNAME=your_email@gmail.com
MAIL_PASSWORD=xxxx xxxx xxxx xxxx
MAIL_FROM=your_email@gmail.com
MAIL_SERVER=smtp.gmail.com
MAIL_PORT=587
MAIL_STARTTLS=true
MAIL_SSL_TLS=false

# Frontend URL for password reset link
FRONTEND_URL=http://localhost:3000
```

---

## 🐳 Run with Docker

```bash
# Start services
docker-compose up --build

# Stop services
docker-compose down

# Stop and remove volumes (deletes database)
docker-compose down -v
```

---

## 💻 Run without Docker

```bash
# Install dependencies
pip install -r requirements.txt

# Run the app (requires PostgreSQL running locally)
uvicorn app.main:app --reload
```