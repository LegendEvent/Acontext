from re import I
from fastapi import APIRouter, Request, Body, Query
from ..schema.pydantic.api.basic import BasicResponse
from ..schema.pydantic.api.v1.response import (
    MQTaskData,
    SessionMessageStatusCheck,
    SessionTasks,
    SimpleId,
)
from ..schema.pydantic.api.v1.request import (
    UUID,
    JSONConfig,
    LocateSession,
    SessionConnectToSpace,
    SessionUpdateConfig,
    OpenAIMessages,
    SessionScratchpadParams,
    SessionTasksParams,
    LocateSpace,
)

V1_SESSION_ROUTER = APIRouter()


@V1_SESSION_ROUTER.post("/")
def create_session(
    request: Request,
    body: JSONConfig = Body(...),
) -> BasicResponse[SimpleId]:
    """Create a new session for a project"""
    pass


@V1_SESSION_ROUTER.delete("/{session_id}")
def delete_session(
    request: Request,
    session_id: UUID,
) -> BasicResponse[bool]:
    """Delete a session and its related data"""
    pass


@V1_SESSION_ROUTER.put("/{session_id}/configs")
def update_session_config(
    request: Request,
    session_id: UUID,
    body: JSONConfig = Body(...),
) -> BasicResponse[bool]:
    """Create a new session for a project"""
    pass


@V1_SESSION_ROUTER.post("/{session_id}/connect_to_space")
def connect_session_to_space(
    request: Request,
    session_id: UUID,
    body: LocateSpace = Body(...),
) -> BasicResponse[bool]:
    """Connect the session to a space, so the learning of this session will be synced to the space"""
    pass


@V1_SESSION_ROUTER.post("/{session_id}/push_openai_messages", tags=["async"])
def push_openai_messages_to_session(
    request: Request,
    session_id: UUID,
    body: OpenAIMessages = Body(...),
) -> BasicResponse[MQTaskData]:
    """Push OpenAI-format messages into this session"""
    pass


@V1_SESSION_ROUTER.get("/{session_id}/session_scratchpad")
def get_session_scratchpad(
    request: Request,
    session_id: UUID,
    param: SessionScratchpadParams = Query(...),
) -> BasicResponse[str]:
    """A helper function to pack all the session context so far into a meaningful string"""
    return BasicResponse[str](
        data="Hello",
        status=200,
        errmsg="",
    )


@V1_SESSION_ROUTER.get("/{session_id}/check_messages_status")
def check_messages_status(
    request: Request, session_id: UUID
) -> BasicResponse[SessionMessageStatusCheck]:
    pass


@V1_SESSION_ROUTER.get("/{session_id}/fetch_tasks")
def fetch_tasks(
    request: Request,
    session_id: UUID,
    param: SessionTasksParams = Query(...),
) -> BasicResponse[SessionTasks]:
    pass
