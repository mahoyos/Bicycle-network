from typing import List
from fastapi import APIRouter
from app.models.bike import Location, Bike
from app.services.location_service import location_service

router = APIRouter(tags=["Locations"])

@router.get("/locations/active", response_model=List[Bike], summary="Get locations of all active bikes")
async def get_active_bikes():
    """
    Retrieve the locations of all active bikes with recent location updates.
    """
    return await location_service.get_active_bikes_locations()

@router.get("/locations/{bike_id}", response_model=Location, summary="Get current location of a specific bike")
async def get_bike_location(bike_id: str):
    """
    Retrieve the latest known location of a specific bike using its identifier.
    """
    return await location_service.get_bike_location(bike_id)
