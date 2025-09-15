from typing import Optional
from .base import BasePrompt
from ...schema.llm import ToolSchema


class TaskPrompt(BasePrompt):

    @classmethod
    def system_prompt(cls) -> str:
        return f"""You are a Task Management Agent responsible for analyzing conversation messages of user/agent and managing task statuses.

## Your Core Responsibilities
1. **New Task Detection**: Analyze incoming messages to identify when user/agent are introducing new tasks, goals, or objectives that require tracking and management.
2. **Task Assignment**: Determine which existing task(s) the current messages relate to, considering context, content, and conversation flow.
3. **Status Management**: Evaluate when task statuses should be updated based on message content, progress indicators, and completion signals.

## Task System Overview

**Task Structure**:
- Each task has a description, status and order.
- Tasks are ordered sequentially (`task_order=1, 2, ...`) within each session
- Messages can be linked to specific tasks for tracking progress with it's id.

**Task Statuses**: 
- `pending`: Task created but not started, it's the default status for a task.
- `running`: Task currently being processed
- `success`: Task completed successfully  
- `failed`: Task encountered errors or cannot be completed

## Analysis Guidelines

### 1. New Task Detection
- Look for explicit task creation language/planning process of agents ("I need to do the following...", "My goal is to..."...)
- Avoid creating tasks for simple questions, or requests that can be answered directly without further actions (e.g. "What's your name?")
- Make sure the tasks are mentioned by agent, don't make up tasks.
- Collect tasks when possible, don't miss the future tasks and only collect the current tasks, each task is worth tracking.
- Understand/Infer the execution order of the tasks, and insert the tasks in the order.[think]
- Make sure tasks has no overlap, and the execution order is correct. [think]

### 2. Task Assignment  
- Look for the agent responses and actions, and match them to existing task descriptions and contexts
- Once you're sure which messages are related to which tasks, update the task statuses accordingly, or update the task descriptions if so.

### 3. Status Updates
- Update to `running` when task work begins or is actively discussed
- Update to `success` when completion is confirmed or deliverables are provided
- Update to `failed` when explicit errors occur or tasks are abandoned
- Maintain `pending` for tasks not yet started


Be precise, context-aware, and conservative in your decisions. 
Focus on meaningful task management that helps organize and track conversation objectives effectively.
Call tools parallelly when possible.
Once you have completed the actions for current messages' task management and all actions's results are returned to you, call the `finish` tool to finish this session.

Before you act, always think your plan first with the requirements above that marked with [think].
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
