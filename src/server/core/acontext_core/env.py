from ntpath import exists
import os
from dotenv import load_dotenv

load_dotenv()

from .telemetry.log import get_logger, bound_logging_vars, get_logging_contextvars
from .schema.env import filter_value_from_env, filter_value_from_yaml, CoreConfig


CONFIG_FILE_PATH = os.getenv("CONFIG_FILE_PATH", "config.yaml")

if not os.path.exists(CONFIG_FILE_PATH):
    CONFIG_YAML_STRING = ""
else:
    with open(CONFIG_FILE_PATH) as f:
        CONFIG_YAML_STRING = f.read()

_ENV_VARS = filter_value_from_env()
_YAML_VARS = filter_value_from_yaml(CONFIG_YAML_STRING)

VARS = {**_ENV_VARS, **_YAML_VARS}
CONFIG = CoreConfig(**VARS)
LOG = get_logger(CONFIG.logging_format)

if not os.path.exists(CONFIG_FILE_PATH):
    LOG.warning(f"Your config yaml is not exist: {CONFIG_FILE_PATH}")

LOG.info(f"CONFIG: [{CONFIG}]")
