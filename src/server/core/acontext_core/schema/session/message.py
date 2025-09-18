from pydantic import BaseModel
from typing import List
from ..orm import Part, ToolCallMeta
from ..utils import asUUID

STRING_TYPES = {"text", "tool-call", "tool-result"}

REPLACE_NAME = {
    "assistant": "agent",
    "tool": "agent_action",
    "tool-result": "agent_action_result",
    "function": "agent_action",
}


def pack_message_line(role: str, part: Part) -> str:
    role = REPLACE_NAME.get(role, role)
    if part.type not in STRING_TYPES:
        return f"<{role}> [{part.type} file: {part.filename}]"
    if part.type == "text":
        return f"<{role}> {part.text}"
    if part.type == "tool-call":
        tool_call_meta = ToolCallMeta(**part.meta)
        return f"<{role}> USE TOOL {tool_call_meta.tool_name}, WITH PARAMS {tool_call_meta.arguments}"


class MessageBlob(BaseModel):
    message_id: asUUID
    role: str
    parts: List[Part]

    def to_string(self) -> str:
        lines = [pack_message_line(self.role, p) for p in self.parts]
        return "\n".join(lines)
