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

    task_order: int
    task_description: str
    task_status: TaskStatus
    raw_message_ids: list[asUUID]

    def to_string(self) -> str:
        return f"Task {self.task_order}: {self.task_description} (Status: {self.task_status})"
