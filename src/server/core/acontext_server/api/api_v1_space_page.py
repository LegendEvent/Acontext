from fastapi import APIRouter, Request, Body, Query
from ..schema.pydantic.api.basic import BasicResponse
from ..schema.pydantic.api.v1.request import UUID, JSONProperty, SpaceCreatePage
from ..schema.pydantic.api.v1.response import SimpleId, SpaceStatusCheck

V1_SPACE_PAGE_ROUTER = APIRouter()


@V1_SPACE_PAGE_ROUTER.post("/")
def create_page(
    request: Request, space_id: UUID, body: SpaceCreatePage = Body(...)
) -> BasicResponse[SimpleId]:
    pass


@V1_SPACE_PAGE_ROUTER.delete("/{page_id}")
def delete_page(request: Request, space_id: UUID, page_id: UUID) -> BasicResponse[bool]:
    pass


@V1_SPACE_PAGE_ROUTER.get("/{page_id}/properties")
def get_page_properties(
    request: Request, space_id: UUID, page_id: UUID
) -> BasicResponse[JSONProperty]:
    pass


@V1_SPACE_PAGE_ROUTER.put("/{page_id}/properties/{property_key}")
def update_page_properties_by_key(
    request: Request,
    space_id: UUID,
    page_id: UUID,
    property_key: str,
    body: JSONProperty = Body(...),
) -> BasicResponse[bool]:
    pass
