import uuid
from sqlalchemy.dialects.postgresql import UUID
from datetime import datetime
from dataclasses import dataclass, field
from pydantic import ValidationError
from sqlalchemy.orm import registry
from sqlalchemy import Column
from sqlalchemy.sql import func
from sqlalchemy.types import DateTime

# Create the registry for dataclass ORM
ORM_BASE = registry()


class BaseMixin:
    __sa_dataclass_metadata_key__ = "db"


@dataclass
class TimestampMixin(BaseMixin):

    created_at: datetime = field(
        init=False,
        metadata={
            "db": Column(
                DateTime(timezone=True), server_default=func.now(), nullable=False
            )
        },
    )
    updated_at: datetime = field(
        init=False,
        metadata={
            "db": Column(
                DateTime(timezone=True),
                server_default=func.now(),
                onupdate=func.now(),
                nullable=False,
            )
        },
    )


@dataclass
class CommonMixin(TimestampMixin):
    """Mixin class for common timestamp fields matching GORM autoCreateTime/autoUpdateTime"""

    id: uuid.UUID = field(
        init=False,
        metadata={
            "db": Column(
                UUID(as_uuid=True),
                primary_key=True,
                default=uuid.uuid4,
                server_default=func.gen_random_uuid(),
            )
        },
    )
