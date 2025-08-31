import asyncio
import json
from dataclasses import dataclass, field
from typing import Callable, Awaitable, Any, Dict, Optional, List
from aio_pika import connect_robust, ExchangeType, Message
from aio_pika.abc import (
    AbstractConnection,
    AbstractChannel,
    AbstractQueue,
    AbstractExchange,
)
from ..env import LOG, CONFIG

BODY_CONTENT_PREVIEW_LENGTH = 100


@dataclass
class ConsumerConfig:
    """Configuration for a single consumer"""

    queue_name: str
    exchange_name: str
    routing_key: str
    handler: Callable[[bytes, Message], Awaitable[Any]] = field(repr=False)
    exchange_type: ExchangeType = ExchangeType.DIRECT
    prefetch_count: int = 10
    durable: bool = True
    auto_delete: bool = False
    exclusive: bool = True
    # Advanced options
    timeout: float = CONFIG.consumer_handler_timeout
    max_retries: int = 3
    retry_delay: float = 5.0
    dead_letter_exchange: Optional[str] = None


@dataclass
class ConnectionConfig:
    """RabbitMQ connection configuration"""

    url: str
    connection_name: str = "acontext_consumer"
    heartbeat: int = 600
    blocked_connection_timeout: int = 300


class AsyncSingleThreadMQConsumer:
    """
    High-performance async RabbitMQ consumer with runtime registration

    Features:
    - Runtime consumer registration
    - Efficient connection pooling
    - Automatic reconnection
    - Error handling and retry logic
    - Dead letter queue support
    - Graceful shutdown
    - Concurrent message processing
    """

    def __init__(self, connection_config: ConnectionConfig):
        self.connection_config = connection_config
        self.connection: Optional[AbstractConnection] = None
        self.channel: Optional[AbstractChannel] = None
        self.consumers: Dict[str, ConsumerConfig] = {}
        self.__running = False
        self.consumer_tasks: List[asyncio.Task] = []
        self._shutdown_event = asyncio.Event()

    @property
    def running(self) -> bool:
        return self.__running

    async def connect(self) -> None:
        """Establish connection to RabbitMQ"""
        try:
            self.connection = await connect_robust(
                self.connection_config.url,
                client_properties={
                    "connection_name": self.connection_config.connection_name
                },
                heartbeat=self.connection_config.heartbeat,
                blocked_connection_timeout=self.connection_config.blocked_connection_timeout,
            )
            self.channel = await self.connection.channel()
            await self.channel.set_qos(
                prefetch_count=CONFIG.rabbitmq_global_qos
            )  # Global prefetch
            LOG.info(
                f"Connected to RabbitMQ (connection: {self.connection_config.connection_name})"
            )
        except Exception as e:
            LOG.error(f"Failed to connect to RabbitMQ: {str(e)}")
            raise

    async def disconnect(self) -> None:
        """Close connection to RabbitMQ"""
        if self.connection and not self.connection.is_closed:
            await self.connection.close()
            LOG.info("Disconnected from RabbitMQ")

    def register_consumer(self, consumer_config: ConsumerConfig) -> None:
        """Register a consumer at runtime"""
        if self.running:
            raise RuntimeError(
                "Cannot register consumers while the consumer is running"
            )

        self.consumers[consumer_config.queue_name] = consumer_config
        LOG.info(
            f"Registered consumer - queue: {consumer_config.queue_name}, "
            f"exchange: {consumer_config.exchange_name}, "
            f"routing_key: {consumer_config.routing_key}"
        )

    async def _process_message(self, config: ConsumerConfig, message: Message) -> None:
        """Process a single message with retry logic"""
        # async with message.process(requeue=True):
        async with message.process(requeue=False):
            retry_count = 0
            max_retries = config.max_retries

            while retry_count <= max_retries:
                try:
                    # process the body to json
                    body = json.loads(message.body.decode("utf-8"))
                    # Call the handler
                    try:
                        await asyncio.wait_for(
                            config.handler(body, message), timeout=config.timeout
                        )
                    except asyncio.TimeoutError:
                        raise TimeoutError(
                            f"Handler timeout after {config.timeout}s - queue: {config.queue_name}"
                        )
                    LOG.info(
                        f"Message processed successfully - queue: {config.queue_name}, "
                        f"body: {message.body[:BODY_CONTENT_PREVIEW_LENGTH]}..."
                    )
                    return  # Success, exit retry loop

                except Exception as e:
                    retry_count += 1
                    _wait_for = config.retry_delay * (retry_count**2)

                    if retry_count <= max_retries:
                        LOG.warning(
                            f"Message processing failed - queue: {config.queue_name}, "
                            f"attempt: {retry_count}/{config.max_retries}, "
                            f"retry after {_wait_for}s, "
                            f"error: {str(e)}"
                        )
                        await asyncio.sleep(_wait_for)  # Exponential backoff
                    else:
                        LOG.error(
                            f"Message processing failed permanently - queue: {config.queue_name}, "
                            f"error: {str(e)}, "
                            f"body: {message.body[:BODY_CONTENT_PREVIEW_LENGTH]}..."
                        )
                        # Message will be sent to DLQ if configured, otherwise rejected
                        raise

    async def _consume_queue(self, config: ConsumerConfig) -> None:
        """Consume messages from a specific queue"""
        try:
            # Set QoS for this consumer
            consumer_channel = await self.connection.channel()
            await consumer_channel.set_qos(prefetch_count=config.prefetch_count)

            queue_info = await self._get_queue_info(consumer_channel, config.queue_name)
            if queue_info is None:
                LOG.info(f"queue: {config.queue_name} doesn't exist, creating")
                queue = await self._setup_consumer_on_channel(config, consumer_channel)
            else:
                LOG.info(f"queue: {config.queue_name} already exists")
                queue = queue_info["queue_object"]

            LOG.info(f"Starting consumer - queue: {config.queue_name}")

            async with queue.iterator() as queue_iter:
                async for message in queue_iter:
                    if self._shutdown_event.is_set():
                        break

                    # Process message in background task for concurrency
                    asyncio.create_task(self._process_message(config, message))

        except Exception as e:
            LOG.error(f"Consumer error - queue: {config.queue_name}, error: {str(e)}")
            raise

    async def _setup_consumer_on_channel(
        self,
        config: ConsumerConfig,
        channel: AbstractChannel,
        queue_arguments: dict = {},
    ) -> AbstractQueue:
        """Setup exchange, queue, and bindings for a consumer on a specific channel"""
        # Declare exchange
        exchange = await channel.declare_exchange(
            config.exchange_name, config.exchange_type, durable=config.durable
        )

        # Setup dead letter exchange if specified
        # TODO: implement dead-letter init
        # if config.dead_letter_exchange:
        #     dlx = await channel.declare_exchange(
        #         config.dead_letter_exchange, ExchangeType.DIRECT, durable=True
        #     )
        #     queue_arguments["x-dead-letter-exchange"] = config.dead_letter_exchange
        #     queue_arguments["x-dead-letter-routing-key"] = f"{config.routing_key}.dead"

        #     # Create dead letter queue
        #     dlq_name = f"{config.queue_name}.dead"
        #     await channel.declare_queue(
        #         dlq_name,
        #         durable=True,
        #         arguments={"x-message-ttl": 86400000},  # 24 hours TTL
        #     )
        #     await dlx.bind(dlq_name, f"{config.routing_key}.dead")

        # Declare queue
        queue = await channel.declare_queue(
            config.queue_name,
            durable=config.durable,
            auto_delete=config.auto_delete,
            exclusive=config.exclusive,
            arguments=queue_arguments,
        )

        # Bind queue to exchange
        await queue.bind(exchange, config.routing_key)

        return queue

    async def _get_queue_info(
        self, channel: AbstractChannel, queue_name: str
    ) -> Optional[Dict[str, Any]]:
        """
        Get detailed information about a queue if it exists.
        Returns None if queue doesn't exist.
        """
        try:
            # Use queue_declare with passive=True to check existence
            queue = await channel.declare_queue(queue_name, passive=True)

            # Get queue information (message count, consumer count, etc.)
            # Note: This requires additional aio_pika features or management API
            return {"name": queue_name, "exists": True, "queue_object": queue}
        except Exception:
            return None

    async def start(self) -> None:
        """Start all registered consumers"""
        if self.running:
            raise RuntimeError("Consumer is already running")

        if not self.consumers:
            raise RuntimeError("No consumers registered")

        if not self.connection or self.connection.is_closed:
            await self.connect()

        self.__running = True
        self._shutdown_event.clear()

        # Start consumer tasks
        for config in self.consumers.values():
            task = asyncio.create_task(self._consume_queue(config))
            self.consumer_tasks.append(task)

        LOG.info(f"Started all consumers (count: {len(self.consumers)})")
        try:
            # Wait for shutdown signal or any task to complete
            done, pending = await asyncio.wait(
                self.consumer_tasks
                + [asyncio.create_task(self._shutdown_event.wait())],
                return_when=asyncio.FIRST_COMPLETED,
            )

            # If shutdown event was triggered, tasks will be cancelled in stop()
            if self._shutdown_event.is_set():
                LOG.info("Shutdown event received")
            else:
                # One of the consumer tasks completed unexpectedly
                for task in done:
                    if task in self.consumer_tasks:
                        try:
                            task.result()  # This will raise the exception if task failed
                        except Exception as e:
                            LOG.error(f"Consumer task failed: {e}")
        finally:
            await self.stop()

    async def stop(self) -> None:
        """Stop all consumers gracefully"""
        if not self.running:
            return

        LOG.info("Stopping consumers...")
        self.__running = False
        self._shutdown_event.set()

        # Cancel all consumer tasks
        for task in self.consumer_tasks:
            task.cancel()

        # Wait for tasks to complete
        if self.consumer_tasks:
            await asyncio.gather(*self.consumer_tasks, return_exceptions=True)

        self.consumer_tasks.clear()
        await self.disconnect()
        LOG.info("All consumers stopped")

    async def health_check(self) -> bool:
        """Check if the consumer is healthy"""
        await self.connect()
        if not self.connection or self.connection.is_closed:
            return False
        return True


# Decorator for easy handler registration
def register_consumer(
    mq_client: AsyncSingleThreadMQConsumer,
    queue_name: str,
    exchange_name: str,
    routing_key: str,
    **kwargs,
):
    """Decorator to register a function as a message handler"""

    def decorator(func: Callable[[bytes, Message], Awaitable[Any]]):
        _consumer_config = ConsumerConfig(
            queue_name=queue_name,
            exchange_name=exchange_name,
            routing_key=routing_key,
            handler=func,
            **kwargs,
        )
        mq_client.register_consumer(_consumer_config)
        return func

    return decorator


MQ_CLIENT = AsyncSingleThreadMQConsumer(
    ConnectionConfig(
        url=CONFIG.rabbitmq_url,
        connection_name=CONFIG.rabbitmq_connection_name,
    )
)
