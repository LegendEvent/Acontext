import asyncio
from contextlib import asynccontextmanager
from fastapi import FastAPI
from acontext_core.di import setup, cleanup, MQ_CLIENT, LOG, DB_CLIENT, S3_CLIENT


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    await setup()
    # Run consumer in the background
    asyncio.create_task(MQ_CLIENT.start())
    yield
    # Shutdown
    await cleanup()


app = FastAPI(lifespan=lifespan)


@app.get("/health")
async def health() -> str:
    if not await MQ_CLIENT.health_check():
        return "MQ consumer error"
    if not await DB_CLIENT.health_check():
        return "DB client error"
    if not await S3_CLIENT.health_check():
        return "S3 client error"
    return "ok"
