import json
import logging
import os

import aio_pika
from aio_pika import DeliveryMode, ExchangeType, Message

logger = logging.getLogger(__name__)

connection: aio_pika.RobustConnection | None = None
channel: aio_pika.Channel | None = None
exchange: aio_pika.Exchange | None = None


async def init_rabbitmq():
    global connection, channel, exchange
    rabbitmq_url = os.getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
    try:
        connection = await aio_pika.connect_robust(rabbitmq_url)
        channel = await connection.channel()
        exchange = await channel.declare_exchange(
            "bike_lifecycle_events", ExchangeType.FANOUT, durable=True
        )
        logger.info("RabbitMQ connection established")
    except Exception as e:
        logger.error(f"Failed to connect to RabbitMQ: {e}")


async def close_rabbitmq():
    global connection, channel, exchange
    try:
        if channel:
            await channel.close()
        if connection:
            await connection.close()
    except Exception as e:
        logger.error(f"Error closing RabbitMQ connection: {e}")
    finally:
        connection = None
        channel = None
        exchange = None


def check_rabbitmq() -> bool:
    if exchange is None or channel is None:
        return False
    return not channel.is_closed


async def _publish_event(action: str, bicycle_id: str):
    if exchange is None:
        logger.error(f"RabbitMQ not available. Failed to publish action={action}, bike_id={bicycle_id}")
        return
    try:
        message = Message(
            body=json.dumps({"bike_id": bicycle_id, "action": action}).encode(),
            delivery_mode=DeliveryMode.PERSISTENT,
        )
        await exchange.publish(message, routing_key="")
        logger.info(f"Published action={action}, bike_id={bicycle_id}")
    except Exception as e:
        logger.error(f"Failed to publish action={action}, bike_id={bicycle_id}: {e}")


async def publish_bike_created(bicycle_id: str):
    await _publish_event("CREATED", bicycle_id)


async def publish_bike_deleted(bicycle_id: str):
    await _publish_event("DELETED", bicycle_id)
