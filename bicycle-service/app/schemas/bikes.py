from datetime import datetime
from typing import Optional
from uuid import UUID

from pydantic import BaseModel, Field, model_validator
from typing import Literal


VALID_BIKE_TYPES = ["Cross", "Mountain Bike", "Route"]


class BikeCreate(BaseModel):
    brand: str = Field(..., max_length=100)
    type: Literal["Cross", "Mountain Bike", "Route"]
    color: str = Field(..., max_length=100)

    @model_validator(mode="before")
    @classmethod
    def reject_id_field(cls, values):
        if isinstance(values, dict) and "id" in values:
            raise ValueError("Field 'id' is not allowed in the request body")
        return values


class BikeUpdate(BaseModel):
    brand: Optional[str] = Field(None, max_length=100)
    type: Optional[Literal["Cross", "Mountain Bike", "Route"]] = None
    color: Optional[str] = Field(None, max_length=100)

    @model_validator(mode="before")
    @classmethod
    def reject_id_and_require_fields(cls, values):
        if isinstance(values, dict) and "id" in values:
            raise ValueError("Field 'id' is not allowed in the request body")
        if isinstance(values, dict):
            provided = {k: v for k, v in values.items() if v is not None}
            if not provided:
                raise ValueError("At least one field must be provided")
        return values


class BikeResponse(BaseModel):
    id: UUID
    brand: str
    type: str
    color: str
    is_active: bool
    created_at: datetime


class BikeListResponse(BaseModel):
    data: list[BikeResponse]
    total: int
    page: int
    total_pages: int
