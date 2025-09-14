import asyncio
import json
from typing import List, Optional
from sqlalchemy import select, func, delete
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
    order: int,
    data: dict,
    status: str = "pending",
) -> Result[asUUID]:
    # Create new task with pending status by default
    task = Task(
        session_id=session_id,
        task_order=order,
        task_data=data,
        task_status=status,
    )

    db_session.add(task)
    await db_session.flush()  # Flush to get the ID

    return Result.resolve(task.id)


async def delete_task(db_session: AsyncSession, task_id: asUUID) -> Result[None]:
    # Fetch the task to delete
    await db_session.execute(delete(Task).where(Task.id == task_id))
    return Result.resolve(None)
