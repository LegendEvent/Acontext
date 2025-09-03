from typing import Optional, List, Dict, Any
from .base import Base, CommonMixin
import uuid
from sqlalchemy import String, ForeignKey, Index, CheckConstraint
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import JSONB, UUID
from pydantic import BaseModel
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .session import Session
    from .message import Message
    from .asset import Asset
    from .message_asset import MessageAsset


class Part(BaseModel):
    """Message part model matching the GORM Part struct"""

    type: str  # "text" | "image" | "audio" | "video" | "file" | "tool-call" | "tool-result" | "data"

    # text part
    text: Optional[str] = None

    # media part
    asset_id: Optional[uuid.UUID] = None
    mime: Optional[str] = None
    filename: Optional[str] = None
    size_b: Optional[int] = None

    # metadata for embedding, ocr, asr, caption, etc.
    meta: Optional[Dict[str, Any]] = None


class Message(Base, CommonMixin):
    __tablename__ = "messages"

    __table_args__ = (
        CheckConstraint(
            "role IN ('user', 'assistant', 'system', 'tool', 'function')",
            name="ck_message_role",
        ),
        Index("ix_message_session_id", "session_id"),
        Index("ix_message_parent_id", "parent_id"),
        Index(
            "ix_message_session_task_status",
            "session_id",
            "session_task_process_status",
        ),
    )

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )

    session_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("sessions.id", ondelete="CASCADE"),
        nullable=False,
    )

    session_task_process_status: Mapped[str] = mapped_column(
        String, nullable=False, server_default="pending"
    )

    parent_id: Mapped[Optional[uuid.UUID]] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("messages.id", ondelete="CASCADE"),
        nullable=True,
    )

    role: Mapped[str] = mapped_column(String, nullable=False)

    parts: Mapped[List[Part]] = mapped_column(JSONB, nullable=False)

    # Relationships
    session: Mapped["Session"] = relationship("Session", back_populates="messages")

    parent: Mapped[Optional["Message"]] = relationship(
        "Message", remote_side="Message.id", back_populates="children"
    )

    children: Mapped[list["Message"]] = relationship(
        "Message", back_populates="parent", cascade="all, delete-orphan"
    )

    assets: Mapped[list["Asset"]] = relationship(
        "Asset", secondary="message_assets", back_populates="messages"
    )
