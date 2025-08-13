from fastapi import APIRouter, Request, Body, Query
from ..schema.pydantic.api.basic import BasicResponse
from ..schema.pydantic.api.v1.response import (
    MQTaskData,
    SessionMessageStatusCheck,
    SessionTasks,
    SimpleId,
)
from ..schema.pydantic.api.v1.request import (
    LocateProject,
    LocateSession,
    SessionConnectToSpace,
    SessionUpdateConfig,
    SessionPushOpenAIMessage,
    SessionGetScratchpad,
    SessionTasksParams,
)

V1_SESSION_ROUTER = APIRouter()


@V1_SESSION_ROUTER.post("/session")
def create_session(
    request: Request,
    body: LocateProject = Body(...),
) -> BasicResponse[SimpleId]:
    """Create a new session for a project"""
    pass


@V1_SESSION_ROUTER.delete("/session")
def create_session(
    request: Request,
    body: LocateSession = Body(...),
) -> BasicResponse[bool]:
    """Delete a session and its related data"""
    pass


@V1_SESSION_ROUTER.put("/config")
def update_session_config(
    request: Request,
    body: SessionUpdateConfig = Body(...),
) -> BasicResponse[bool]:
    """Create a new session for a project"""
    pass


@V1_SESSION_ROUTER.post("/connect_session_to_space")
def connect_session_to_space(
    request: Request,
    body: SessionConnectToSpace = Body(...),
) -> BasicResponse[bool]:
    """Connect the session to a space, so the learning of this session will be synced to the space"""
    pass


@V1_SESSION_ROUTER.post("/push_openai_messages")
def push_openai_messages_to_session(
    request: Request,
    body: SessionPushOpenAIMessage = Body(...),
) -> BasicResponse[MQTaskData]:
    """Push OpenAI-format messages into this session"""
    pass


@V1_SESSION_ROUTER.get("/session_scratchpad")
def get_session_scratchpad(
    request: Request,
    param: SessionGetScratchpad = Query(...),
) -> BasicResponse[str]:
    """A helper function to pack all the session context so far into a meaningful string"""
    return BasicResponse[str](
        data="Hello",
        status=200,
        errmsg="",
    )


@V1_SESSION_ROUTER.get("/check_messages_status")
def check_messages_status(
    request: Request,
    param: LocateSession = Query(...),
) -> BasicResponse[SessionMessageStatusCheck]:
    pass


@V1_SESSION_ROUTER.get("/fetch_tasks")
def fetch_tasks(
    request: Request,
    param: SessionTasksParams = Query(...),
) -> BasicResponse[SessionTasks]:
    pass
