import pytest
from datetime import datetime, timezone
from unittest.mock import AsyncMock, MagicMock

from app.repositories.bike_repository import BikeRepository

@pytest.fixture
def repo(mocker):
    repository = BikeRepository()
    # Mock the database collection
    mock_db = mocker.patch("app.repositories.bike_repository.get_db")
    
    mock_collection = AsyncMock()
    mock_db.return_value = {"bikes": mock_collection}
    
    return repository, mock_collection

@pytest.mark.asyncio
async def test_get_bike_found(repo):
    repository, mock_collection = repo
    now = datetime.now(timezone.utc)
    
    mock_collection.find_one.return_value = {
        "_id": "bike_1",
        "created_at": now,
        "updated_at": now,
        "location": None
    }
    
    bike = await repository.get_bike("bike_1")
    assert bike is not None
    assert bike.id == "bike_1"
    mock_collection.find_one.assert_called_once_with({"_id": "bike_1"})

@pytest.mark.asyncio
async def test_get_bike_not_found(repo):
    repository, mock_collection = repo
    mock_collection.find_one.return_value = None
    
    bike = await repository.get_bike("bike_missing")
    assert bike is None

@pytest.mark.asyncio
async def test_create_bike(repo, mocker):
    repository, mock_collection = repo
    # Mock get_bike to return None initially so it gets created
    mocker.patch.object(repository, "get_bike", return_value=None)
    
    bike = await repository.create_bike("new_bike")
    
    assert bike.id == "new_bike"
    mock_collection.insert_one.assert_called_once()
    
@pytest.mark.asyncio
async def test_delete_bike(repo):
    repository, mock_collection = repo
    mock_result = MagicMock()
    mock_result.deleted_count = 1
    mock_collection.delete_one.return_value = mock_result
    
    result = await repository.delete_bike("bike_1")
    assert result is True
    mock_collection.delete_one.assert_called_once_with({"_id": "bike_1"})

@pytest.mark.asyncio
async def test_update_location(repo, mocker):
    repository, mock_collection = repo
    now = datetime.now(timezone.utc)
    
    mock_result = MagicMock()
    mock_result.modified_count = 1
    mock_collection.update_one.return_value = mock_result
    
    # Mock get_bike so it returns something after update
    mock_bike = MagicMock()
    mocker.patch.object(repository, "get_bike", return_value=mock_bike)
    
    bike = await repository.update_location("bike_1", 10.0, 20.0, now)
    
    assert bike is mock_bike
    mock_collection.update_one.assert_called_once()
    
    call_args = mock_collection.update_one.call_args[0]
    filter_arg, update_arg = call_args
    assert filter_arg == {"_id": "bike_1"}
    assert "$set" in update_arg
    assert "location" in update_arg["$set"]

@pytest.mark.asyncio
async def test_get_active_bikes(repo):
    repository, mock_collection = repo
    now = datetime.now(timezone.utc)
    
    mock_cursor = AsyncMock()
    mock_cursor.to_list.return_value = [
        {
            "_id": "bike_active",
            "created_at": now,
            "updated_at": now,
            "location": {"latitude": 10.0, "longitude": 20.0, "timestamp": now}
        }
    ]
    mock_collection.find = MagicMock(return_value=mock_cursor)
    
    bikes = await repository.get_active_bikes()
    
    assert len(bikes) == 1
    assert bikes[0].id == "bike_active"
    assert bikes[0].location.latitude == 10.0
    mock_collection.find.assert_called_once_with({"location": {"$ne": None}})
