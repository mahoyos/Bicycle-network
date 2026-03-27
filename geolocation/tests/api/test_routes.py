import pytest
from datetime import datetime, timezone
from fastapi.testclient import TestClient
from unittest.mock import AsyncMock

from app.main import app
from app.models.bike import Bike, Location

client = TestClient(app)

@pytest.fixture
def mock_location_service(mocker):
    # Mock the location_service inside the routes module
    return mocker.patch("app.api.routes.location_service", spec=True)

@pytest.mark.asyncio
async def test_get_active_bikes_locations(mock_location_service):
    # Arrange
    now = datetime.now(timezone.utc)
    mock_bike = Bike(
        _id="bike_1",
        created_at=now,
        updated_at=now,
        location=Location(latitude=40.7128, longitude=-74.0060, timestamp=now)
    )
    mock_location_service.get_active_bikes_locations = AsyncMock(return_value=[mock_bike])

    # Act
    # TestClient in fastapi is synchronous
    response = client.get("/api/v1/locations/active")

    # Assert
    assert response.status_code == 200
    data = response.json()
    assert len(data) == 1
    assert data[0]["_id"] == "bike_1"
    assert data[0]["location"]["latitude"] == 40.7128

@pytest.mark.asyncio
async def test_get_bike_location_success(mock_location_service):
    # Arrange
    now = datetime.now(timezone.utc)
    mock_location = Location(latitude=40.7128, longitude=-74.0060, timestamp=now)
    mock_location_service.get_bike_location = AsyncMock(return_value=mock_location)

    # Act
    response = client.get("/api/v1/locations/bike_1")

    # Assert
    assert response.status_code == 200
    data = response.json()
    assert data["latitude"] == 40.7128
    assert data["longitude"] == -74.0060

@pytest.mark.asyncio
async def test_get_bike_location_not_found(mock_location_service):
    # Arrange
    from fastapi import HTTPException
    mock_location_service.get_bike_location = AsyncMock(side_effect=HTTPException(status_code=404, detail="Bike not found"))

    # Act
    response = client.get("/api/v1/locations/non_existent")

    # Assert
    assert response.status_code == 404
    assert response.json()["detail"] == "Bike not found"
