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


async def app(scope, receive, send):
    if scope["type"] == "lifespan":
        startup_task = None
        while True:
            message = await receive()
            if message["type"] == "lifespan.startup":
                assert await MQ_CLIENT.health_check()
                startup_task = asyncio.create_task(MQ_CLIENT.start())
                await send({"type": "lifespan.startup.complete"})
            elif message["type"] == "lifespan.shutdown":
                await MQ_CLIENT.stop()
                if startup_task:
                    await startup_task
                await send({"type": "lifespan.shutdown.complete"})
                return
    elif scope["type"] == "http":
        await send(
            {
                "type": "http.response.start",
                "status": 404,
                "headers": [(b"content-type", b"text/plain; charset=utf-8")],
            }
        )
        await send({"type": "http.response.body", "body": b"not_found"})
