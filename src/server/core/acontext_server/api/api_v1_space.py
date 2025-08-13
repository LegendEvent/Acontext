from fastapi import APIRouter, Request, Body, Query
from ..schema.pydantic.api.basic import BasicResponse
from ..schema.pydantic.api.v1.request import (
    LocateProject,
    LocateSpace,
    SpaceUpdateConfig,
    SpaceFind,
)
from ..schema.pydantic.api.v1.response import SimpleId, SpaceStatusCheck

V1_SPACE_ROUTER = APIRouter()


@V1_SPACE_ROUTER.post("/space")
def create_space(
    request: Request,
    body: LocateProject = Body(...),
) -> BasicResponse[SimpleId]:
    """Create a new space for a project"""
    pass


@V1_SPACE_ROUTER.delete("/space")
def delete_space(
    request: Request,
    body: LocateSpace = Body(...),
) -> BasicResponse[bool]:
    """delete for a project"""
    pass


@V1_SPACE_ROUTER.put("/config")
def update_space_config(
    request: Request,
    body: SpaceUpdateConfig = Body(...),
) -> BasicResponse[bool]:
    """update the config of a space"""
    pass


@V1_SPACE_ROUTER.get("/find")
def find_experiences_in_space(
    request: Request,
    body: SpaceFind = Body(...),
) -> BasicResponse[str]:
    """find experiences in a space"""
    pass


@V1_SPACE_ROUTER.get("/check_space_status")
def check_space_status(
    request: Request,
    param: LocateSpace = Query(...),
) -> BasicResponse[SpaceStatusCheck]:
    pass
