import json
from typing import Callable, Any, Optional
import pika
from pika.exceptions import AMQPConnectionError, AMQPChannelError
from pika.spec import Basic, BasicProperties

from ...env import LOG
from .connection import RabbitMQConnection, RabbitMQConfig


class TaskConsumer:
    def __init__(
        self,
        connection: RabbitMQConnection,
        callback: Callable[[dict], None],
        prefetch_count: int = 1,
    ):
        self.connection = connection
        self.callback = callback
        self.prefetch_count = prefetch_count
        self._consumer_tag: Optional[str] = None

    def _process_message(
        self,
        channel: pika.channel.Channel,
        method: Basic.Deliver,
        properties: BasicProperties,
        body: bytes,
    ) -> None:
        """Process the received message"""
        try:
            message = json.loads(body.decode())
            LOG.info(f"Received message: {message}")

            # Process the message
            self.callback(message)

            # Acknowledge the message
            channel.basic_ack(delivery_tag=method.delivery_tag)
            LOG.info("Message processed successfully")

        except json.JSONDecodeError as e:
            LOG.error(f"Failed to decode message: {str(e)}")
            # Reject the message without requeue as it's malformed
            channel.basic_reject(delivery_tag=method.delivery_tag, requeue=False)

        except Exception as e:
            LOG.error(f"Error processing message: {str(e)}")
            # Reject and requeue the message for retry
            channel.basic_reject(delivery_tag=method.delivery_tag, requeue=True)

    def start_consuming(self) -> None:
        """Start consuming messages from the queue"""
        while True:
            try:
                channel = self.connection.get_channel()
                if channel is None:
                    LOG.error("Failed to get channel")
                    continue

                # Set QoS
                channel.basic_qos(prefetch_count=self.prefetch_count)

                # Start consuming
                self._consumer_tag = channel.basic_consume(
                    queue=self.connection.config.queue,
                    on_message_callback=self._process_message,
                )

                LOG.info(
                    f"Started consuming from queue: {self.connection.config.queue}"
                )
                channel.start_consuming()

            except (AMQPConnectionError, AMQPChannelError) as e:
                LOG.error(f"Connection error: {str(e)}")
                # Connection will be retried automatically by RabbitMQConnection
                continue

            except Exception as e:
                LOG.error(f"Unexpected error: {str(e)}")
                raise

    def stop_consuming(self) -> None:
        """Stop consuming messages"""
        if self._consumer_tag and self.connection.channel:
            self.connection.channel.basic_cancel(self._consumer_tag)
            self._consumer_tag = None
            LOG.info("Stopped consuming messages")
