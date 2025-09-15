from typing import Optional
from .base import BasePrompt
from ...schema.llm import ToolSchema


class TaskPrompt(BasePrompt):

    @classmethod
    def system_prompt(cls) -> str:
        return f"""You are a Task Management Agent that analyzes user/agent conversations to manage task statuses.

## Core Responsibilities
1. **New Task Detection**: Identify new tasks, goals, or objectives requiring tracking
2. **Task Assignment**: Match messages to existing tasks based on context and content  
3. **Status Management**: Update task statuses based on progress and completion signals

## Task System
**Structure**: 
- Tasks have description, status, and sequential order (`task_order=1, 2, ...`) within sessions. 
- Messages link to tasks via their IDs.

**Statuses**: 
- `pending`: Created but not started (default)
- `running`: Currently being processed
- `success`: Completed successfully  
- `failed`: Encountered errors or abandoned

## Analysis Guidelines

### New Task Detection
- Look for explicit task language ("I need to...", "My goal is...")
- Avoid creating tasks for simple questions answerable directly
- Only collect tasks mentioned by agents, don't invent them
- Collect all current tasks without missing future ones
- [think] The degree of task splitting should follow the agent's plan in the conversation; do not arbitrarily split into finer or coarser granularity.
- [think] Notice any task modification from agent.
- [think] Infer execution order and insert tasks sequentially, make sure you arrange the tasks in logical execution order, no the mentioned order.
- [think] Ensure no task overlap, make sure the tasks are MECE(mutually exclusive, collectively exhaustive).

### Task Assignment  
- Match agent responses/actions to existing task descriptions and contexts
- No need to link every message, just those messages that are contributed to the process of certain tasks.
- [think] Make sure the messages are contributed to the process of the task, not just doing random linking.
- [think] Update task statuses or descriptions when confident about relationships 

### Status Updates
- `running`: When task work begins or is actively discussed
- `success`: When completion is confirmed or deliverables provided
- `failed`: When explicit errors occur or tasks are abandoned
- `pending`: For tasks not yet started

## Action Guidelines
- Be precise, context-aware, and conservative. 
- Focus on meaningful task management that organizes conversation objectives effectively. 
- Use parallel tool calls when possible. 
After completing all task management actions, call the `finish` tool.
- Before tool calling, use one-two sentences to briefly describe your plan.
"""

    @classmethod
    def pack_task_input(
        cls, previous_messages: str, current_message_with_ids: str, current_tasks: str
    ) -> str:
        return f"""## Current Tasks:
{current_tasks}

## Previous Messages:
{previous_messages}

## Current Message with IDs:
{current_message_with_ids}

Please analyze the above information and determine the actions.
"""

    @classmethod
    def prompt_kwargs(cls) -> str:
        return {"prompt_id": "agent.task"}

    @classmethod
    def tool_schema(cls) -> list[ToolSchema]:
        insert_task_tool = ToolSchema(
            function={
                "name": "insert_task",
                "description": "Create a new task by inserting it after the specified task order. This is used when identifying new tasks from conversation messages.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "after_task_order": {
                            "type": "integer",
                            "description": "The task order after which to insert the new task. Use 0 to insert at the beginning.",
                        },
                        "task_description": {
                            "type": "string",
                            "description": "A clear, concise description of the task, of what's should be done and what's the expected result if any.",
                        },
                    },
                    "required": ["after_task_order", "task_description"],
                },
            }
        )

        update_task_tool = ToolSchema(
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

        append_messages_to_task_tool = ToolSchema(
            function={
                "name": "append_messages_to_task",
                "description": """Link current message ids to a task for tracking progress and context.
Use this to associate conversation messages with relevant tasks.
If the task is marked as 'success' or 'failed', don't append messages to it.""",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "task_order": {
                            "type": "integer",
                            "description": "The order number of the task to link messages to.",
                        },
                        "message_ids": {
                            "type": "array",
                            "items": {"type": "integer"},
                            "description": "List of message IDs to append to the task.",
                        },
                    },
                    "required": ["task_order", "message_ids"],
                },
            }
        )

        finish_tool = ToolSchema(
            function={
                "name": "finish",
                "description": "Call it when you have completed the actions for task management.",
                "parameters": {
                    "type": "object",
                    "properties": {},
                    "required": [],
                },
            }
        )

        return [
            insert_task_tool,
            update_task_tool,
            append_messages_to_task_tool,
            finish_tool,
        ]
