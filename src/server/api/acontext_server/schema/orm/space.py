from .base import Base, CommonMixin
import uuid
from sqlalchemy import Column, String, Integer, ForeignKey
from sqlalchemy.orm import (
    Mapped,
    mapped_column,
    relationship,
    declarative_mixin,
    declared_attr,
)
from sqlalchemy.dialects.postgresql import JSONB, UUID
from .project import ProjectMixin
from ..pydantic.orm.base import SpaceRow


class Space(Base, CommonMixin, ProjectMixin):
    __tablename__ = "spaces"
    __use_pydantic__ = SpaceRow
    configs: Mapped[dict] = mapped_column(JSONB, nullable=True)


@declarative_mixin
class SpaceMixin:
    """Mixin for models that belong to a space"""

    space_id: Mapped[str] = mapped_column(
        UUID(as_uuid=True), ForeignKey("spaces.id", ondelete="CASCADE"), nullable=False
    )

    @declared_attr
    def space(cls) -> Mapped["Space"]:
        return relationship(
            "Space", back_populates="spaces", cascade="all, delete-orphan"
        )
