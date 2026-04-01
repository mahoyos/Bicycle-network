import asyncio
from datetime import UTC, datetime, timedelta
from uuid import uuid4

import pytest
from jose import jwt

SECRET_KEY = "test-secret-key-for-hmac-256"
WRONG_SECRET_KEY = "wrong-secret-key-for-hmac-256"


def create_token(role: str = "admin", expired: bool = False, wrong_key: bool = False) -> str:
    payload = {
        "sub": str(uuid4()),
        "role": role,
        "exp": datetime.now(UTC) + (timedelta(hours=-1) if expired else timedelta(hours=1)),
    }
    key = WRONG_SECRET_KEY if wrong_key else SECRET_KEY
    return jwt.encode(payload, key, algorithm="HS256")


ADMIN_TOKEN = create_token(role="admin")
USER_TOKEN = create_token(role="user")
EXPIRED_TOKEN = create_token(expired=True)
INVALID_TOKEN = create_token(wrong_key=True)


@pytest.fixture
def admin_headers():
    return {"Authorization": f"Bearer {create_token(role='admin')}"}


@pytest.fixture
def user_headers():
    return {"Authorization": f"Bearer {create_token(role='user')}"}


@pytest.fixture
def expired_headers():
    return {"Authorization": f"Bearer {create_token(expired=True)}"}


@pytest.fixture
def invalid_headers():
    return {"Authorization": f"Bearer {create_token(wrong_key=True)}"}
