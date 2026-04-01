from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    mongodb_url: str = "mongodb://localhost:27017"
    mongodb_db_name: str = "geolocation_db"
    
    rabbitmq_url: str = "amqp://guest:guest@localhost:5672/"
    rabbitmq_exchange_lifecycle: str = "bike_lifecycle_events"
    rabbitmq_queue_lifecycle: str = "bike_lifecycle_events.geolocation"
    rabbitmq_queue_location: str = "bike_location_events"

    api_prefix: str = "/api/v1"

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

settings = Settings()
