from dataclasses import dataclass
from ....infra.db import AsyncSession
from ....schema.utils import asUUID


@dataclass
class TaskCtx:
    db_session: AsyncSession
    session_id: asUUID
    task_ids_index: list[asUUID]
    message_ids_index: list[asUUID]
