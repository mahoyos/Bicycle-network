import os
import asyncpg

pool: asyncpg.Pool | None = None


async def init_db():
    global pool
    database_url = os.getenv("DATABASE_URL", "postgresql://user:password@localhost:5432/bikes_db")
    pool = await asyncpg.create_pool(dsn=database_url)

    if os.getenv("RUN_MIGRATIONS", "true").lower() == "true":
        migration_path = os.path.join(os.path.dirname(__file__), "migrations", "init.sql")
        with open(migration_path, "r") as f:
            sql = f.read()
        async with pool.acquire() as conn:
            await conn.execute(sql)


async def close_db():
    global pool
    if pool:
        await pool.close()
        pool = None


async def get_pool() -> asyncpg.Pool:
    if pool is None:
        raise RuntimeError("Database pool is not initialized")
    return pool
