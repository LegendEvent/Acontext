from ..data import message as MD
from ...infra.db import DB_CLIENT
from ...schema.session.task import TaskStatus
from ...schema.session.message import MessageBlob
from ...schema.utils import asUUID
from ...env import LOG, CONFIG
from ...llm.agent import task as AT


async def process_session_pending_message(session_id: asUUID):
    pending_message_ids = None
    try:
        async with DB_CLIENT.get_session_context() as session:
            r = await MD.unpending_session_messages_to_running(session, session_id)
            pending_message_ids, eil = r.unpack()
            if eil:
                LOG.error(f"Exception while unpending session messages: {eil}")
                return

        async with DB_CLIENT.get_session_context() as session:
            r = await MD.fetch_messages_data_by_ids(session, pending_message_ids)
            messages, eil = r.unpack()
            if eil:
                LOG.error(f"Exception while fetching session messages: {eil}")
                return

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
        async with DB_CLIENT.get_session_context() as session:
            await MD.update_message_status_to(
                session, pending_message_ids, TaskStatus.SUCCESS
            )
    except Exception as e:
        if pending_message_ids is None:
            raise e
        LOG.error(
            f"Exception while processing session pending message: {e}, rollback {len(pending_message_ids)} message status to pending"
        )
        async with DB_CLIENT.get_session_context() as session:
            await MD.update_message_status_to(
                session, pending_message_ids, TaskStatus.FAILED
            )
        raise e
