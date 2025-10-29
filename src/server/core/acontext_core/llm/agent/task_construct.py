import re
from typing import List, Dict, Any
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
async def construct_agent_curd(
    project_id: asUUID,
    space_id: asUUID,
    task_id: asUUID,
    sop_data: Dict[str, Any],
    max_iterations=3,
) -> Result[Dict[str, Any]]:
    """
    Construct Agent - Process SOP data and build into Space
    
    Args:
        project_id: Project ID
        space_id: Space ID
        task_id: Task ID
        sop_data: SOP data
        max_iterations: Maximum iterations
    
    Returns:
        Result[Dict[str, Any]]: Processing result
    """
    LOG.info(f"Enter construct_agent_curd for task {task_id}")
    
    try:
        # 1. Print received SOP data
        LOG.info(f"SOP Data received: {sop_data}")
        print(f"🔧 Construct Agent - Processing SOP for Task {task_id}")
        print(f"📋 SOP Data: {sop_data}")
        
        # 2. Process SOP data
        processed_result = await _process_sop_data(sop_data)
        
        # 3. Print processing result
        LOG.info(f"Processed result: {processed_result}")
        print(f"✅ SOP Processing Complete!")
        print(f"📊 Result: {processed_result}")
        
        # 4. Mark completion
        completion_status = {
            "status": "completed",
            "task_id": str(task_id),
            "space_id": str(space_id),
            "processed_sop": processed_result,
            "timestamp": "2025-01-28T00:00:00Z"
        }
        
        LOG.info(f"Construct agent completed successfully: {completion_status}")
        print(f"🎯 Construct Agent - COMPLETED!")
        print(f"📈 Final Status: {completion_status}")
        
        return Result.resolve(completion_status)
        
    except Exception as e:
        LOG.error(f"Construct agent error: {e}")
        print(f"❌ Construct Agent Error: {e}")
        return Result.error(f"Construct agent failed: {str(e)}")


async def _process_sop_data(sop_data: Dict[str, Any]) -> Dict[str, Any]:
    """
    Core logic for processing SOP data
    
    Args:
        sop_data: SOP data
    
    Returns:
        Dict[str, Any]: Processing result
    """
    LOG.info("Processing SOP data...")
    
    # Extract SOP information
    use_when = sop_data.get("use_when", "Unknown scenario")
    notes = sop_data.get("notes", "No notes provided")
    sop_steps = sop_data.get("sop", [])
    
    # Process each SOP step
    processed_steps = []
    for i, step in enumerate(sop_steps):
        tool_name = step.get("tool_name", f"unknown_tool_{i}")
        argument_template = step.get("argument_template", {})
        
        processed_step = {
            "step_id": i + 1,
            "tool_name": tool_name,
            "argument_template": argument_template,
            "status": "processed"
        }
        processed_steps.append(processed_step)
    
    # Build processing result
    result = {
        "sop_info": {
            "use_when": use_when,
            "notes": notes,
            "total_steps": len(sop_steps)
        },
        "processed_steps": processed_steps,
        "construction_status": "ready_for_space_insertion"
    }
    
    LOG.info(f"SOP processing completed: {len(processed_steps)} steps processed")
    return result
