from pydantic import BaseModel
from typing import List
from ..orm import Part
from ..utils import asUUID

STRING_TYPES = {"text", "tool-call", "tool-result"}

REPLACE_NAME = {
    "assistant": "agent",
    "tool": "agent_action",
    "function": "agent_action",
}


def pack_message_line(role: str, part: Part) -> str:
    role = REPLACE_NAME.get(role, role)
    if part.type not in STRING_TYPES:
        return f"<{role}> [{part.type} file: {part.filename}]"
    if part.type == "text":
        return f"<{role}> {part.text}"
    if part.type == "tool-call":
        return f"<{role}> USE TOOL {part.meta['function_name']}, WITH PARAMS {part.meta['parameters']}"


class MessageBlob(BaseModel):
    message_id: asUUID
    role: str
    parts: List[Part]

    def to_string(self) -> str:
        lines = [pack_message_line(self.role, p) for p in self.parts]
        return "\n".join(lines)
