import asyncio
from ..env import LOG, CONFIG
from ..telemetry.log import bound_logging_vars
from ..infra.redis import REDIS_CLIENT
from ..infra.db import DB_CLIENT
from ..infra.async_mq import (
    register_consumer,
    MQ_CLIENT,
    Message,
    ConsumerConfigData,
    SpecialHandler,
)
from ..schema.mq.session import InsertNewMessage
from ..schema.session.task import TaskStatus
from ..schema.session.message import MessageBlob
from .constants import EX, RK
from .data import message as MD
from .controller import message as MC


async def check_session_message_lock_or_set(session_id: str) -> bool:
    async with REDIS_CLIENT.get_client_context() as client:
        _session_lock = f"session.message.insert.lock.{session_id}"
        # Use SET with NX (not exists) and EX (expire) for atomic lock acquisition
        result = await client.set(
            _session_lock,
            "1",
            nx=True,  # Only set if key doesn't exist
            ex=CONFIG.session_message_processing_timeout_seconds,
        )
        # Returns True if the lock was acquired (key didn't exist), False if it already existed
        return result is not None


async def release_session_message_lock(session_id: str):
    async with REDIS_CLIENT.get_client_context() as client:
        _session_lock = f"session.message.insert.lock.{session_id}"
        await client.delete(_session_lock)


@register_consumer(
    mq_client=MQ_CLIENT,
    config=ConsumerConfigData(
        exchange_name=EX.session_message,
        routing_key=RK.session_message_insert,
        queue_name="session.message.insert.entry",
    ),
)
async def insert_new_message(body: InsertNewMessage, message: Message):
    LOG.debug(f"Insert new message {body.message_id}")
    async with DB_CLIENT.get_session_context() as read_session:
        r = await MD.get_latest_message_ids(read_session, body.session_id)
        message_ids, eil = r.unpack()
        if eil:
            LOG.error(f"Exception while fetching session messages: {eil}")
            return
        if not len(message_ids):
            LOG.debug(f"No pending message found for session {body.session_id}, ignore")
            return
        latest_pending_message_id = message_ids[0]
        if body.message_id != latest_pending_message_id:
            LOG.debug(
                f"Message {body.message_id} is not the latest pending message, ignore"
            )
            return

        r = await MD.session_message_length(read_session, body.session_id)
        pending_message_length, eil = r.unpack()
        if eil:
            LOG.error(f"Exception while fetching session messages: {eil}")
            return
        if pending_message_length < CONFIG.session_message_buffer_max_turns:
            LOG.debug(
                f"Session message buffer is not full, wait for next turn/idle notify"
            )
            await MQ_CLIENT.publish(
                exchange_name=EX.session_message,
                routing_key=RK.session_message_buffer_notify,
                body=body.model_dump_json(),
            )
            return

    _l = await check_session_message_lock_or_set(str(body.session_id))
    if not _l:
        LOG.info(
            f"Current Session is processing. "
            f"wait {CONFIG.session_message_session_lock_wait_seconds} seconds for next resend. "
            f"Message {body.message_id}"
        )
        await asyncio.sleep(CONFIG.session_message_session_lock_wait_seconds)
        await message.reject(requeue=True)  # requeue to insert queue
        return

    try:
        LOG.info(
            f"Session message buffer is full (size: {pending_message_length}), start process"
        )
        await MC.process_session_pending_message(body.session_id)
    finally:
        await release_session_message_lock(str(body.session_id))


register_consumer(
    MQ_CLIENT,
    config=ConsumerConfigData(
        exchange_name=EX.session_message,
        routing_key=RK.session_message_buffer_notify,
        queue_name="session.message.buffer.notify",
        message_ttl_seconds=CONFIG.session_message_buffer_ttl_seconds,
        need_dlx_queue=True,
        use_dlx_ex_rk=(EX.session_message, RK.session_message_buffer_process),
    ),
)(SpecialHandler.NO_PROCESS)


@register_consumer(
    mq_client=MQ_CLIENT,
    config=ConsumerConfigData(
        exchange_name=EX.session_message,
        routing_key=RK.session_message_buffer_process,
        queue_name="session.message.buffer.process",
    ),
)
async def buffer_new_message(body: InsertNewMessage, message: Message):
    async with DB_CLIENT.get_session_context() as session:
        r = await MD.get_latest_message_ids(session, body.session_id)
        message_ids, eil = r.unpack()
        if eil:
            LOG.error(f"Exception while fetching latest message id {eil}")
            return
        if not len(message_ids):
            LOG.debug(f"No pending message found for session {body.session_id}, ignore")
            return
        latest_pending_message_id = message_ids[0]
        if body.message_id != latest_pending_message_id:
            LOG.debug(
                f"Message {body.message_id} is not the latest pending message, ignore"
            )
            return
    LOG.info(
        f"Message {body.message_id} IDLE for {CONFIG.session_message_buffer_ttl_seconds} seconds, process it now"
    )
    _l = await check_session_message_lock_or_set(str(body.session_id))
    if not _l:
        LOG.info(
            f"Current Session is processing, resend Message {body.message_id} to insert queue."
        )
        await MQ_CLIENT.publish(
            exchange_name=EX.session_message,
            routing_key=RK.session_message_insert,
            body=body.model_dump_json(),
        )
        return
    try:
        await MC.process_session_pending_message(body.session_id)
    finally:
        await release_session_message_lock(str(body.session_id))
