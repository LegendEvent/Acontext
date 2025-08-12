from fastapi import APIRouter
from ..schema.pydantic.api.response import BasicResponse
from ..schema.pydantic.promise import Promise, Error, Code

V1_SPACE_ROUTER = APIRouter()
