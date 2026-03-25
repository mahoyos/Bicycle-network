import os
import pytest

# Override environment variables BEFORE any app imports
os.environ["TESTING"] = "true"
os.environ["DATABASE_URL"] = "sqlite://"
os.environ["SECRET_KEY"] = "42261b2a17084be4ddcc0a7ca546606ead66c2a8073d6362ab6002126faa8791"
os.environ["GOOGLE_CLIENT_ID"] = "test-client-id"
os.environ["GOOGLE_CLIENT_SECRET"] = "test-client-secret"
os.environ["GOOGLE_REDIRECT_URI"] = "http://localhost:8000/auth/google/callback"
os.environ["MAIL_USERNAME"] = "test@example.com"
os.environ["MAIL_PASSWORD"] = "testpassword"
os.environ["MAIL_FROM"] = "test@example.com"
os.environ["POSTGRES_USER"] = "test"
os.environ["POSTGRES_PASSWORD"] = "test"
os.environ["POSTGRES_DB"] = "test"

from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool
from unittest.mock import AsyncMock, patch

from app.main import app
from app.core.database import Base, get_db

# SQLite in-memory database for testing
SQLALCHEMY_DATABASE_URL = "sqlite://"

engine = create_engine(
    SQLALCHEMY_DATABASE_URL,
    connect_args={"check_same_thread": False},
    poolclass=StaticPool,
)
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def override_get_db():
    db = TestingSessionLocal()
    try:
        yield db
    finally:
        db.close()


@pytest.fixture(autouse=True)
def setup_database():
    """Create all tables before each test and drop them after."""
    Base.metadata.create_all(bind=engine)
    yield
    Base.metadata.drop_all(bind=engine)


@pytest.fixture
def client():
    """Test client with overridden database dependency."""
    app.dependency_overrides[get_db] = override_get_db
    with TestClient(app) as c:
        yield c
    app.dependency_overrides.clear()


@pytest.fixture
def mock_email():
    """Mock email sending to avoid real SMTP calls during tests."""
    with patch(
        "app.services.auth_service.send_password_recovery_email",
        new_callable=AsyncMock
    ) as mock:
        yield mock


@pytest.fixture
def registered_user(client):
    """Creates a registered user and returns its credentials."""
    payload = {"email": "test@example.com", "password": "password123"}
    client.post("/auth/register", json=payload)
    return payload


@pytest.fixture
def auth_tokens(client, registered_user):
    """Returns access and refresh tokens for a registered user."""
    response = client.post("/auth/login", json=registered_user)
    return response.json()