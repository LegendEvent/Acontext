from pydantic import BaseModel
from datetime import datetime
from ...utils import UUID
from .data import SessionMessageStatus, SessionTaskStatus, BlockType
from .request import JSONProperty


class SimpleId(BaseModel):
    id: UUID


class MQTaskData(BaseModel):
    task_id: UUID


class SessionMessageStatusCheck(BaseModel):
    pending: int
    processing: int
    completed: int
    failed: int


class SessionTask(BaseModel):
    task_step: int
    task_name: str
    task_description: str
    task_status: SessionTaskStatus
    learned_hints: list[str]
    failed_attempts: list[str]


class SessionTasks(BaseModel):
    plan: str
    tasks: list[SessionTask]


class SpaceStatusCheck(BaseModel):
    already_blocks: int
    session_connection_num: int
    last_updated_at: datetime


class ReturnPage(JSONProperty):
    page_id: UUID


class ReturnBlock(JSONProperty):
    block_id: UUID
    type: BlockType


class PageChildren(BaseModel):
    child_pages: list[ReturnPage]
    child_blocks: list[ReturnBlock]


class BlockChildren(BaseModel):
    child_blocks: list[ReturnBlock]
