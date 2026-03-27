from typing import List, Optional
from fastapi import HTTPException
from app.models.bike import Bike, Location
from app.repositories.bike_repository import bike_repository

class LocationService:
    async def get_bike_location(self, bike_id: str) -> Location:
        bike = await bike_repository.get_bike(bike_id)
        if not bike:
            raise HTTPException(status_code=404, detail="Bike not found")
        if not bike.location:
             raise HTTPException(status_code=404, detail="Bike location not available yet")
        return bike.location

    async def get_active_bikes_locations(self) -> List[Bike]:
        return await bike_repository.get_active_bikes()

location_service = LocationService()
