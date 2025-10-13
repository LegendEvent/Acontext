from ..data import message as MD
from ...infra.db import DB_CLIENT
from ...schema.session.task import TaskStatus
from ...schema.session.message import MessageBlob
from ...schema.utils import asUUID
from ...llm.agent import task as AT
from ...schema.result import ResultError
from ...env import LOG, DEFAULT_CORE_CONFIG
from ...schema.config import ProjectConfig


async def process_space_pending_task(
    project_config: ProjectConfig, space_id: asUUID, task_id: asUUID
):
    pass
