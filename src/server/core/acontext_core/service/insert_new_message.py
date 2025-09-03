from ..env import LOG, CONFIG
from ..telemetry.log import bound_logging_vars
from ..infra.async_mq import (
    register_consumer,
    MQ_CLIENT,
    Message,
    ConsumerConfigData,
    SpecialHandler,
)
from ..schema.mq.session import InsertNewMessage
from .constants import EX, RK


@register_consumer(
    mq_client=MQ_CLIENT,
    config=ConsumerConfigData(
        exchange_name=EX.session_message,
        routing_key=RK.session_message_insert,
        queue_name="session.message.insert.entry",
    ),
)
async def insert_new_message(body: InsertNewMessage, message: Message):
    with bound_logging_vars(project_id=body.project_id):
        LOG.info(f"Inserting new message, {body}")


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
    with bound_logging_vars(project_id=body.project_id):
        LOG.info(f"New buffer message, {body}")
