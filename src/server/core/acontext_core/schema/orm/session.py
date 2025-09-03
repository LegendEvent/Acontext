from typing import Optional
from .base import Base, CommonMixin
import uuid
from sqlalchemy import ForeignKey, Index
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import JSONB, UUID
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .project import Project
    from .space import Space
    from .message import Message


class Session(Base, CommonMixin):
    __tablename__ = "sessions"

    __table_args__ = (
        Index("ix_session_project_id", "project_id"),
        Index("ix_session_space_id", "space_id"),
        Index("ix_session_session_project_id", "id", "project_id"),
    )

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )

    project_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("projects.id", ondelete="CASCADE"),
        nullable=False,
    )

    space_id: Mapped[Optional[uuid.UUID]] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("spaces.id", ondelete="CASCADE"),
        nullable=True,
    )

    configs: Mapped[dict] = mapped_column(JSONB, nullable=True)

    # Relationships
    project: Mapped["Project"] = relationship("Project", back_populates="sessions")

    space: Mapped[Optional["Space"]] = relationship("Space", back_populates="sessions")

    messages: Mapped[list["Message"]] = relationship(
        "Message", back_populates="session", cascade="all, delete-orphan"
    )
