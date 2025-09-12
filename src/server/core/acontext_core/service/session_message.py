from ..env import LOG, CONFIG
from ..telemetry.log import bound_logging_vars
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


@register_consumer(
    mq_client=MQ_CLIENT,
    config=ConsumerConfigData(
        exchange_name=EX.session_message,
        routing_key=RK.session_message_insert,
        queue_name="session.message.insert.entry",
    ),
)
async def insert_new_message(body: InsertNewMessage, message: Message):
    async with DB_CLIENT.get_session_context() as read_session:
        r = await MD.session_message_length(read_session, body.session_id)
        pending_message_length, eil = r.unpack()
        if eil:
            LOG.error(f"Exception while fetching session messages: {eil}")
            return
        if pending_message_length < CONFIG.session_message_buffer_max_turns:
            LOG.info(f"Session message buffer is not full, wait for next turn")
            return
    await MC.process_session_pending_message(body.session_id)


register_consumer(
    MQ_CLIENT,
    config=ConsumerConfigData(
        exchange_name=EX.session_message,
        routing_key=RK.session_message_insert,
        queue_name="session.message.insert.notify.buffer",
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
        r = await MD.check_session_message_status(session, message_id=body.message_id)
        message_status, eil = r.unpack()
        if eil:
            LOG.error(f"Exception while checking message status {eil}")
            return

        if message_status != "pending":
            LOG.info(f"Message {body.message_id} already processed")
            return
    LOG.info(f"Unprocessed message timeout, process it now")
    await MC.process_session_pending_message(body.session_id)
