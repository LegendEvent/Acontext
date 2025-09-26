import asyncio
import traceback
from acontext_core.entry import MQ_CLIENT, LOG, setup, cleanup


async def app(scope, receive, send):
    if scope["type"] == "lifespan":
        while True:
            message = await receive()
            if message["type"] == "lifespan.startup":
                try:
                    await setup()
                except Exception as e:
                    LOG.error(f"Startup failed: {str(e)}")
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
