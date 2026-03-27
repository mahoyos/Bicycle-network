from uuid import UUID

import asyncpg


class BikesRepository:
    def __init__(self, pool: asyncpg.Pool):
        self.pool = pool

    async def create(self, brand: str, bike_type: str, color: str) -> dict:
        query = """
            INSERT INTO bikes (brand, type, color)
            VALUES ($1, $2, $3)
            RETURNING id, brand, type, color, is_active, created_at
        """
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(query, brand, bike_type, color)
        return dict(row)

    async def get_by_id(self, bike_id: UUID) -> dict | None:
        query = "SELECT id, brand, type, color, is_active, created_at FROM bikes WHERE id = $1 AND is_active = TRUE"
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(query, bike_id)
        return dict(row) if row else None

    async def list_bikes(
        self,
        types: list[str] | None = None,
        page: int = 1,
        limit: int = 10,
    ) -> tuple[list[dict], int]:
        offset = (page - 1) * limit

        if types:
            # Case-insensitive matching: normalize filter values to match DB values
            type_lower_map = {t.lower(): t for t in ["Cross", "Mountain Bike", "Route"]}
            normalized = [type_lower_map.get(t.lower(), t) for t in types]

            count_query = "SELECT COUNT(*) FROM bikes WHERE type = ANY($1) AND is_active = TRUE"
            data_query = """
                SELECT id, brand, type, color, is_active, created_at FROM bikes
                WHERE type = ANY($1) AND is_active = TRUE
                ORDER BY created_at DESC
                LIMIT $2 OFFSET $3
            """
            async with self.pool.acquire() as conn:
                total = await conn.fetchval(count_query, normalized)
                rows = await conn.fetch(data_query, normalized, limit, offset)
        else:
            count_query = "SELECT COUNT(*) FROM bikes WHERE is_active = TRUE"
            data_query = """
                SELECT id, brand, type, color, is_active, created_at FROM bikes
                WHERE is_active = TRUE
                ORDER BY created_at DESC
                LIMIT $1 OFFSET $2
            """
            async with self.pool.acquire() as conn:
                total = await conn.fetchval(count_query)
                rows = await conn.fetch(data_query, limit, offset)

        return [dict(row) for row in rows], total

    async def update(self, bike_id: UUID, fields: dict) -> dict | None:
        if not fields:
            return None

        set_clauses = []
        values = []
        idx = 1
        for key, value in fields.items():
            set_clauses.append(f"{key} = ${idx}")
            values.append(value)
            idx += 1

        values.append(bike_id)
        query = f"""
            UPDATE bikes SET {', '.join(set_clauses)}
            WHERE id = ${idx} AND is_active = TRUE
            RETURNING id, brand, type, color, is_active, created_at
        """
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(query, *values)
        return dict(row) if row else None

    async def delete(self, bike_id: UUID) -> bool:
        query = "UPDATE bikes SET is_active = FALSE WHERE id = $1 AND is_active = TRUE RETURNING id"
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(query, bike_id)
        return row is not None
