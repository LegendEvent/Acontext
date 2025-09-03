from pydantic import BaseModel
from ..utils import UUID


class InsertNewMessage(BaseModel):
    project_id: UUID
    session_id: UUID
    message_id: UUID
