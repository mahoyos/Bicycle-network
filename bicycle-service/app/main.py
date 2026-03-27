import os
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.responses import JSONResponse

from app.database.connection import close_db, get_pool, init_db
from app.dependencies.auth import get_current_user, require_admin
from app.messaging.rabbitmq import check_rabbitmq, close_rabbitmq, init_rabbitmq
from app.routers.bikes import router as bikes_router


@asynccontextmanager
async def lifespan(app: FastAPI):
    await init_db()
    await init_rabbitmq()
    yield
    await close_rabbitmq()
    await close_db()


app = FastAPI(title="Bicycle Service", lifespan=lifespan)

if os.getenv("DISABLE_AUTH", "false").lower() == "true":
    async def _no_auth():
        return {"sub": "test-user", "role": "admin"}
    app.dependency_overrides[get_current_user] = _no_auth
    app.dependency_overrides[require_admin] = _no_auth

app.include_router(bikes_router)


@app.get("/health")
async def health_check():
    return {"status": "ok", "service": "bicycle-service"}


@app.get("/ready")
async def readiness_check():
    pool = await get_pool()
    async with pool.acquire() as conn:
        await conn.fetchval("SELECT 1")
    if not check_rabbitmq():
        return JSONResponse(status_code=503, content={"status": "unavailable", "reason": "rabbitmq not connected"})
    return {"status": "ready"}
