import os
import sys
import logging
import structlog
from ..util.terminal_color import TerminalColorMarks


def __get_json_logger():
    shared_processors = [
        structlog.contextvars.merge_contextvars,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.stdlib.PositionalArgumentsFormatter(),
        structlog.stdlib.ExtraAdder(),
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.CallsiteParameterAdder(
            [
                structlog.processors.CallsiteParameter.LINENO,
                structlog.processors.CallsiteParameter.PATHNAME,
            ]
        ),
    ]

    structlog_processors = shared_processors + [
        structlog.processors.dict_tracebacks,
        structlog.stdlib.ProcessorFormatter.wrap_for_formatter,
    ]

    structlog.configure(
        processors=structlog_processors,
        logger_factory=structlog.stdlib.LoggerFactory(),
        cache_logger_on_first_use=True,
    )

    formatter = structlog.stdlib.ProcessorFormatter(
        foreign_pre_chain=shared_processors,
        processors=[
            structlog.stdlib.ProcessorFormatter.remove_processors_meta,
            structlog.processors.JSONRenderer(),
        ],
    )

    handler = logging.StreamHandler(stream=sys.stdout)
    handler.setFormatter(formatter)

    logger = logging.getLogger("acontext-core")
    logger.addHandler(handler)
    logger.setLevel(logging.INFO)
    return logger


def __get_text_logger():
    formatter = logging.Formatter(
        f"{TerminalColorMarks.BOLD}{TerminalColorMarks.BLUE}%(name)s |{TerminalColorMarks.END} %(asctime)s - %(levelname)s - %(message)s"
    )
    handler = logging.StreamHandler()
    handler.setFormatter(formatter)
    logger = logging.getLogger("acontext-core")
    logger.setLevel(logging.INFO)
    logger.addHandler(handler)
    return logger


def get_logger(format="text") -> logging.Logger:
    if format == "json":
        LOG = __get_json_logger()
    else:
        LOG = __get_text_logger()
    return LOG
