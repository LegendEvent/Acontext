from pydantic import BaseModel
from fastapi.responses import JSONResponse
from typing import Generic, TypeVar, Type, Optional
from .error_code import Code
from .api.response import BasicResponse

T = TypeVar("T")
R = TypeVar("R", bound=BasicResponse)


class Error(BaseModel):
    status: Code = Code.SUCCESS
    errmsg: str = ""

    @classmethod
    def init(cls, status: Code, errmsg: str) -> "Error":
        return cls(status=status, errmsg=errmsg)


class Promise(BaseModel, Generic[T]):
    data: Optional[T]
    error: Error

    @classmethod
    def resolve(cls, data: T) -> "Promise[T]":
        return cls(data=data, error=Error())

    @classmethod
    def reject(cls, error: Error) -> "Promise[T]":
        assert error.status != Code.SUCCESS, "status must not be SUCCESS"
        return cls(data=None, error=error)

    def unpack(self) -> tuple[Optional[T], Optional[BasicResponse]]:
        if self.error.status != Code.SUCCESS:
            return None, self.error
        return self.data, None

    def to_response(self, response_type: Type[R]) -> R:
        val_value = response_type(
            data=self.data, status=self.error.status, errmsg=self.error.errmsg
        )
        return JSONResponse(
            content=val_value.model_dump(),
            status_code=self.error.status.value,
        )
