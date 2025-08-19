from fastapi import APIRouter, Request, Body, Query
from ..schema.pydantic.api.basic import BasicResponse
from ..schema.pydantic.api.v1.request import (
    UUID,
    JSONProperty,
    SpaceCreatePage,
    SpaceMoveChildrenPages,
    SpaceMoveChildrenBlocks,
    Pagination,
)
from ..schema.pydantic.api.v1.response import SimpleId, PageChildren

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


@V1_SPACE_PAGE_ROUTER.patch("/{page_id}/properties")
def update_page_properties(
    request: Request,
    space_id: UUID,
    page_id: UUID,
    body: JSONProperty = Body(...),
) -> BasicResponse[bool]:
    pass


@V1_SPACE_PAGE_ROUTER.get("/{page_id}/children")
def get_page_children(
    request: Request,
    space_id: UUID,
    page_id: UUID | str,
    param: Pagination = Query(default_factory=Pagination),
) -> BasicResponse[PageChildren]:
    """page_id can be a uuid or `root`"""
    pass


@V1_SPACE_PAGE_ROUTER.patch("/{page_id}/move_children_pages")
def move_children_pages(
    request: Request,
    space_id: UUID,
    page_id: UUID,
    body: SpaceMoveChildrenPages = Body(...),
) -> BasicResponse[bool]:
    pass


@V1_SPACE_PAGE_ROUTER.patch("/{page_id}/move_children_blocks")
def move_children_blocks(
    request: Request,
    space_id: UUID,
    page_id: UUID,
    body: SpaceMoveChildrenBlocks = Body(...),
) -> BasicResponse[bool]:
    pass
