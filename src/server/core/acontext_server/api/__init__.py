from fastapi import APIRouter
from .api_v1_check import V1_CHECK_ROUTER
from .api_v1_space import V1_SPACE_ROUTER

V1_ROUTER = APIRouter()
V1_ROUTER.include_router(V1_CHECK_ROUTER)
V1_ROUTER.include_router(V1_SPACE_ROUTER)
