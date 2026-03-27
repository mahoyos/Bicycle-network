import os

from fastapi import Depends, HTTPException, Request
from jose import ExpiredSignatureError, JWTError, jwt


def _get_token_from_header(request: Request) -> str:
    auth_header = request.headers.get("Authorization")
    if not auth_header:
        raise HTTPException(status_code=401, detail="Missing authorization header")

    parts = auth_header.split(" ")
    if len(parts) != 2 or parts[0] != "Bearer":
        raise HTTPException(status_code=401, detail="Invalid authorization header format")

    return parts[1]


async def get_current_user(request: Request) -> dict:
    token = _get_token_from_header(request)

    public_key = os.getenv("JWT_PUBLIC_KEY", "")
    algorithm = os.getenv("JWT_ALGORITHM", "RS256")

    try:
        payload = jwt.decode(token, public_key, algorithms=[algorithm])
        return payload
    except ExpiredSignatureError:
        raise HTTPException(status_code=401, detail="Token has expired")
    except JWTError:
        raise HTTPException(status_code=403, detail="Invalid token")


async def require_admin(request: Request) -> dict:
    payload = await get_current_user(request)

    if payload.get("role") != "admin":
        raise HTTPException(status_code=403, detail="Admin role required")

    return payload
