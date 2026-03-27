import asyncio
from datetime import UTC, datetime, timedelta
from uuid import uuid4

import pytest
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from jose import jwt

# Generate RSA key pair for tests
_private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
_public_key = _private_key.public_key()

PRIVATE_KEY_PEM = _private_key.private_bytes(
    encoding=serialization.Encoding.PEM,
    format=serialization.PrivateFormat.PKCS8,
    encryption_algorithm=serialization.NoEncryption(),
).decode()

PUBLIC_KEY_PEM = _public_key.public_bytes(
    encoding=serialization.Encoding.PEM,
    format=serialization.PublicFormat.SubjectPublicKeyInfo,
).decode()

# A different key pair for "invalid signature" tests
_other_private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
OTHER_PRIVATE_KEY_PEM = _other_private_key.private_bytes(
    encoding=serialization.Encoding.PEM,
    format=serialization.PrivateFormat.PKCS8,
    encryption_algorithm=serialization.NoEncryption(),
).decode()


def create_token(role: str = "admin", expired: bool = False, wrong_key: bool = False) -> str:
    payload = {
        "sub": str(uuid4()),
        "role": role,
        "exp": datetime.now(UTC) + (timedelta(hours=-1) if expired else timedelta(hours=1)),
    }
    key = OTHER_PRIVATE_KEY_PEM if wrong_key else PRIVATE_KEY_PEM
    return jwt.encode(payload, key, algorithm="RS256")


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
