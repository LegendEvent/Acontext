from .base import Base, CommonMixin
import uuid
from sqlalchemy import ForeignKey, Index
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import JSONB, UUID
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .project import Project
    from .session import Session


class Space(Base, CommonMixin):
    __tablename__ = "spaces"

    __table_args__ = (Index("ix_space_space_project_id", "id", "project_id"),)

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )

    project_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("projects.id", ondelete="CASCADE"),
        nullable=False,
    )

    configs: Mapped[dict] = mapped_column(JSONB, nullable=True)

    # Relationships
    project: Mapped["Project"] = relationship("Project", back_populates="spaces")

    sessions: Mapped[list["Session"]] = relationship(
        "Session", back_populates="space", cascade="all, delete-orphan"
    )
