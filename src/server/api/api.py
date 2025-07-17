from fastapi import FastAPI
from contextlib import asynccontextmanager
import logging
import uvicorn
from server.api.acontext_server.api.api_v1_router import V1_ROUTER
from acontext_server.env import LOG
from acontext_server.client.db import init_database, close_database


def configure_logging():
    # Configure uvicorn's loggers to use your format
    uvicorn_access = logging.getLogger("uvicorn.access")

    # Clear existing handlers
    uvicorn_access.handlers.clear()

    uvicorn_access.name = "acontext"

    # Add your custom handler to uvicorn loggers
    custom_handler = LOG.handlers[0] if LOG.handlers else None
    if custom_handler:
        uvicorn_access.addHandler(custom_handler)

    # Set log levels
    uvicorn_access.setLevel(logging.INFO)


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Configure logging on startup
    configure_logging()
    await init_database()
    yield
    await close_database()


app = FastAPI(lifespan=lifespan)
app.include_router(V1_ROUTER, prefix="/api/v1")
