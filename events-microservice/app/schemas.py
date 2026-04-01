from pydantic import BaseModel
from datetime import datetime
from typing import Optional
from enum import Enum


class EventType(str, Enum):
    Route = "Route"
    Tour = "Tour"
    Competition = "Competition"


class EventBase(BaseModel):
    name: str
    type: EventType
    date: datetime
    description: str
    start_location_lat: float
    start_location_lng: float
    end_location_lat: float
    end_location_lng: float


class EventCreate(EventBase):
    pass


class Event(EventBase):
    id: int

    class Config:
        from_attributes = True


class RegistrationBase(BaseModel):
    user_id: str
    event_id: int


class RegistrationCreate(BaseModel):
    user_id: str


class Registration(RegistrationBase):
    id: int
    registration_date: datetime

    class Config:
        from_attributes = True
