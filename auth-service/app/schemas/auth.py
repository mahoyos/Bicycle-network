from pydantic import BaseModel, EmailStr, field_validator
from uuid import UUID
from datetime import datetime
from typing import Optional


# ── Register Schema ─────────────────────────────────────────────────────────────

class RegisterRequest(BaseModel):
    email: EmailStr
    password: str

    @field_validator("password")
    @classmethod
    def validate_password(cls, v: str) -> str:
        if len(v) < 8:
            raise ValueError("Password must be at least 8 characters long.")
        if len(v.encode("utf-8")) > 72:
            raise ValueError("Password not must be at most 72 characters long.")
        return v

class RegisterResponse(BaseModel):
    id: UUID
    email: str
    created_at: datetime

    class Config:
        from_attributes = True


# ── Login Schema ───────────────────────────────────────────────────────────────────

class LoginRequest(BaseModel):
    email: EmailStr
    password: str


class TokenResponse(BaseModel):
    access_token: str
    refresh_token: str
    token_type: str = "bearer"


# ── Refresh Token Schema ────────────────────────────────────────────────────────────

class RefreshRequest(BaseModel):
    refresh_token: str


# ── Password Recovery Schema ────────────────────────────────────────────────────────

class PasswordRecoveryRequest(BaseModel):
    email: EmailStr


class PasswordResetRequest(BaseModel):
    token: str
    new_password: str

# ── Current User ───────────────────────────────────────────────────────────────────

class UserResponse(BaseModel):
    id: UUID
    email: str
    is_active: bool
    is_locked: bool
    created_at: datetime

    class Config:
        from_attributes = True

# ── Generic Schema ──────────────────────────────────────────────────────────────────

class MessageResponse(BaseModel):
    message: str