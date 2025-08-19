from typing import Type

from ...env import LOG
from .connection import RabbitMQConfig, RabbitMQConnection
from .consumer import TaskConsumer
from .handler import TaskHandler, ExampleTaskHandler


def run_consumer(handler_class: Type[TaskHandler] = ExampleTaskHandler) -> None:
    """
    Run the MQ consumer with the specified task handler

    Args:
        handler_class: The task handler class to use for processing messages
    """
    config = RabbitMQConfig.from_config()
    handler = handler_class()

    try:
        with RabbitMQConnection(config) as connection:
            consumer = TaskConsumer(
                connection=connection, callback=handler.handle, prefetch_count=1
            )
            LOG.info("Starting consumer...")
            consumer.start_consuming()
    except KeyboardInterrupt:
        LOG.info("Shutting down consumer...")
    except Exception as e:
        LOG.error(f"Error running consumer: {str(e)}")
        raise


if __name__ == "__main__":
    run_consumer()
