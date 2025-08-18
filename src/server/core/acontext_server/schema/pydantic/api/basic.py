from pydantic import BaseModel
from enum import StrEnum
from typing import Optional, TypeVar, Generic
from ..utils import UUID
from ..error_code import Code

T = TypeVar("T", dict, BaseModel)


class BasicResponse(BaseModel, Generic[T]):
    data: Optional[T] = None
    status: Code = Code.SUCCESS
    errmsg: str = ""


class BlockType(StrEnum):
    TEXT = "text"
    WORKFLOW = "workflow"
