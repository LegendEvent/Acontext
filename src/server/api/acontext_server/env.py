import os
import dotenv
from dataclasses import dataclass, field
from .telemetry.log import LOG

dotenv.load_dotenv()


@dataclass
class Env:
    clickhouse_url: str = field(default_factory=lambda: os.getenv("CLICKHOUSE_URL", ""))
    database_url: str = field(default_factory=lambda: os.getenv("DATABASE_URL", ""))
    redis_url: str = field(default_factory=lambda: os.getenv("REDIS_URL", ""))

    database_pool_size: int = field(
        default_factory=lambda: int(os.getenv("DATABASE_POOL_SIZE", 50))
    )
    clickhouse_pool_size: int = field(
        default_factory=lambda: int(os.getenv("CLICKHOUSE_POOL_SIZE", 50))
    )
    redis_pool_size: int = field(
        default_factory=lambda: int(os.getenv("REDIS_POOL_SIZE", 50))
    )

    @classmethod
    def from_env(cls):
        inst = cls()
        LOG.info(f"{inst}")
        return inst


ENV = Env.from_env()
