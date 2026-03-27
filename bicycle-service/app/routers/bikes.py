from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query

from app.database.connection import get_pool
from app.dependencies.auth import get_current_user, require_admin
from app.schemas.bikes import (
    VALID_BIKE_TYPES,
    BikeCreate,
    BikeListResponse,
    BikeResponse,
    BikeUpdate,
)
from app.services.bikes import BikesService

router = APIRouter(prefix="/bikes", tags=["bikes"])


def _validate_uuid(bike_id: str) -> UUID:
    try:
        return UUID(bike_id)
    except (ValueError, AttributeError):
        raise HTTPException(status_code=400, detail="Invalid UUID format")


@router.post("", status_code=201, response_model=BikeResponse)
async def create_bike(
    data: BikeCreate,
    _user: dict = Depends(require_admin),
    pool=Depends(get_pool),
):
    service = BikesService(pool)
    try:
        return await service.create_bike(data)
    except RuntimeError as e:
        raise HTTPException(status_code=503, detail=str(e))


@router.get("", response_model=BikeListResponse)
async def list_bikes(
    type: str | None = Query(None),
    page: str = Query("1"),
    limit: str = Query("10"),
    _user: dict = Depends(get_current_user),
    pool=Depends(get_pool),
):
    # Validate page
    try:
        page_int = int(page)
    except ValueError:
        raise HTTPException(status_code=400, detail="Page must be a positive integer")

    if page_int < 1:
        raise HTTPException(status_code=400, detail="Page must be a positive integer")

    # Validate limit
    try:
        limit_int = int(limit)
    except ValueError:
        raise HTTPException(status_code=400, detail="Limit must be a positive integer")

    if limit_int < 1:
        raise HTTPException(status_code=400, detail="Limit must be a positive integer")

    if limit_int > 100:
        raise HTTPException(status_code=400, detail="Limit must not exceed 100")

    # Validate type filter
    types = None
    if type is not None:
        type_values = [t.strip() for t in type.split(",")]
        valid_lower = {v.lower(): v for v in VALID_BIKE_TYPES}
        for t in type_values:
            if t.lower() not in valid_lower:
                raise HTTPException(
                    status_code=400,
                    detail=f"Invalid type '{t}'. Valid types are: {VALID_BIKE_TYPES}",
                )
        types = type_values

    service = BikesService(pool)
    return await service.list_bikes(types=types, page=page_int, limit=limit_int)


@router.get("/{bike_id}", response_model=BikeResponse)
async def get_bike(
    bike_id: str,
    _user: dict = Depends(get_current_user),
    pool=Depends(get_pool),
):
    valid_id = _validate_uuid(bike_id)
    service = BikesService(pool)
    result = await service.get_bike(valid_id)
    if result is None:
        raise HTTPException(status_code=404, detail=f"Bicycle {bike_id} not found")
    return result


@router.put("/{bike_id}", response_model=BikeResponse)
async def update_bike(
    bike_id: str,
    data: BikeUpdate,
    _user: dict = Depends(require_admin),
    pool=Depends(get_pool),
):
    valid_id = _validate_uuid(bike_id)
    service = BikesService(pool)
    result = await service.update_bike(valid_id, data)
    if result is None:
        raise HTTPException(status_code=404, detail=f"Bicycle {bike_id} not found")
    return result


@router.delete("/{bike_id}")
async def delete_bike(
    bike_id: str,
    _user: dict = Depends(require_admin),
    pool=Depends(get_pool),
):
    valid_id = _validate_uuid(bike_id)
    service = BikesService(pool)
    try:
        deleted = await service.delete_bike(valid_id)
    except RuntimeError as e:
        raise HTTPException(status_code=503, detail=str(e))
    if not deleted:
        raise HTTPException(status_code=404, detail=f"Bicycle {bike_id} not found")
    return {"message": f"Bicycle {bike_id} deleted successfully"}
