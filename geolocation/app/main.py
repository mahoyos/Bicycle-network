import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI
from app.config import settings
from app.core.database import db_manager
from app.core.rabbitmq import rabbitmq_manager
from app.api.routes import router as locations_router
from app.services.event_handler import handle_lifecycle_event, handle_location_event

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    logger.info("Starting up Geolocation Microservice...")
    await db_manager.connect()
    await rabbitmq_manager.connect()
    
    # Start consumer for lifecycle events
    await rabbitmq_manager.start_listening(
        queue_name=settings.rabbitmq_queue_lifecycle,
        callback=handle_lifecycle_event
    )
    
    # Start consumer for location updates
    await rabbitmq_manager.start_listening(
        queue_name=settings.rabbitmq_queue_location,
        callback=handle_location_event
    )
    
    yield
    
    # Shutdown
    logger.info("Shutting down Geolocation Microservice...")
    await rabbitmq_manager.close()
    await db_manager.close()

app = FastAPI(
    title="Geolocation Microservice",
    description="Microservice to manage and expose bike locations",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(locations_router, prefix=settings.api_prefix)

@app.get("/health", tags=["Health"], summary="Health Check")
async def health_check():
    return {"status": "ok"}
