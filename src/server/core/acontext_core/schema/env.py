import os
import yaml
from pydantic import BaseModel, Field
from typing import Optional, Any


class CoreConfig(BaseModel):
    llm_api_key: str
    llm_base_url: Optional[str] = None
    logging_format: str = "text"

    # RabbitMQ Configuration
    rabbitmq_url: str = "amqp://acontext:helloworld@localhost:5672/"

    # Database Configuration
    database_pool_size: int = 64
    database_url: str = "postgresql://acontext:helloworld@localhost:5432/acontext"

    # Redis Configuration
    redis_pool_size: int = 32
    redis_url: str = "redis://:helloworld@localhost:6379"


def filter_value_from_env() -> dict[str, Any]:
    config_keys = CoreConfig.model_fields.keys()
    env_already_keys = {}
    for key in config_keys:
        value = os.getenv(key.upper(), None)
        if value is None:
            continue
        env_already_keys[key] = value
    return env_already_keys


def filter_value_from_yaml(yaml_string) -> dict[str, Any]:
    yaml_config_data: dict | None = yaml.safe_load(yaml_string)
    if yaml_config_data is None:
        return {}

    yaml_already_keys = {}
    config_keys = CoreConfig.model_fields.keys()
    for key in config_keys:
        value = yaml_config_data.get(key, None)
        if value is None:
            continue
        yaml_already_keys[key] = value
    return yaml_already_keys
