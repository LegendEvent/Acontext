import pika
from pika.adapters.blocking_connection import BlockingChannel
from pika.exceptions import AMQPConnectionError, AMQPChannelError
import time
from typing import Optional
from dataclasses import dataclass

from ...env import CONFIG, LOG
from .constants import Exchange, Queue, RoutingKey


@dataclass
class RabbitMQConfig:
    url: str
    exchange: Exchange = Exchange.ACONTEXT
    queue: Queue = Queue.TASKS
    routing_key: RoutingKey = RoutingKey.TASK

    @classmethod
    def from_config(cls) -> "RabbitMQConfig":
        return cls(
            url=CONFIG.rabbitmq_url,
        )


class RabbitMQConnection:
    def __init__(self, config: RabbitMQConfig):
        self.config = config
        self.connection = None
        self.channel = None
        self.should_reconnect = True
        self.reconnect_delay = 5  # seconds

    def connect(self) -> None:
        """Establishes connection to RabbitMQ server with automatic reconnection"""
        while self.should_reconnect:
            try:
                if self.connection is None or self.connection.is_closed:
                    parameters = pika.URLParameters(self.config.url)
                    self.connection = pika.BlockingConnection(parameters)
                    LOG.info(f"Connected to RabbitMQ at {self.config.url}")

                if self.channel is None or self.channel.is_closed:
                    self.channel = self.connection.channel()
                    # Declare exchange
                    self.channel.exchange_declare(
                        exchange=self.config.exchange,
                        exchange_type="direct",
                        durable=True,
                    )
                    # Declare queue
                    self.channel.queue_declare(queue=self.config.queue, durable=True)
                    # Bind queue to exchange
                    self.channel.queue_bind(
                        exchange=self.config.exchange,
                        queue=self.config.queue,
                        routing_key=self.config.routing_key,
                    )
                    LOG.info(
                        f"Channel established and queue '{self.config.queue}' bound to exchange '{self.config.exchange}'"
                    )
                break

            except AMQPConnectionError as e:
                LOG.error(f"Failed to connect to RabbitMQ: {str(e)}")
                if self.should_reconnect:
                    time.sleep(self.reconnect_delay)
                    continue
                raise

    def close(self) -> None:
        """Closes the connection to RabbitMQ"""
        self.should_reconnect = False
        if self.channel and not self.channel.is_closed:
            self.channel.close()
        if self.connection and not self.connection.is_closed:
            self.connection.close()
        LOG.info("Closed RabbitMQ connection")

    def get_channel(self) -> Optional[BlockingChannel]:
        """Returns the current channel, reconnecting if necessary"""
        if self.channel is None or self.channel.is_closed:
            self.connect()
        return self.channel

    def __enter__(self) -> "RabbitMQConnection":
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        self.close()
