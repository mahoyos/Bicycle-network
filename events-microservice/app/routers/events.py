from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from sqlalchemy.exc import SQLAlchemyError
import logging
from datetime import datetime
from .. import crud, models, schemas
from ..database import get_db

router = APIRouter()


@router.post("/events/", response_model=schemas.Event)
def create_event(event: schemas.EventCreate, db: Session = Depends(get_db)):
    try:
        return crud.create_event(db=db, event=event)
    except SQLAlchemyError as e:
        logging.error(f"Database error while creating event: {e}")
        db.rollback()
        raise HTTPException(
            status_code=500, detail="Database error occurred while creating event")


@router.get("/events/", response_model=list[schemas.Event])
def read_events(skip: int = 0, limit: int = 100, name: str = None, type: str = None, date: datetime = None, description: str = None, db: Session = Depends(get_db)):
    try:
        events = crud.get_events(
            db, skip=skip, limit=limit, name=name, type=type, date=date, description=description)
        return events
    except SQLAlchemyError:
        return []


@router.get("/events/{event_id}", response_model=schemas.Event)
def read_event(event_id: int, db: Session = Depends(get_db)):
    try:
        db_event = crud.get_event(db, event_id=event_id)
        if db_event is None:
            raise HTTPException(status_code=404, detail="Event not found")
        return db_event
    except SQLAlchemyError as e:
        logging.error(f"Database error while fetching event {event_id}: {e}")
        raise HTTPException(
            status_code=500, detail="Database error occurred while fetching event")


@router.put("/events/{event_id}", response_model=schemas.Event)
def update_event(event_id: int, event: schemas.EventCreate, db: Session = Depends(get_db)):
    try:
        db_event = crud.update_event(db, event_id=event_id, event=event)
        if db_event is None:
            raise HTTPException(status_code=404, detail="Event not found")
        return db_event
    except SQLAlchemyError as e:
        logging.error(f"Database error while updating event {event_id}: {e}")
        raise HTTPException(
            status_code=500, detail="Database error occurred while updating event")


@router.post("/events/{event_id}/registrations", response_model=schemas.Registration)
def register_for_event(event_id: int, registration: schemas.RegistrationCreate, db: Session = Depends(get_db)):
    try:
        # Verificar que el evento existe
        event = crud.get_event(db, event_id)
        if not event:
            raise HTTPException(status_code=404, detail="Event not found")
        # Verificar que el usuario no esté ya inscrito
        existing_registration = crud.get_registration(
            db, registration.user_id, event_id)
        if existing_registration:
            raise HTTPException(
                status_code=400, detail="User is already registered for this event")
        return crud.create_registration(db, registration.user_id, event_id)
    except SQLAlchemyError as e:
        logging.error(
            f"Database error while registering for event {event_id}: {e}")
        db.rollback()
        raise HTTPException(
            status_code=500, detail="Database error occurred while registering")


@router.delete("/events/{event_id}/registrations/{user_id}")
def unregister_from_event(event_id: int, user_id: str, db: Session = Depends(get_db)):
    try:
        registration = crud.delete_registration(db, user_id, event_id)
        if not registration:
            raise HTTPException(
                status_code=404, detail="User is not registered for this event")
        return {"message": "Successfully unregistered from event"}
    except SQLAlchemyError as e:
        logging.error(
            f"Database error while unregistering from event {event_id}: {e}")
        db.rollback()
        raise HTTPException(
            status_code=500, detail="Database error occurred while unregistering")


@router.get("/events/registrations/user/{user_id}", response_model=list[schemas.Registration])
def get_user_registrations(user_id: str, db: Session = Depends(get_db)):
    try:
        registrations = crud.get_user_registrations(db, user_id)
        return registrations
    except SQLAlchemyError as e:
        logging.error(
            f"Database error while fetching registrations for user {user_id}: {e}")
        return []


@router.delete("/events/{event_id}")
def delete_event(event_id: int, db: Session = Depends(get_db)):
    try:
        db_event = crud.delete_event(db, event_id=event_id)
        if db_event is None:
            raise HTTPException(status_code=404, detail="Event not found")
        return {"message": "Event deleted successfully"}
    except SQLAlchemyError as e:
        logging.error(f"Database error while deleting event {event_id}: {e}")
        db.rollback()
        raise HTTPException(
            status_code=500, detail="Database error occurred while deleting event")
