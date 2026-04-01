import os
from datetime import UTC, datetime
from unittest.mock import AsyncMock, patch
from uuid import uuid4

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient

from tests.conftest import SECRET_KEY, create_token

# Set env vars before importing app
os.environ["JWT_PUBLIC_KEY"] = SECRET_KEY
os.environ["JWT_ALGORITHM"] = "HS256"

from app.main import app
from app.database.connection import get_pool
from app.schemas.bikes import BikeListResponse, BikeResponse

BIKE_ID = uuid4()
NOW = datetime.now(UTC)
SAMPLE_BIKE = {
    "id": BIKE_ID,
    "brand": "Trek",
    "type": "Mountain Bike",
    "color": "Red",
    "is_active": True,
    "created_at": NOW,
}


def bike_response():
    return BikeResponse(**SAMPLE_BIKE)


# --- Mock pool dependency ---

class FakePool:
    pass


async def override_get_pool():
    return FakePool()


app.dependency_overrides[get_pool] = override_get_pool


@pytest_asyncio.fixture
async def client():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as c:
        yield c


# ===================== POST /bikes =====================

@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.create_bike", new_callable=AsyncMock)
async def test_create_bike_success(mock_create, client, admin_headers):
    mock_create.return_value = bike_response()
    body = {"brand": "Trek", "type": "Mountain Bike", "color": "Red"}
    resp = await client.post("/bikes", json=body, headers=admin_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["brand"] == "Trek"
    assert "id" in data


@pytest.mark.asyncio
async def test_create_bike_missing_field(client, admin_headers):
    body = {"type": "Mountain Bike", "color": "Red"}
    resp = await client.post("/bikes", json=body, headers=admin_headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_create_bike_invalid_type(client, admin_headers):
    body = {"brand": "Trek", "type": "BMX", "color": "Red"}
    resp = await client.post("/bikes", json=body, headers=admin_headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_create_bike_id_in_body(client, admin_headers):
    body = {"id": str(uuid4()), "brand": "Trek", "type": "Mountain Bike", "color": "Red"}
    resp = await client.post("/bikes", json=body, headers=admin_headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_create_bike_no_token(client):
    body = {"brand": "Trek", "type": "Mountain Bike", "color": "Red"}
    resp = await client.post("/bikes", json=body)
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_create_bike_expired_token(client, expired_headers):
    body = {"brand": "Trek", "type": "Mountain Bike", "color": "Red"}
    resp = await client.post("/bikes", json=body, headers=expired_headers)
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_create_bike_invalid_token(client, invalid_headers):
    body = {"brand": "Trek", "type": "Mountain Bike", "color": "Red"}
    resp = await client.post("/bikes", json=body, headers=invalid_headers)
    assert resp.status_code == 403


@pytest.mark.asyncio
async def test_create_bike_user_role_forbidden(client, user_headers):
    body = {"brand": "Trek", "type": "Mountain Bike", "color": "Red"}
    resp = await client.post("/bikes", json=body, headers=user_headers)
    assert resp.status_code == 403


# ===================== GET /bikes =====================

@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.list_bikes", new_callable=AsyncMock)
async def test_list_bikes_success_as_admin(mock_list, client, admin_headers):
    mock_list.return_value = BikeListResponse(
        data=[bike_response()], total=1, page=1, total_pages=1
    )
    resp = await client.get("/bikes", headers=admin_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert len(data["data"]) == 1
    assert data["total"] == 1


@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.list_bikes", new_callable=AsyncMock)
async def test_list_bikes_success_as_user(mock_list, client, user_headers):
    mock_list.return_value = BikeListResponse(
        data=[bike_response()], total=1, page=1, total_pages=1
    )
    resp = await client.get("/bikes", headers=user_headers)
    assert resp.status_code == 200


@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.list_bikes", new_callable=AsyncMock)
async def test_list_bikes_empty_registry(mock_list, client, admin_headers):
    mock_list.return_value = BikeListResponse(data=[], total=0, page=1, total_pages=0)
    resp = await client.get("/bikes", headers=admin_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["data"] == []
    assert data["total"] == 0


@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.list_bikes", new_callable=AsyncMock)
async def test_list_bikes_filter_single_type(mock_list, client, admin_headers):
    mock_list.return_value = BikeListResponse(
        data=[bike_response()], total=1, page=1, total_pages=1
    )
    resp = await client.get("/bikes?type=Cross", headers=admin_headers)
    assert resp.status_code == 200
    mock_list.assert_called_once()
    call_kwargs = mock_list.call_args
    assert call_kwargs.kwargs.get("types") == ["Cross"] or call_kwargs[1].get("types") == ["Cross"]


@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.list_bikes", new_callable=AsyncMock)
async def test_list_bikes_filter_multiple_types(mock_list, client, admin_headers):
    mock_list.return_value = BikeListResponse(
        data=[bike_response()], total=1, page=1, total_pages=1
    )
    resp = await client.get("/bikes?type=Cross,Route", headers=admin_headers)
    assert resp.status_code == 200


@pytest.mark.asyncio
async def test_list_bikes_invalid_type_filter(client, admin_headers):
    resp = await client.get("/bikes?type=Scooter", headers=admin_headers)
    assert resp.status_code == 400
    assert "Valid types" in resp.json()["detail"] or "valid types" in resp.json()["detail"].lower()


@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.list_bikes", new_callable=AsyncMock)
async def test_list_bikes_page_out_of_range(mock_list, client, admin_headers):
    mock_list.return_value = BikeListResponse(data=[], total=1, page=999, total_pages=1)
    resp = await client.get("/bikes?page=999", headers=admin_headers)
    assert resp.status_code == 200
    assert resp.json()["data"] == []


@pytest.mark.asyncio
async def test_list_bikes_limit_exceeds_max(client, admin_headers):
    resp = await client.get("/bikes?limit=200", headers=admin_headers)
    assert resp.status_code == 400


@pytest.mark.asyncio
async def test_list_bikes_invalid_pagination(client, admin_headers):
    resp = await client.get("/bikes?page=abc", headers=admin_headers)
    assert resp.status_code == 400


@pytest.mark.asyncio
async def test_list_bikes_no_token(client):
    resp = await client.get("/bikes")
    assert resp.status_code == 401


# ===================== GET /bikes/{id} =====================

@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.get_bike", new_callable=AsyncMock)
async def test_get_bike_success_as_admin(mock_get, client, admin_headers):
    mock_get.return_value = bike_response()
    resp = await client.get(f"/bikes/{BIKE_ID}", headers=admin_headers)
    assert resp.status_code == 200
    assert resp.json()["brand"] == "Trek"


@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.get_bike", new_callable=AsyncMock)
async def test_get_bike_success_as_user(mock_get, client, user_headers):
    mock_get.return_value = bike_response()
    resp = await client.get(f"/bikes/{BIKE_ID}", headers=user_headers)
    assert resp.status_code == 200


@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.get_bike", new_callable=AsyncMock)
async def test_get_bike_not_found(mock_get, client, admin_headers):
    mock_get.return_value = None
    fake_id = uuid4()
    resp = await client.get(f"/bikes/{fake_id}", headers=admin_headers)
    assert resp.status_code == 404
    assert str(fake_id) in resp.json()["detail"]


@pytest.mark.asyncio
async def test_get_bike_invalid_uuid(client, admin_headers):
    resp = await client.get("/bikes/abc", headers=admin_headers)
    assert resp.status_code == 400


@pytest.mark.asyncio
async def test_get_bike_no_token(client):
    resp = await client.get(f"/bikes/{BIKE_ID}")
    assert resp.status_code == 401


# ===================== PUT /bikes/{id} =====================

@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.update_bike", new_callable=AsyncMock)
async def test_update_bike_success(mock_update, client, admin_headers):
    updated = BikeResponse(
        id=BIKE_ID, brand="Specialized", type="Mountain Bike", color="Red", is_active=True, created_at=NOW
    )
    mock_update.return_value = updated
    resp = await client.put(f"/bikes/{BIKE_ID}", json={"brand": "Specialized"}, headers=admin_headers)
    assert resp.status_code == 200
    assert resp.json()["brand"] == "Specialized"


@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.update_bike", new_callable=AsyncMock)
async def test_update_bike_only_updated_fields_change(mock_update, client, admin_headers):
    updated = BikeResponse(
        id=BIKE_ID, brand="Specialized", type="Mountain Bike", color="Red", is_active=True, created_at=NOW
    )
    mock_update.return_value = updated
    resp = await client.put(f"/bikes/{BIKE_ID}", json={"brand": "Specialized"}, headers=admin_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["brand"] == "Specialized"
    assert data["color"] == "Red"  # unchanged


@pytest.mark.asyncio
async def test_update_bike_id_in_body(client, admin_headers):
    body = {"id": str(uuid4()), "brand": "Trek"}
    resp = await client.put(f"/bikes/{BIKE_ID}", json=body, headers=admin_headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_update_bike_empty_body(client, admin_headers):
    resp = await client.put(f"/bikes/{BIKE_ID}", json={}, headers=admin_headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_update_bike_invalid_type(client, admin_headers):
    resp = await client.put(f"/bikes/{BIKE_ID}", json={"type": "BMX"}, headers=admin_headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.update_bike", new_callable=AsyncMock)
async def test_update_bike_not_found(mock_update, client, admin_headers):
    mock_update.return_value = None
    fake_id = uuid4()
    resp = await client.put(f"/bikes/{fake_id}", json={"brand": "Trek"}, headers=admin_headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_update_bike_user_role_forbidden(client, user_headers):
    resp = await client.put(f"/bikes/{BIKE_ID}", json={"brand": "Trek"}, headers=user_headers)
    assert resp.status_code == 403


# ===================== DELETE /bikes/{id} =====================

@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.delete_bike", new_callable=AsyncMock)
async def test_delete_bike_success(mock_delete, client, admin_headers):
    mock_delete.return_value = True
    resp = await client.delete(f"/bikes/{BIKE_ID}", headers=admin_headers)
    assert resp.status_code == 200
    assert "deleted successfully" in resp.json()["message"]


@pytest.mark.asyncio
@patch("app.services.bikes.BikesService.delete_bike", new_callable=AsyncMock)
async def test_delete_bike_not_found(mock_delete, client, admin_headers):
    mock_delete.return_value = False
    fake_id = uuid4()
    resp = await client.delete(f"/bikes/{fake_id}", headers=admin_headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
@patch("app.messaging.rabbitmq.publish_bike_deleted", new_callable=AsyncMock)
@patch("app.services.bikes.BikesService.delete_bike", new_callable=AsyncMock)
async def test_delete_bike_returns_200_when_rabbitmq_fails(mock_delete, mock_publish, client, admin_headers):
    mock_delete.return_value = True
    resp = await client.delete(f"/bikes/{BIKE_ID}", headers=admin_headers)
    assert resp.status_code == 200


@pytest.mark.asyncio
async def test_delete_bike_user_role_forbidden(client, user_headers):
    resp = await client.delete(f"/bikes/{BIKE_ID}", headers=user_headers)
    assert resp.status_code == 403


@pytest.mark.asyncio
async def test_delete_bike_no_token(client):
    resp = await client.delete(f"/bikes/{BIKE_ID}")
    assert resp.status_code == 401


# ===================== GET /health =====================

@pytest.mark.asyncio
async def test_health_check(client):
    resp = await client.get("/health")
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "ok"
    assert data["service"] == "bicycle-service"
