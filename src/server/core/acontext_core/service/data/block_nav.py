from sqlalchemy import String
from typing import List, Optional
from sqlalchemy import select, delete, update, func
from sqlalchemy import select, delete, update
from sqlalchemy.orm import selectinload
from sqlalchemy.orm.attributes import flag_modified
from sqlalchemy.ext.asyncio import AsyncSession
from ...schema.orm.block import (
    BLOCK_TYPE_FOLDER,
    BLOCK_TYPE_PAGE,
    BLOCK_TYPE_SOP,
    BLOCK_TYPE_TEXT,
)
from ...schema.orm import Block, ToolReference, ToolSOP, Space
from ...schema.utils import asUUID
from ...schema.result import Result
from ...schema.block.sop_block import SOPData


async def fetch_block_by_id(
    db_session: AsyncSession, space_id: asUUID, block_id: asUUID
) -> Result[Block]:
    block = await db_session.get(Block, block_id)
    if block is None:
        return Result.reject(f"Block {block_id} not found")
    return Result.resolve(block)


async def fetch_block_children_by_id(
    db_session: AsyncSession, space_id: asUUID, block_id: asUUID, allow_types: list[str]
) -> Result[List[Block]]:
    query = (
        select(Block)
        .where(Block.space_id == space_id)
        .where(
            Block.parent_id == block_id,
            Block.type.in_(allow_types),
        )
    )
    result = await db_session.execute(query)
    blocks = result.scalars().all()
    return Result.resolve(blocks)


async def list_paths_under_block(
    db_session: AsyncSession,
    space_id: asUUID,
    depth: int,
    block_id: Optional[asUUID] = None,
    path_prefix: str = "",
) -> Result[dict[str, asUUID]]:
    # 1. find the block, assert it either root or folder
    if block_id is not None:
        r = await fetch_block_by_id(db_session, space_id, block_id)
        if not r.ok():
            return r
        folder_block = r.data
        if folder_block.type != BLOCK_TYPE_FOLDER:
            return Result.reject(
                f"Block {block_id}(type {folder_block.type}) is not a folder"
            )

    # 2. list all page and folder block
    r = await fetch_block_children_by_id(
        db_session, space_id, block_id, [BLOCK_TYPE_FOLDER, BLOCK_TYPE_PAGE]
    )
    if not r.ok():
        return r
    blocks = r.data

    # 3. build path dictionary
    path_dict: dict[str, asUUID] = {}

    for block in blocks:
        path_dict[f"{path_prefix}{block.title}"] = block.id

        # Recursively fetch paths for folder blocks
        if block.type == BLOCK_TYPE_FOLDER and depth > 0:
            r = await list_paths_under_block(
                db_session,
                space_id,
                depth - 1,
                block.id,
            )
            if not r.ok():
                return r
            sub_paths = r.data
            # Add sub-paths with current folder as prefix
            for sub_path, sub_id in sub_paths.items():
                path_dict[f"{path_prefix}{block.title}/{sub_path}"] = sub_id

    return Result.resolve(path_dict)
