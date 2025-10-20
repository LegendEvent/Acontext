import re
from typing import List
from urllib import response
from ...env import LOG, DEFAULT_CORE_CONFIG, bound_logging_vars
from ...infra.db import AsyncSession, DB_CLIENT
from ...schema.result import Result
from ...schema.utils import asUUID
from ...schema.session.task import TaskSchema, TaskStatus
from ...schema.session.message import MessageBlob
from ...service.data import task as TD
from ..complete import llm_complete, response_to_sendable_message
from ..prompt.task import TaskPrompt, TASK_TOOLS
from ...util.generate_ids import track_process
from ..tool.sop_lib.ctx import SOPCtx


@track_process
async def sop_agent_curd(
    project_id: asUUID,
    space_id: asUUID,
    planning_task: TaskSchema,
    current_task: TaskSchema,
    max_iterations=3,  # task curd agent only receive one turn of actions
):
    pass
