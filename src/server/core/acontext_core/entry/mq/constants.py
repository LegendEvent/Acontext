from enum import StrEnum


class Exchange(StrEnum):
    ACONTEXT = "acontext"


class Queue(StrEnum):
    TASKS = "acontext_tasks"


class RoutingKey(StrEnum):
    TASK = "task"
