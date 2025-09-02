import asyncio
from acontext_core.infra.async_mq import (
    MQ_CLIENT,
    Message,
    ConsumerConfigData,
    register_consumer,
)


consumer_config = ConsumerConfigData(
    queue_name="hello_world_queue",
    exchange_name="hello_exchange",
    routing_key="hello.world",
    # timeout=1,
)


@register_consumer(
    mq_client=MQ_CLIENT,
    config=consumer_config,
)
async def hello_world_handler(body: dict, message: Message) -> None:
    """Simple hello world message handler"""
    print(body)


async def app():
    await MQ_CLIENT.start()


if __name__ == "__main__":
    asyncio.run(app())
