import logging
from datetime import datetime, timezone
from typing import List, Optional
from motor.motor_asyncio import AsyncIOMotorCollection

from app.core.database import get_db
from app.models.bike import Bike, Location

logger = logging.getLogger(__name__)

class BikeRepository:
    def _get_collection(self) -> AsyncIOMotorCollection:
        return get_db()["bikes"]

    async def get_bike(self, bike_id: str) -> Optional[Bike]:
        collection = self._get_collection()
        doc = await collection.find_one({"_id": bike_id})
        if doc:
            return Bike(**doc)
        return None

    async def create_bike(self, bike_id: str) -> Bike:
        collection = self._get_collection()
        
        # Check if already exists to be idempotent
        existing = await self.get_bike(bike_id)
        if existing:
            return existing

        now = datetime.now(timezone.utc)
        bike_doc = {
            "_id": bike_id,
            "created_at": now,
            "updated_at": now,
            "location": None
        }
        await collection.insert_one(bike_doc)
        logger.info(f"Bike {bike_id} created in repository.")
        return Bike(**bike_doc)

    async def delete_bike(self, bike_id: str) -> bool:
        collection = self._get_collection()
        result = await collection.delete_one({"_id": bike_id})
        if result.deleted_count > 0:
            logger.info(f"Bike {bike_id} deleted from repository.")
            return True
        return False

    async def update_location(self, bike_id: str, latitude: float, longitude: float, timestamp: Optional[datetime] = None) -> Optional[Bike]:
        collection = self._get_collection()
        now = datetime.now(timezone.utc)
        evt_timestamp = timestamp or now
        
        update_data = {
            "location": {
                "latitude": latitude,
                "longitude": longitude,
                "timestamp": evt_timestamp
            },
            "updated_at": now
        }
        
        result = await collection.update_one(
            {"_id": bike_id},
            {"$set": update_data}
        )
        
        if result.modified_count > 0:
            logger.info(f"Bike {bike_id} location updated.")
            return await self.get_bike(bike_id)
        
        logger.warning(f"Could not update location for bike {bike_id}. Bike may not exist.")
        return None

    async def get_active_bikes(self) -> List[Bike]:
        # FR-33: retrieving the locations of all active bikes with recent location updates.
        collection = self._get_collection()
        # Find all bikes that have a location set.
        cursor = collection.find({"location": {"$ne": None}})
        bikes = await cursor.to_list(length=1000)
        return [Bike(**doc) for doc in bikes]

bike_repository = BikeRepository()
