from fastapi import APIRouter
from .api_v1_check import V1_CHECK_ROUTER
from .api_v1_space import V1_SPACE_ROUTER
from .api_v1_space_page import V1_SPACE_OPS_ROUTER
from .api_v1_session import V1_SESSION_ROUTER

V1_ROUTER = APIRouter()
V1_ROUTER.include_router(V1_CHECK_ROUTER, prefix="/api/v1/check", tags=["chore"])
V1_ROUTER.include_router(V1_SESSION_ROUTER, prefix="/api/v1/sessions", tags=["session"])
V1_ROUTER.include_router(V1_SPACE_ROUTER, prefix="/api/v1/spaces", tags=["space"])
V1_ROUTER.include_router(
    V1_SPACE_OPS_ROUTER,
    prefix="/api/v1/spaces/{space_id}/pages",
    tags=["space_page"],
)
