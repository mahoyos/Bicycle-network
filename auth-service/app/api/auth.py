from fastapi import APIRouter, Depends, Request, HTTPException
from sqlalchemy.orm import Session

from app.core.database import get_db
from app.core.config import settings
from app.core.limiter import limiter
from app.core.dependencies import get_current_user
from app.models.user import User
from app.schemas.auth import (
    RegisterRequest, RegisterResponse,
    LoginRequest, TokenResponse,
    RefreshRequest, PasswordRecoveryRequest,
    PasswordResetRequest, MessageResponse,
    UserResponse
)
from app.services.auth_service import auth_service
from app.services.oauth_service import oauth

router = APIRouter(prefix="/auth", tags=["Authentication"])

# ── Standard Auth ─────────────────────────────────────────────────────────────

@router.post("/register", response_model=RegisterResponse, status_code=201)
@limiter.limit("10/minute")
def register(request: Request, payload: RegisterRequest, db: Session = Depends(get_db)):
    """Register a new user."""
    user = auth_service.register(db, payload.email, payload.password)
    return user

@router.post("/login", response_model=TokenResponse)
@limiter.limit("10/minute")
def login(request: Request, payload: LoginRequest, db: Session = Depends(get_db)):
    """Login with email and password. Returns access and refresh token."""
    return auth_service.login(db, payload.email, payload.password)

@router.post("/refresh", response_model=dict)
def refresh_token(payload: RefreshRequest, db: Session = Depends(get_db)):
    """Get new access token using a valid refresh token."""
    return auth_service.refresh_access_token(db, payload.refresh_token)

@router.post("/logout", response_model=MessageResponse)
def logout(payload: RefreshRequest, db: Session = Depends(get_db)):
    """Logout user by revoking the refresh token."""
    auth_service.revoke_refresh_token(db, payload.refresh_token)
    return {"message": "Logged out successfully."}

@router.post("/password-recovery", response_model=MessageResponse)
@limiter.limit("5/minute")
async def password_recovery(request: Request, payload: PasswordRecoveryRequest, db: Session = Depends(get_db)):
    """Initiate password recovery process by sending a reset link to the user's email."""
    message = await auth_service.request_password_recovery(db, payload.email)
    return {"message": message}

@router.post("/password-reset", response_model=MessageResponse)
def password_reset(payload: PasswordResetRequest, db: Session = Depends(get_db)):
    """Reset password using a valid recovery token."""
    auth_service.reset_password(db, payload.token, payload.new_password)
    return {"message": "Password has been reset successfully."}

# ── Current User Endpoint ───────────────────────────────────────────────────────
@router.get("/me", response_model=UserResponse)
def get_current_user_info(current_user: User = Depends(get_current_user)):
    """Get current authenticated user's information."""
    return current_user

# ── Google OAuth2 ─────────────────────────────────────────────────────────────

@router.get("/google/login")
async def google_login(request: Request):
    """Redirect user to Google's OAuth2 consent screen."""
    return await oauth.google.authorize_redirect(request, settings.GOOGLE_REDIRECT_URI)

@router.get("/google/callback")
async def google_callback(request: Request, db: Session = Depends(get_db)):
    """Google callback. Create or get user and return JWT tokens."""
    token = await oauth.google.authorize_access_token(request)
    user_info = token.get("userinfo")
    
    if not user_info or not user_info.get("email"):
        raise HTTPException(status_code=400, detail="Failed to retrieve user info from Google.")
    
    tokens = auth_service.login_or_register_google(db, user_info["email"])
    return tokens