import os
import uuid
from sqlalchemy import Column, String, Boolean, Integer, DateTime
from sqlalchemy.sql import func

TESTING = os.environ.get("TESTING", "false").lower() == "true"

if TESTING:
    from sqlalchemy import String as UUIDType
    def uuid_default():
        return str(uuid.uuid4())
else:
    from sqlalchemy.dialects.postgresql import UUID as UUIDType
    def uuid_default():
        return uuid.uuid4()

from app.core.database import Base


class User(Base):
    __tablename__ = "users"

    id = Column(UUIDType(as_uuid=not TESTING) if not TESTING else String(36),
                primary_key=True, default=uuid_default)
    email = Column(String, unique=True, nullable=False, index=True)
    hashed_password = Column(String, nullable=False)
    is_active = Column(Boolean, default=True)
    is_locked = Column(Boolean, default=False)
    failed_attempts = Column(Integer, default=0)
    locked_until = Column(DateTime(timezone=True), nullable=True)
    reset_token = Column(String, nullable=True)
    reset_token_expires = Column(DateTime(timezone=True), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())


class RefreshToken(Base):
    __tablename__ = "refresh_tokens"

    id = Column(UUIDType(as_uuid=not TESTING) if not TESTING else String(36),
                primary_key=True, default=uuid_default)
    user_id = Column(String(36) if TESTING else UUIDType(as_uuid=True),
                     nullable=False, index=True)
    token = Column(String, unique=True, nullable=False)
    is_revoked = Column(Boolean, default=False)
    expires_at = Column(DateTime(timezone=True), nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())