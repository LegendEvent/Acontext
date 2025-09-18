from typing import List

from ...env import LOG
from ...infra.db import AsyncSession, DB_CLIENT
from ...schema.result import Result
from ...schema.utils import asUUID
from ...schema.session.task import TaskSchema, TaskStatus
from ...schema.session.message import MessageBlob
from ...service.data import task as TD
from ..complete import llm_complete
from ..prompt.task import TaskPrompt
from ...util.generate_ids import track_process


def pack_task_section(tasks: List[TaskSchema]) -> str:
    section = "\n".join([f"- {t.to_string()}" for t in tasks])
    return section


def pack_previous_messages_section(messages: list[MessageBlob]) -> str:
    return "\n".join([m.to_string() for m in messages])


def pack_current_message_with_ids(messages: list[MessageBlob]) -> str:
    return "\n".join(
        [f"<message id={i}> {m.to_string()} </message>" for i, m in enumerate(messages)]
    )


@track_process
async def task_agent_curd(
    session_id: asUUID,
    previous_messages: List[MessageBlob],
    messages: List[MessageBlob],
    max_iterations=3,
) -> Result[None]:
    async with DB_CLIENT.get_session_context() as db_session:
        r = await TD.fetch_current_tasks(db_session, session_id)
        tasks, eil = r.unpack()
        if eil:
            return r

    task_section = pack_task_section(tasks)
    previous_messages_section = pack_previous_messages_section(previous_messages)
    current_messages_section = pack_current_message_with_ids(messages)

    json_tools = [tool.model_dump() for tool in TaskPrompt.tool_schema()]
    already_iterations = 0
    while already_iterations < max_iterations:
        r = await llm_complete(
            prompt=TaskPrompt.pack_task_input(
                previous_messages_section, current_messages_section, task_section
            ),
            system_prompt=TaskPrompt.system_prompt(),
            tools=json_tools,
            prompt_kwargs=TaskPrompt.prompt_kwargs(),
        )
        llm_return, eil = r.unpack()
        if eil:
            return r
        LOG.info(f"LLM Response: {llm_return.model_dump_json()}")
        if not llm_return.tool_calls:
            LOG.info("No tool calls found, stop iterations")
            break

        use_tools = llm_return.tool_calls
        already_iterations += 1
    return r
