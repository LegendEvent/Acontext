from pydantic import BaseModel
from typing import Generic, TypeVar, Type, Optional
from .error_code import Code

T = TypeVar("T")


class Error(BaseModel):
    status: Code = Code.SUCCESS
    errmsg: str = ""

    @classmethod
    def init(cls, status: Code, errmsg: str) -> "Error":
        return cls(status=status, errmsg=errmsg)

    def __str__(self) -> str:
        return f"Error(status={self.status}, errmsg={self.errmsg})"


class Result(BaseModel, Generic[T]):
    data: Optional[T]
    error: Error

    @classmethod
    def resolve(cls, data: T) -> "Result[T]":
        return cls(data=data, error=Error())

    @classmethod
    def reject(cls, errmsg: str, status: Code = Code.INTERNAL_ERROR) -> "Result[T]":
        assert status != Code.SUCCESS, "status must not be SUCCESS"
        return cls(data=None, error=Error.init(status, errmsg))

    def unpack(self) -> tuple[Optional[T], Optional[Error]]:
        if self.error.status != Code.SUCCESS:
            return None, self.error
        return self.data, None
