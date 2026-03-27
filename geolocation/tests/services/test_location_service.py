import pytest
from datetime import datetime, timezone
from unittest.mock import AsyncMock
from fastapi import HTTPException

from app.services.location_service import location_service
from app.models.bike import Bike, Location

@pytest.fixture
def mock_bike_repo(mocker):
    return mocker.patch("app.services.location_service.bike_repository", spec=True)

@pytest.mark.asyncio
async def test_get_bike_location_success(mock_bike_repo):
    now = datetime.now(timezone.utc)
    mock_bike = Bike(
        _id="bike_1",
        created_at=now,
        updated_at=now,
        location=Location(latitude=10.0, longitude=20.0, timestamp=now)
    )
    mock_bike_repo.get_bike = AsyncMock(return_value=mock_bike)

    location = await location_service.get_bike_location("bike_1")
    assert location.latitude == 10.0
    assert location.longitude == 20.0

@pytest.mark.asyncio
async def test_get_bike_location_not_found(mock_bike_repo):
    mock_bike_repo.get_bike = AsyncMock(return_value=None)

    with pytest.raises(HTTPException) as exc:
        await location_service.get_bike_location("bike_1")
    
    assert exc.value.status_code == 404
    assert exc.value.detail == "Bike not found"

@pytest.mark.asyncio
async def test_get_bike_location_no_location_yet(mock_bike_repo):
    now = datetime.now(timezone.utc)
    mock_bike = Bike(
        _id="bike_1",
        created_at=now,
        updated_at=now,
        location=None
    )
    mock_bike_repo.get_bike = AsyncMock(return_value=mock_bike)

    with pytest.raises(HTTPException) as exc:
        await location_service.get_bike_location("bike_1")
    
    assert exc.value.status_code == 404
    assert exc.value.detail == "Bike location not available yet"

@pytest.mark.asyncio
async def test_get_active_bikes_locations(mock_bike_repo):
    now = datetime.now(timezone.utc)
    mock_bike = Bike(
        _id="bike_1",
        created_at=now,
        updated_at=now,
        location=Location(latitude=10.0, longitude=20.0, timestamp=now)
    )
    mock_bike_repo.get_active_bikes = AsyncMock(return_value=[mock_bike])

    bikes = await location_service.get_active_bikes_locations()
    assert len(bikes) == 1
    assert bikes[0].location.latitude == 10.0
