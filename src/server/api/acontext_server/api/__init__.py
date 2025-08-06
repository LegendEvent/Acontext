from fastapi import APIRouter
from .api_v1_check import V1_CHECK_ROUTER

V1_ROUTER = APIRouter()
V1_ROUTER.include_router(V1_CHECK_ROUTER)
