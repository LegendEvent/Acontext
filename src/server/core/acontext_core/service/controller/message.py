from ..data import message as MD
from ...infra.db import DB_CLIENT
from ...schema.session.task import TaskStatus
from ...schema.session.message import MessageBlob
from ...schema.utils import asUUID
from ...env import LOG, CONFIG
from ...llm.agent import task as AT


async def process_session_pending_message(session_id: asUUID):
    try:
        pending_message_ids = []
        async with DB_CLIENT.get_session_context() as session:
            r = await MD.fetch_session_messages(session, session_id, status="pending")
            messages, eil = r.unpack()
            if eil:
                LOG.error(f"Exception while fetching session messages: {eil}")
                return
            for m in messages:
                m.session_task_process_status = TaskStatus.RUNNING.value
            await session.flush()
            pending_message_ids.extend([m.id for m in messages])
            r = await MD.fetch_previous_messages_by_datetime(
                session, session_id, messages[0].created_at, limit=1
            )
            previous_messages, eil = r.unpack()
            if eil:
                LOG.error(f"Exception while fetching previous messages: {eil}")
                return
            messages_data = [
                MessageBlob(message_id=m.id, role=m.role, parts=m.parts)
                for m in messages
            ]
            previous_messages_data = [
                MessageBlob(message_id=m.id, role=m.role, parts=m.parts)
                for m in previous_messages
            ]

        r = await AT.task_agent_curd(session_id, previous_messages_data, messages_data)
    except Exception as e:
        LOG.error(
            f"Exception while processing session pending message: {e}, rollback {len(pending_message_ids)} message status to pending"
        )
        async with DB_CLIENT.get_session_context() as session:
            await MD.rollback_message_status_to_pending(session, pending_message_ids)
        raise e
