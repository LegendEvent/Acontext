from enum import StrEnum
from pydantic import BaseModel
from ..utils import asUUID


class TaskStatus(StrEnum):
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"


class TaskSchema(BaseModel):
    id: asUUID
    session_id: asUUID

    order: int
    task_description: str
    status: TaskStatus
    data: dict
    space_digested: bool
    raw_message_ids: list[asUUID]

    def to_string(self) -> str:
        return f"Task {self.order}: {self.task_description} (Status: {self.status})"
