import logging
from app.repositories.bike_repository import bike_repository
from app.models.bike import BikeLifecycleEvent, LocationEvent

logger = logging.getLogger(__name__)

async def handle_lifecycle_event(payload: dict):
    try:
        event = BikeLifecycleEvent(**payload)
        logger.info(f"Processing lifecycle event: {event.action} for bike {event.bike_id}")
        
        if event.action.upper() == "CREATED":
            await bike_repository.create_bike(event.bike_id)
        elif event.action.upper() == "DELETED":
            await bike_repository.delete_bike(event.bike_id)
        else:
            logger.warning(f"Unknown lifecycle action: {event.action}")
            
    except Exception as e:
        logger.error(f"Error handling lifecycle event: {e}")
        # Depending on requirements, we can choose to swallow validation errors or raise them
        # Re-raising will nack the message in RabbitMQ and potentially put it in a dead letter queue
        raise

async def handle_location_event(payload: dict):
    try:
        event = LocationEvent(**payload)
        logger.info(f"Processing location event for bike {event.bike_id}")
        
        # FR-36 Validate bike before location update. `update_location` only updates if the document exists.
        updated_bike = await bike_repository.update_location(
            bike_id=event.bike_id,
            latitude=event.latitude,
            longitude=event.longitude,
            timestamp=event.timestamp
        )
        
        if not updated_bike:
            logger.warning(f"Location update ignored. Bike {event.bike_id} not found in repository.")
            
    except Exception as e:
        logger.error(f"Error handling location event: {e}")
        raise
