from .base import Base, CommonMixin
import uuid
from sqlalchemy import (
    Column,
    String,
    Integer,
    ForeignKey,
    ForeignKeyConstraint,
    UniqueConstraint,
)
from sqlalchemy.orm import (
    Mapped,
    mapped_column,
    relationship,
    declarative_mixin,
    declared_attr,
)
from sqlalchemy.dialects.postgresql import JSONB, UUID


class Space(Base, CommonMixin):
    __tablename__ = "spaces"

    __table_args__ = (
        # This ensures id is unique within each project
        UniqueConstraint("project_id", "id"),
    )

    project_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("projects.id", ondelete="CASCADE"),
        primary_key=True,
    )
    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )

    configs: Mapped[dict] = mapped_column(JSONB, nullable=True)

    project: Mapped["Project"] = relationship(  # type: ignore
        "Project", back_populates="spaces"
    )
