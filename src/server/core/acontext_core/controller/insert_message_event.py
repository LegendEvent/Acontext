from ..env import LOG


async def new_insert_message_event(payload: dict):
    LOG.info(f"New Inserted Message: {payload}")
