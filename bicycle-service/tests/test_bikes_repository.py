from datetime import datetime, UTC
from unittest.mock import AsyncMock, MagicMock
from uuid import uuid4

import pytest

from app.repositories.bikes import BikesRepository

BIKE_ID = uuid4()
NOW = datetime.now(UTC)
SAMPLE_ROW = {
    "id": BIKE_ID,
    "brand": "Trek",
    "type": "Mountain Bike",
    "color": "Red",
    "created_at": NOW,
}


class FakeRecord(dict):
    pass


def make_record(data: dict) -> FakeRecord:
    return FakeRecord(data)


class FakeConnection:
    """A fake async connection with async methods."""

    def __init__(self):
        self.fetchrow = AsyncMock()
        self.fetchval = AsyncMock()
        self.fetch = AsyncMock()


class FakeAcquireContext:
    """Mimics the async context manager returned by pool.acquire()."""

    def __init__(self, conn: FakeConnection):
        self.conn = conn

    async def __aenter__(self):
        return self.conn

    async def __aexit__(self, *args):
        pass


@pytest.fixture
def conn():
    return FakeConnection()


@pytest.fixture
def mock_pool(conn):
    pool = MagicMock()
    pool.acquire.return_value = FakeAcquireContext(conn)
    return pool


@pytest.fixture
def repo(mock_pool):
    return BikesRepository(mock_pool)


@pytest.mark.asyncio
async def test_create_executes_insert(repo, conn):
    conn.fetchrow.return_value = make_record(SAMPLE_ROW)

    result = await repo.create("Trek", "Mountain Bike", "Red")

    assert result["brand"] == "Trek"
    call_args = conn.fetchrow.call_args[0]
    assert "INSERT INTO bikes" in call_args[0]
    assert call_args[1] == "Trek"
    assert call_args[2] == "Mountain Bike"
    assert call_args[3] == "Red"


@pytest.mark.asyncio
async def test_get_by_id_queries_by_uuid(repo, conn):
    conn.fetchrow.return_value = make_record(SAMPLE_ROW)

    result = await repo.get_by_id(BIKE_ID)

    assert result["id"] == BIKE_ID
    call_args = conn.fetchrow.call_args[0]
    assert "WHERE id = $1" in call_args[0]


@pytest.mark.asyncio
async def test_get_by_id_returns_none_when_not_found(repo, conn):
    conn.fetchrow.return_value = None

    result = await repo.get_by_id(uuid4())
    assert result is None


@pytest.mark.asyncio
async def test_dynamic_update_includes_only_provided_fields(repo, conn):
    updated = SAMPLE_ROW.copy()
    updated["brand"] = "Specialized"
    conn.fetchrow.return_value = make_record(updated)

    result = await repo.update(BIKE_ID, {"brand": "Specialized"})

    call_args = conn.fetchrow.call_args[0]
    query = call_args[0]
    assert "brand = $1" in query
    assert "color" not in query.split("SET")[1].split("WHERE")[0]
    assert call_args[1] == "Specialized"


@pytest.mark.asyncio
async def test_dynamic_update_multiple_fields(repo, conn):
    updated = SAMPLE_ROW.copy()
    updated["brand"] = "Specialized"
    updated["color"] = "Blue"
    conn.fetchrow.return_value = make_record(updated)

    result = await repo.update(BIKE_ID, {"brand": "Specialized", "color": "Blue"})

    call_args = conn.fetchrow.call_args[0]
    query = call_args[0]
    assert "brand = $1" in query
    assert "color = $2" in query


@pytest.mark.asyncio
async def test_pagination_offset_calculated_correctly(repo, conn):
    conn.fetchval.return_value = 25
    conn.fetch.return_value = [make_record(SAMPLE_ROW)]

    await repo.list_bikes(page=3, limit=10)

    # The fetch call should have offset = (3-1)*10 = 20
    fetch_args = conn.fetch.call_args[0]
    assert fetch_args[1] == 10   # limit
    assert fetch_args[2] == 20   # offset


@pytest.mark.asyncio
async def test_type_filter_uses_any_clause(repo, conn):
    conn.fetchval.return_value = 5
    conn.fetch.return_value = [make_record(SAMPLE_ROW)]

    await repo.list_bikes(types=["Cross", "Route"], page=1, limit=10)

    fetch_args = conn.fetch.call_args[0]
    query = fetch_args[0]
    assert "ANY($1)" in query


@pytest.mark.asyncio
async def test_list_without_type_filter(repo, conn):
    conn.fetchval.return_value = 0
    conn.fetch.return_value = []

    bikes, total = await repo.list_bikes(page=1, limit=10)

    assert bikes == []
    assert total == 0
    fetch_args = conn.fetch.call_args[0]
    assert "ANY" not in fetch_args[0]


@pytest.mark.asyncio
async def test_delete_returns_true_when_found(repo, conn):
    conn.fetchrow.return_value = make_record({"id": BIKE_ID})

    result = await repo.delete(BIKE_ID)
    assert result is True


@pytest.mark.asyncio
async def test_delete_returns_false_when_not_found(repo, conn):
    conn.fetchrow.return_value = None

    result = await repo.delete(uuid4())
    assert result is False
