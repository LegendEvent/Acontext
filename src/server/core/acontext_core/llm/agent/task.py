from typing import List
from ...infra.db import AsyncSession
from ...schema.result import Result
from ...schema.utils import asUUID
from ...schema.session.task import TaskSchema, TaskStatus
from ...schema.session.message import MessageBlob
from ...service.data import task as TD


def pack_task_section(tasks: List[TaskSchema]) -> str:
    return "\n".join([t.to_string() for t in tasks])


async def process_current_messages(
    db_session: AsyncSession,
    session_id: asUUID,
    previous_messages: List[MessageBlob],
    messages: List[MessageBlob],
) -> Result[None]:
    r = await TD.fetch_current_tasks(db_session, session_id)
    tasks, eil = r.unpack()
    if eil:
        return r

    pack_task_section = pack_task_section(tasks)
    print(pack_task_section)
