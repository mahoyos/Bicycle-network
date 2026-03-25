from datetime import datetime, timezone
from typing import Optional
from pydantic import BaseModel, Field, ConfigDict

def get_utc_now() -> datetime:
    return datetime.now(timezone.utc)

class Location(BaseModel):
    latitude: float
    longitude: float
    timestamp: datetime = Field(default_factory=get_utc_now)

class Bike(BaseModel):
    id: str = Field(alias="_id")
    location: Optional[Location] = None
    created_at: datetime = Field(default_factory=get_utc_now)
    updated_at: datetime = Field(default_factory=get_utc_now)

    model_config = ConfigDict(populate_by_name=True)

class LocationEvent(BaseModel):
    bike_id: str
    latitude: float
    longitude: float
    timestamp: Optional[datetime] = None

class BikeLifecycleEvent(BaseModel):
    bike_id: str
    action: str  # e.g., "CREATED", "DELETED"
