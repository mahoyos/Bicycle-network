import math
from uuid import UUID

import asyncpg

from app.messaging.rabbitmq import check_rabbitmq, publish_bike_created, publish_bike_deleted
from app.repositories.bikes import BikesRepository
from app.schemas.bikes import BikeCreate, BikeListResponse, BikeResponse, BikeUpdate


class BikesService:
    def __init__(self, pool: asyncpg.Pool):
        self.repository = BikesRepository(pool)

    async def create_bike(self, data: BikeCreate) -> BikeResponse:
        if not check_rabbitmq():
            raise RuntimeError("RabbitMQ is not available")
        result = await self.repository.create(
            brand=data.brand,
            bike_type=data.type,
            color=data.color,
        )
        await publish_bike_created(str(result["id"]))
        return BikeResponse(**result)

    async def get_bike(self, bike_id: UUID) -> BikeResponse | None:
        result = await self.repository.get_by_id(bike_id)
        if result is None:
            return None
        return BikeResponse(**result)

    async def list_bikes(
        self,
        types: list[str] | None = None,
        page: int = 1,
        limit: int = 10,
    ) -> BikeListResponse:
        bikes, total = await self.repository.list_bikes(types=types, page=page, limit=limit)
        total_pages = math.ceil(total / limit) if total > 0 else 0
        return BikeListResponse(
            data=[BikeResponse(**b) for b in bikes],
            total=total,
            page=page,
            total_pages=total_pages,
        )

    async def update_bike(self, bike_id: UUID, data: BikeUpdate) -> BikeResponse | None:
        existing = await self.repository.get_by_id(bike_id)
        if existing is None:
            return None

        fields = {}
        if data.brand is not None:
            fields["brand"] = data.brand
        if data.type is not None:
            fields["type"] = data.type
        if data.color is not None:
            fields["color"] = data.color

        if not fields:
            return BikeResponse(**existing)

        result = await self.repository.update(bike_id, fields)
        return BikeResponse(**result) if result else None

    async def delete_bike(self, bike_id: UUID) -> bool:
        if not check_rabbitmq():
            raise RuntimeError("RabbitMQ is not available")
        existing = await self.repository.get_by_id(bike_id)
        if existing is None:
            return False

        deleted = await self.repository.delete(bike_id)
        if deleted:
            await publish_bike_deleted(str(bike_id))
        return deleted
