from abc import ABC, abstractmethod
from typing import Dict, Any

from ...env import LOG


class TaskHandler(ABC):
    @abstractmethod
    def handle(self, task_data: Dict[str, Any]) -> None:
        """Handle the task data"""
        pass


class ExampleTaskHandler(TaskHandler):
    def handle(self, task_data: Dict[str, Any]) -> None:
        """Example implementation of task handling"""
        task_type = task_data.get("type")
        task_payload = task_data.get("payload", {})

        LOG.info(f"Processing task type: {task_type}")
        LOG.info(f"Task payload: {task_payload}")

        # Add your task processing logic here
        if task_type == "example":
            self._handle_example_task(task_payload)
        else:
            LOG.warning(f"Unknown task type: {task_type}")

    def _handle_example_task(self, payload: Dict[str, Any]) -> None:
        """Handle example task type"""
        # Add your specific task handling logic here
        LOG.info("Handling example task")
