import asyncio
import json
from typing import List, Optional
from sqlalchemy import select, delete, update
from sqlalchemy.ext.asyncio import AsyncSession
from pydantic import ValidationError
from ...schema.orm import Task
from ...schema.result import Result
from ...schema.utils import asUUID
from ...infra.s3 import S3_CLIENT
from ...env import LOG


async def fetch_current_tasks(
    db_session: AsyncSession, session_id: asUUID, status: str = None
) -> Result[List[Task]]:
    query = (
        select(Task)
        .where(Task.session_id == session_id)
        .order_by(Task.task_order.asc())
    )
    if status:
        query = query.where(Task.task_status == status)
    result = await db_session.execute(query)
    tasks = list(result.scalars().all())
    return Result.resolve(tasks)


async def update_task(
    db_session: AsyncSession,
    task_id: asUUID,
    status: str = None,
    order: int = None,
    data: dict = None,
) -> Result[Task]:
    # Fetch the task to update
    query = select(Task).where(Task.id == task_id)
    result = await db_session.execute(query)
    task = result.scalars().first()

    if task is None:
        return Result.reject(f"Task {task_id} not found")

    # Update only the non-None parameters
    if status is not None:
        task.task_status = status
    if order is not None:
        task.task_order = order
    if data is not None:
        task.task_data = data

    # Changes will be committed when the session context exits
    return Result.resolve(task)


async def insert_task(
    db_session: AsyncSession,
    session_id: asUUID,
    after_order: int,
    data: dict,
    status: str = "pending",
) -> Result[Task]:
    # Lock all tasks in this session to prevent concurrent modifications
    lock_query = (
        select(Task.id)
        .where(Task.session_id == session_id)
        .with_for_update()  # This locks the rows
    )
    await db_session.execute(lock_query)

    # Step 1: Move all tasks that need to be shifted to temporary negative values
    assert after_order >= 0
    temp_update_stmt = (
        update(Task)
        .where(Task.session_id == session_id)
        .where(Task.task_order > after_order)
        .values(task_order=-Task.task_order)
    )
    await db_session.execute(temp_update_stmt)
    await db_session.flush()

    # Step 2: Update them back to positive values, incremented by 1
    final_update_stmt = (
        update(Task)
        .where(Task.session_id == session_id)
        .where(Task.task_order < 0)
        .values(task_order=-Task.task_order + 1)
    )
    await db_session.execute(final_update_stmt)
    await db_session.flush()

    # Step 3: Create new task
    task = Task(
        session_id=session_id,
        task_order=after_order + 1,
        task_data=data,
        task_status=status,
    )

    db_session.add(task)
    await db_session.flush()
    return Result.resolve(task)


async def delete_task(db_session: AsyncSession, task_id: asUUID) -> Result[None]:
    # Fetch the task to delete
    await db_session.execute(delete(Task).where(Task.id == task_id))
    return Result.resolve(None)
