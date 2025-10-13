import asyncio
from ..env import LOG, DEFAULT_CORE_CONFIG
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
from ..schema.mq.space import NewTaskComplete
from .constants import EX, RK
from .data import project as PD
from .data import task as TD
from .data import session as SD
from .controller import space_task as STC


@register_consumer(
    mq_client=MQ_CLIENT,
    config=ConsumerConfigData(
        exchange_name=EX.space_task,
        routing_key=RK.space_task_complete,
        queue_name=RK.space_task_complete,
    ),
)
async def complete_new_task(body: NewTaskComplete, message: Message):
    async with DB_CLIENT.session() as db_session:
        r = await SD.fetch_session(db_session, body.session_id)
        if not r.ok():
            LOG.error(f"Session {body.session_id} not found, error: {r.error}")
            return
        session_data, _ = r.unpack()
        if session_data.space_id is None:
            LOG.info(f"Session {body.session_id} has no linked space")
            return
        SPACE_ID = session_data.space_id
        r = await TD.fetch_task(db_session, body.task_id)
        if not r.ok():
            LOG.error(f"Task {body.task_id} not found, error: {r.error}")
            return
        TASK_DATA, _ = r.unpack()
        r = await TD.set_task_space_digested(db_session, body.task_id)
        if not r.ok():
            LOG.error(
                f"Failed to update task {body.task_id} space digested, error: {r.error}"
            )
            return
        already_digested, _ = r.unpack()
        if already_digested:
            LOG.info(f"Task {body.task_id} is already digested")
            return

        r = await PD.get_project_config(db_session, body.project_id)
        project_config, eil = r.unpack()
        if eil:
            return

    await STC.process_space_pending_task(project_config, SPACE_ID, TASK_DATA.id)
