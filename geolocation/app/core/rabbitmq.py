import logging
import json
import asyncio
from typing import Callable, Awaitable
import aio_pika
from aio_pika.abc import AbstractIncomingMessage
from app.config import settings

logger = logging.getLogger(__name__)

class RabbitMQManager:
    def __init__(self):
        self.connection = None
        self.channel = None

    async def connect(self):
        logger.info("Connecting to RabbitMQ...")
        loop = asyncio.get_running_loop()
        # Retry connection if RabbitMQ is not ready yet
        for i in range(5):
            try:
                self.connection = await aio_pika.connect_robust(
                    settings.rabbitmq_url, loop=loop
                )
                self.channel = await self.connection.channel()
                await self.channel.set_qos(prefetch_count=10)
                logger.info("Connected to RabbitMQ.")
                return
            except Exception as e:
                logger.warning(f"RabbitMQ connection failed {e}. Retrying in 5s...")
                await asyncio.sleep(5)
        raise Exception("Failed to connect to RabbitMQ after multiple attempts.")

    async def close(self):
        if self.connection:
            logger.info("Closing RabbitMQ connection...")
            await self.connection.close()
            logger.info("RabbitMQ connection closed.")

    async def start_listening(self, queue_name: str, callback: Callable[[dict], Awaitable[None]]):
        if not self.channel:
            raise Exception("Channel not initialized. Call connect() first.")
        
        queue = await self.channel.declare_queue(queue_name, durable=True)
        
        async def process_message(message: AbstractIncomingMessage):
            async with message.process():
                try:
                    payload = json.loads(message.body.decode())
                    logger.debug(f"Received message on {queue_name}: {payload}")
                    await callback(payload)
                except json.JSONDecodeError:
                    logger.error("Failed to decode message body as JSON.")
                except Exception as e:
                    logger.error(f"Error processing message on {queue_name}: {e}")
                    raise

        await queue.consume(process_message)
        logger.info(f"Started listening to queue: {queue_name}")

rabbitmq_manager = RabbitMQManager()
