from datetime import datetime, timedelta, timezone
from typing import Optional
import secrets
from sqlalchemy.orm import Session
from fastapi import HTTPException, status

from app.core.config import settings
from app.core.security import hash_password, verify_password, create_access_token, create_refresh_token, decode_token
from app.models.user import User, RefreshToken
from app.core.email import send_password_recovery_email
from app.models.user import User, RefreshToken


class AuthService:

    # ── Register ─────────────────────────────────────────────────────────────

    def register(self, db: Session, email: str, password: str) -> User:
        # Check for duplicate email
        existing = db.query(User).filter(User.email == email).first()
        if existing:
            raise HTTPException(
                status_code=status.HTTP_409_CONFLICT,
                detail="Email already registered."
            )

        user = User(
            email=email,
            hashed_password=hash_password(password)
        )
        db.add(user)
        db.commit()
        db.refresh(user)
        return user

    # ── Login ─────────────────────────────────────────────────────────────────

    def login(self, db: Session, email: str, password: str) -> dict:
        user = db.query(User).filter(User.email == email).first()

        if not user:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid credentials."
            )

        # Check if account is locked
        if user.is_locked and user.locked_until:
            now = datetime.now(timezone.utc)
            locked_until = user.locked_until.replace(tzinfo=timezone.utc) if user.locked_until.tzinfo is None else user.locked_until
            if now < locked_until:
                remaining = int((locked_until - now).total_seconds() / 60)
                raise HTTPException(
                    status_code=status.HTTP_423_LOCKED,
                    detail=f"Account locked. Try again in {remaining} minutes."
                )
            else:
                # Lockout expired, unlock account
                user.is_locked = False
                user.failed_attempts = 0
                user.locked_until = None

        # Verify password
        if not verify_password(password, user.hashed_password):
            user.failed_attempts += 1

            if user.failed_attempts >= settings.MAX_LOGIN_ATTEMPTS:
                user.is_locked = True
                user.locked_until = datetime.now(timezone.utc) + timedelta(minutes=settings.LOCKOUT_MINUTES)
                db.commit()
                raise HTTPException(
                    status_code=status.HTTP_423_LOCKED,
                    detail=f"Account locked. Try again in {settings.LOCKOUT_MINUTES} minutes."
                )

            db.commit()
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail=f"Invalid credentials. Failed attempts: {user.failed_attempts}/{settings.MAX_LOGIN_ATTEMPTS}."
            )

        # Successful login — reset failed attempts
        user.failed_attempts = 0
        user.is_locked = False
        user.locked_until = None
        db.commit()

        return self._generate_tokens(db, user)

    # ── Refresh Token ─────────────────────────────────────────────────────────

    def refresh_access_token(self, db: Session, refresh_token: str) -> dict:
        payload = decode_token(refresh_token)

        if not payload or payload.get("type") != "refresh":
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid refresh token."
            )

        # Check token in DB and not revoked
        db_token = db.query(RefreshToken).filter(
            RefreshToken.token == refresh_token,
            RefreshToken.is_revoked == False
        ).first()

        if not db_token:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Refresh token not found or revoked."
            )

        user_id = payload.get("sub")
        access_token = create_access_token(data={"sub": user_id})
        return {"access_token": access_token, "token_type": "bearer"}

    # ── Revoke Refresh Token ──────────────────────────────────────────────────

    def revoke_refresh_token(self, db: Session, refresh_token: str) -> None:
        db_token = db.query(RefreshToken).filter(
            RefreshToken.token == refresh_token
        ).first()

        if db_token:
            db_token.is_revoked = True
            db.commit()

    # ── Password Recovery ─────────────────────────────────────────────────────

    async def request_password_recovery(self, db: Session, email: str) -> str:
        user = db.query(User).filter(User.email == email).first()

        # Always return success to avoid user enumeration attacks
        if not user:
            return "If the email exists, you will receive a password recovery link."

        reset_token = secrets.token_urlsafe(32)
        user.reset_token = reset_token
        user.reset_token_expires = datetime.now(timezone.utc) + timedelta(hours=1)
        db.commit()

        await send_password_recovery_email(user.email, reset_token)

        return "If the email exists, you will receive a password recovery link."

    def reset_password(self, db: Session, token: str, new_password: str) -> None:
        user = db.query(User).filter(User.reset_token == token).first()

        if not user or not user.reset_token_expires:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Invalid recovery token."
            )

        expires = user.reset_token_expires.replace(tzinfo=timezone.utc) if user.reset_token_expires.tzinfo is None else user.reset_token_expires
        if datetime.now(timezone.utc) > expires:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="The recovery token has expired."
            )

        user.hashed_password = hash_password(new_password)
        user.reset_token = None
        user.reset_token_expires = None
        db.commit()

    # ── Google OAuth ──────────────────────────────────────────────────────────

    def login_or_register_google(self, db: Session, email: str) -> dict:
        user = db.query(User).filter(User.email == email).first()

        if not user:
            # Auto-register user
            user = User(
                email=email,
                hashed_password="",
                is_active=True
            )
            db.add(user)
            db.commit()
            db.refresh(user)

        return self._generate_tokens(db, user)

    # ── Internal helpers ──────────────────────────────────────────────────────

    def _generate_tokens(self, db: Session, user: User) -> dict:
        access_token = create_access_token(data={"sub": str(user.id)})
        refresh_token = create_refresh_token(data={"sub": str(user.id)})

        # Persist refresh token
        expires_at = datetime.now(timezone.utc) + timedelta(days=settings.REFRESH_TOKEN_EXPIRE_DAYS)
        db_token = RefreshToken(
            user_id=user.id,
            token=refresh_token,
            expires_at=expires_at
        )
        db.add(db_token)
        db.commit()

        return {
            "access_token": access_token,
            "refresh_token": refresh_token,
            "token_type": "bearer"
        }


auth_service = AuthService()