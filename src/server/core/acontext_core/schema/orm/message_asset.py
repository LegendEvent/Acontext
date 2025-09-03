import uuid
from dataclasses import dataclass, field
from sqlalchemy import ForeignKey, Column
from sqlalchemy.dialects.postgresql import UUID
from typing import TYPE_CHECKING
from .base import ORM_BASE, TimestampMixin

if TYPE_CHECKING:
    from .message import Message
    from .asset import Asset


@ORM_BASE.mapped
@dataclass
class MessageAsset(TimestampMixin):
    __tablename__ = "message_assets"

    message_id: uuid.UUID = field(
        metadata={
            "db": Column(
                UUID(as_uuid=True),
                ForeignKey("messages.id", ondelete="CASCADE"),
                primary_key=True,
                index=True,
            )
        }
    )

    asset_id: uuid.UUID = field(
        metadata={
            "db": Column(
                UUID(as_uuid=True),
                ForeignKey("assets.id", ondelete="CASCADE"),
                primary_key=True,
                index=True,
            )
        }
    )
