from fastapi import APIRouter, Request, Body, Query
from ..schema.pydantic.api.basic import BasicResponse
from ..schema.pydantic.api.v1.request import (
    UUID,
    JSONConfig,
    LocateSpaceParentPage,
    SpaceCreateBlock,
)
from ..schema.pydantic.api.v1.response import SimpleId, SpaceStatusCheck

V1_SPACE_OPS_ROUTER = APIRouter()


@V1_SPACE_OPS_ROUTER.post("/")
def create_page(
    request: Request, space_id: UUID, body: JSONConfig = Body(...)
) -> BasicResponse[SimpleId]:
    pass


@V1_SPACE_OPS_ROUTER.delete("/{page_id}")
def delete_page(request: Request, space_id: UUID, page_id: UUID) -> BasicResponse[bool]:
    pass


@V1_SPACE_OPS_ROUTER.put("/{page_id}/properties/{property_id}")
def update_page_properties_by_id(
    request: Request,
    space_id: UUID,
    page_id: UUID,
    property_id: UUID,
    body: JSONConfig = Body(...),
) -> BasicResponse[bool]:
    pass


@V1_SPACE_OPS_ROUTER.get("/{page_id}/properties")
def get_page_properties(
    request: Request, space_id: UUID, page_id: UUID
) -> BasicResponse[JSONConfig]:
    pass
