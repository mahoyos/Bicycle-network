from sqlalchemy.orm import Session
from . import models, schemas


def get_event(db: Session, event_id: int):
    return db.query(models.Event).filter(models.Event.id == event_id).first()


def get_events(db: Session, skip: int = 0, limit: int = 100, name: str = None, type: str = None, date=None, description: str = None):
    query = db.query(models.Event)
    if name:
        query = query.filter(models.Event.name.ilike(f"%{name}%"))
    if type:
        query = query.filter(models.Event.type == type)
    if date:
        query = query.filter(models.Event.date == date)
    if description:
        query = query.filter(
            models.Event.description.ilike(f"%{description}%"))
    return query.offset(skip).limit(limit).all()


def create_event(db: Session, event: schemas.EventCreate):
    db_event = models.Event(**event.dict())
    db.add(db_event)
    db.flush()
    db.commit()
    return db_event


def update_event(db: Session, event_id: int, event: schemas.EventCreate):
    db_event = db.query(models.Event).filter(
        models.Event.id == event_id).first()
    if db_event:
        for key, value in event.dict().items():
            setattr(db_event, key, value)
        db.commit()
    return db_event


def delete_event(db: Session, event_id: int):
    db_event = db.query(models.Event).filter(
        models.Event.id == event_id).first()
    if db_event:
        db.delete(db_event)
        db.commit()
    return db_event


def create_registration(db: Session, user_id: int, event_id: int):
    db_registration = models.Registration(user_id=user_id, event_id=event_id)
    db.add(db_registration)
    db.flush()
    db.commit()
    return db_registration


def delete_registration(db: Session, user_id: int, event_id: int):
    db_registration = db.query(models.Registration).filter(
        models.Registration.user_id == user_id,
        models.Registration.event_id == event_id
    ).first()
    if db_registration:
        db.delete(db_registration)
        db.commit()
    return db_registration


def get_user_registrations(db: Session, user_id: int):
    return db.query(models.Registration).filter(models.Registration.user_id == user_id).all()


def get_registration(db: Session, user_id: int, event_id: int):
    return db.query(models.Registration).filter(
        models.Registration.user_id == user_id,
        models.Registration.event_id == event_id
    ).first()
