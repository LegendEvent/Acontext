from ..data import message as MD
from ...infra.db import DB_CLIENT
from ...schema.session.task import TaskStatus
from ...schema.session.message import MessageBlob
from ...schema.utils import asUUID
from ...env import LOG, CONFIG


async def process_session_pending_message(session_id: asUUID):
    async with DB_CLIENT.get_session_context() as session:
        r = await MD.fetch_session_messages(session, session_id, status="pending")
        messages, eil = r.unpack()
        if eil:
            LOG.error(f"Exception while fetching session messages: {eil}")
            return
        for m in messages:
            m.session_task_process_status = TaskStatus.RUNNING.value

        messages_data = [
            MessageBlob(message_id=m.id, role=m.role, parts=m.parts) for m in messages
        ]
    for m in messages_data:
        print(m.to_string())

    # TODO
