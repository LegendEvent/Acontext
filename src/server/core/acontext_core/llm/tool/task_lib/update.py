from typing import Any
from ....infra.db import AsyncSession
from ..base import Tool, ToolPool
from ....schema.llm import ToolSchema
from ....schema.utils import asUUID
from ....schema.result import Result
from ....schema.orm import Task
from ....service.data import task as TD
from ....env import LOG
from .ctx import TaskCtx


async def update_task_handler(
    ctx: TaskCtx,
    llm_arguments: dict,
) -> Result[str]:
    task_order = llm_arguments.get("task_order", None)
    if not task_order:
        return Result.resolve(
            f"You must provide a task order argument, so that we can update the task. Updating failed."
        )
    if task_order > len(ctx.task_ids_index) or task_order < 1:
        return Result.resolve(
            f"Task order {task_order} is out of range, updating failed."
        )
    actually_task_id = ctx.task_ids_index[task_order - 1]
    status = llm_arguments.get("task_status", None)
    description = llm_arguments.get("task_description", None)
    r = await TD.update_task(
        ctx.db_session,
        actually_task_id,
        status=status,
        patch_data={
            "task_description": description,
        },
    )
    t, eil = r.unpack()
    if eil:
        return r
    return Result.resolve(f"Task {t.task_order} updated")


_update_task_tool = (
    Tool()
    .use_schema(
        ToolSchema(
            function={
                "name": "update_task",
                "description": """Update an existing task's description and/or status. 
Use this when task progress changes or task details need modification.
Mostly use it to update the task status, if you're confident about a task is running, completed or failed.
Only when the conversation explicitly mention certain task's purpose should be modified, then use this tool to update the task description.""",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "task_order": {
                            "type": "integer",
                            "description": "The order number of the task to update.",
                        },
                        "task_status": {
                            "type": "string",
                            "enum": ["pending", "running", "success", "failed"],
                            "description": "New status for the task. Use 'pending' for not started, 'running' for in progress, 'success' for completed, 'failed' for encountered errors.",
                        },
                        "task_description": {
                            "type": "string",
                            "description": "Update description for the task, of what's should be done and what's the expected result if any. (optional).",
                        },
                    },
                    "required": ["task_order"],
                },
            }
        )
    )
    .use_handler(update_task_handler)
)
