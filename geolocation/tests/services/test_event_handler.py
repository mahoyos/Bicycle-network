import pytest
from unittest.mock import AsyncMock

from app.services.event_handler import handle_lifecycle_event, handle_location_event

@pytest.fixture
def mock_bike_repo(mocker):
    return mocker.patch("app.services.event_handler.bike_repository", spec=True)

@pytest.mark.asyncio
async def test_handle_lifecycle_event_created(mock_bike_repo):
    mock_bike_repo.create_bike = AsyncMock()
    
    payload = {
        "bike_id": "bike_new",
        "action": "CREATED",
        "timestamp": "2026-03-24T12:00:00Z"
    }
    
    await handle_lifecycle_event(payload)
    mock_bike_repo.create_bike.assert_called_once_with("bike_new")

@pytest.mark.asyncio
async def test_handle_lifecycle_event_deleted(mock_bike_repo):
    mock_bike_repo.delete_bike = AsyncMock()
    
    payload = {
        "bike_id": "bike_del",
        "action": "DELETED",
        "timestamp": "2026-03-24T12:00:00Z"
    }
    
    await handle_lifecycle_event(payload)
    mock_bike_repo.delete_bike.assert_called_once_with("bike_del")

@pytest.mark.asyncio
async def test_handle_lifecycle_event_invalid_schema(mock_bike_repo):
    payload = {
        "bike_id": "bike_wrong",
        # missing action
    }
    
    # Should swallow ValidationError and not raise Exception
    await handle_lifecycle_event(payload)
    mock_bike_repo.create_bike.assert_not_called()
    mock_bike_repo.delete_bike.assert_not_called()

@pytest.mark.asyncio
async def test_handle_location_event(mock_bike_repo):
    mock_bike_repo.update_location = AsyncMock()
    
    payload = {
        "bike_id": "bike_loc",
        "latitude": 40.7128,
        "longitude": -74.0060,
        "timestamp": "2026-03-24T12:00:00Z"
    }
    
    await handle_location_event(payload)
    mock_bike_repo.update_location.assert_called_once()
    
    # Verify call args
    call_args = mock_bike_repo.update_location.call_args[1]
    assert call_args["bike_id"] == "bike_loc"
    assert call_args["latitude"] == 40.7128
    assert call_args["longitude"] == -74.0060
