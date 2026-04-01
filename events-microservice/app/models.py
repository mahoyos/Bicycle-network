from sqlalchemy import Column, Integer, Float, String, DateTime, Enum, ForeignKey
from sqlalchemy.sql import func
from .database import Base
import enum


class EventType(enum.Enum):
    Route = "Route"
    Tour = "Tour"
    Competition = "Competition"


class Event(Base):
    __tablename__ = "Events"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, index=True)
    type = Column(Enum(EventType), nullable=False)
    date = Column(DateTime(timezone=True), nullable=False)
    description = Column(String, nullable=False)
    start_location_lat = Column(Float, nullable=False)
    start_location_lng = Column(Float, nullable=False)
    end_location_lat = Column(Float, nullable=False)
    end_location_lng = Column(Float, nullable=False)


class Registration(Base):
    __tablename__ = "Registrations"

    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(String, nullable=False)
    event_id = Column(Integer, ForeignKey("Events.id"), nullable=False)
    registration_date = Column(DateTime, default=func.now())
