from datetime import datetime
import uuid
from pydantic import BaseModel, ConfigDict, ValidationError
from sqlalchemy.orm import DeclarativeBase
from sqlalchemy.orm import Mapped, mapped_column, declarative_mixin
from sqlalchemy.sql import func
from sqlalchemy.types import Integer, DateTime, Boolean
from sqlalchemy.dialects.postgresql import UUID


class Base(DeclarativeBase):
    """Base class for all ORM models with Pydantic integration"""

    # Pydantic configuration for all models
    __pydantic_config__ = ConfigDict(
        from_attributes=True,
        validate_assignment=True,
        arbitrary_types_allowed=True,
        str_strip_whitespace=True,
        validate_default=True,
    )

    def __new__(cls, *args, **kwargs):
        """Override __new__ to add validation before object creation"""
        # Get the Pydantic model for validation
        pydantic_model = getattr(cls, "__use_pydantic__", None)
        if pydantic_model is None:
            return super().__new__(cls)

        try:
            pydantic_model.model_validate(kwargs)
        except ValidationError as e:
            model_name = cls.__name__
            raise ValueError(f"{model_name} validation failed: {e}") from e

        # Create the instance normally
        return super().__new__(cls)


@declarative_mixin
class CommonMixin:
    """Mixin class for common timestamp fields and soft deletion"""

    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), nullable=False
    )
    last_updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
        onupdate=func.now(),
        nullable=False,
    )


@declarative_mixin
class GlobalMixin:
    """For models that don't belong to projects (like Project itself)"""

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
