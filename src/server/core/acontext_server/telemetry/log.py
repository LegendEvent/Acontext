import logging
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


get_global_logger()
