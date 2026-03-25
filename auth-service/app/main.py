from contextlib import asynccontextmanager
from fastapi import FastAPI
from starlette.middleware.sessions import SessionMiddleware
from slowapi import _rate_limit_exceeded_handler
from slowapi.errors import RateLimitExceeded
from slowapi.middleware import SlowAPIMiddleware

from app.core.config import settings
from app.core.database import Base, engine
from app.core.limiter import limiter
from app.api.auth import router as auth_router


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Create database tables on startup — only runs when the server starts, not during tests."""
    Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(
    title="Auth Service - Red Bicicletas",
    description="Authentication microservice for the Red Bicicletas platform.",
    version="1.0.0",
    lifespan=lifespan
)

# Add rate limiting and session middleware
app.state.limiter = limiter
app.add_exception_handler(RateLimitExceeded, _rate_limit_exceeded_handler)
app.add_middleware(SlowAPIMiddleware)

# Required for OAuth2 Google flow
app.add_middleware(SessionMiddleware, secret_key=settings.SECRET_KEY)

app.include_router(auth_router)


@app.get("/health")
def health_check():
    return {"status": "ok", "service": "auth-service"}