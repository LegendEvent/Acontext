import asyncio
import traceback
from acontext_core.entry import MQ_CLIENT, setup, cleanup
from acontext_core.infra.async_mq import ConsumerConfigData, register_consumer, Message


consumer_config = ConsumerConfigData(
    queue_name="hello_world_queue",
    exchange_name="hello_exchange",
    routing_key="hello.world",
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
        while True:
            message = await receive()
            if message["type"] == "lifespan.startup":
                try:
                    await setup()
                except Exception as e:
                    print(traceback.format_exc())
                    await send({"type": "lifespan.startup.failed", "message": str(e)})
                    return
                asyncio.create_task(MQ_CLIENT.start())
                await send({"type": "lifespan.startup.complete"})
            elif message["type"] == "lifespan.shutdown":
                await cleanup()
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
