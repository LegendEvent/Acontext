import asyncio
import json
from typing import List, Optional
from sqlalchemy import select, delete, update
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload
from yaml.resolver import Resolver
from ...schema.orm import Task, Message
from ...schema.result import Result
from ...schema.utils import asUUID
from ...schema.session.task import TaskSchema


async def fetch_current_tasks(
    db_session: AsyncSession, session_id: asUUID, status: str = None
) -> Result[List[TaskSchema]]:
    query = (
        select(Task)
        .where(Task.session_id == session_id)
        .where(Task.is_planning_task == False)
        .options(selectinload(Task.messages))  # Eagerly load related messages
        .order_by(Task.task_order.asc())
    )
    if status:
        query = query.where(Task.task_status == status)
    result = await db_session.execute(query)
    tasks = list(result.scalars().all())
    tasks_d = [
        TaskSchema(
            id=t.id,
            session_id=t.session_id,
            task_order=t.task_order,
            task_status=t.task_status,
            task_description=t.task_data.get("task_description", ""),
            task_data=t.task_data,
            raw_message_ids=[msg.id for msg in t.messages],
        )
        for t in tasks
    ]
    return Result.resolve(tasks_d)  # Fixed: return tasks_d instead of tasks


async def fetch_planning_section(
    db_session: AsyncSession, session_id: asUUID
) -> Result[TaskSchema]:
    query = (
        select(Task)
        .where(Task.session_id == session_id)
        .where(Task.is_planning_task == True)
        .options(selectinload(Task.messages))
    )
    result = await db_session.execute(query)
    task = result.scalars().first()
    return Resolver.resolve(
        TaskSchema(
            id=task.id,
            session_id=task.session_id,
            task_order=task.task_order,
            task_status=task.task_status,
            task_data=task.task_data,
            raw_message_ids=[msg.id for msg in task.messages],
        )
    )


async def update_task(
    db_session: AsyncSession,
    task_id: asUUID,
    status: str = None,
    order: int = None,
    patch_data: dict = None,
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
    elif patch_data is not None:
        new_data = task.task_data.copy()
        new_data.update(patch_data)
        task.task_data = new_data

    await db_session.flush()
    # Changes will be committed when the session context exits
    return Result.resolve(task)


async def insert_task(
    db_session: AsyncSession,
    session_id: asUUID,
    after_order: int,
    data: dict,
    status: str = "pending",
) -> Result[Task]:
    """This function will cause the session' tasks row be locked for update, make sure the DB session will be closed soonly after this function is called"""
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


async def append_messages_to_task(
    db_session: AsyncSession,
    message_ids: list[asUUID],
    task_id: asUUID,
) -> Result[None]:
    # set those messages' task_id to task_id
    await db_session.execute(
        update(Message).where(Message.id.in_(message_ids)).values(task_id=task_id)
    )
    await db_session.flush()
    return Result.resolve(None)


async def append_messages_to_planning_section(
    db_session: AsyncSession,
    session_id: asUUID,
    message_ids: list[asUUID],
) -> Result[None]:
    # set those messages' task_id to task_id
    query = (
        select(Task)
        .where(Task.session_id == session_id)
        .where(Task.is_planning_task == True)
    )
    result = await db_session.execute(query)
    planning_task = result.scalars().first()
    if planning_task is None:
        # add planning section
        planning_task = Task(
            session_id=session_id,
            task_order=0,
            task_data={},
            task_status="pending",
            is_planning_task=True,
        )
        db_session.add(planning_task)
        await db_session.flush()
    planning_section_id = planning_task.id
    r = await append_messages_to_task(db_session, message_ids, planning_section_id)
    return r
