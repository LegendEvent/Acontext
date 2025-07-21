import logging
import json
from ..util.terminal_color import TerminalColorMarks

LOG = None


def get_global_logger(level: int = logging.INFO):
    global LOG
    if LOG is not None:
        return LOG
    formatter = logging.Formatter(
        f"{TerminalColorMarks.BOLD}{TerminalColorMarks.BLUE}%(name)s |{TerminalColorMarks.END}  %(levelname)s - %(asctime)s  -  %(message)s"
    )
    handler = logging.StreamHandler()
    handler.setFormatter(formatter)
    logger = logging.getLogger("acontext")
    logger.setLevel(level)
    logger.addHandler(handler)
    LOG = logger
    return LOG


def L_(project_id, space_id, messages, session_id=None, **kwargs):
    headers = [f"(project_id: {project_id})", f"(space_id: {space_id})"]
    if session_id:
        headers.append(f"(session_id: {session_id})")
    for k, v in kwargs.items():
        headers.append(f"({k}: {v})")
    log_header = " ".join(headers)
    oneline_message = messages.replace("\n", "; ")
    return f"{log_header} {oneline_message}"


get_global_logger()
