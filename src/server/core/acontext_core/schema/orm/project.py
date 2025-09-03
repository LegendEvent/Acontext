from .base import Base, CommonMixin
import uuid
from sqlalchemy import String, Index
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import JSONB, UUID
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .space import Space
    from .session import Session


class Project(Base, CommonMixin):
    __tablename__ = "projects"

    __table_args__ = (Index("ix_project_secret_key", "secret_key", unique=True),)

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )

    secret_key: Mapped[str] = mapped_column(String(64), nullable=False)

    configs: Mapped[dict] = mapped_column(JSONB, nullable=True)

    # Relationships
    spaces: Mapped[list["Space"]] = relationship(
        "Space", back_populates="project", cascade="all, delete-orphan"
    )

    sessions: Mapped[list["Session"]] = relationship(
        "Session", back_populates="project", cascade="all, delete-orphan"
    )
