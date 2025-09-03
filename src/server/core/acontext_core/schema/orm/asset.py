from .base import Base, CommonMixin
import uuid
from sqlalchemy import String, BigInteger, Index
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import UUID
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .message import Message
    from .message_asset import MessageAsset


class Asset(Base, CommonMixin):
    __tablename__ = "assets"

    __table_args__ = (Index("u_bucket_key", "bucket", "s3_key", unique=True),)

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )

    bucket: Mapped[str] = mapped_column(String, nullable=False)

    s3_key: Mapped[str] = mapped_column(String, nullable=False)

    etag: Mapped[str] = mapped_column(String, nullable=True)

    sha256: Mapped[str] = mapped_column(String, nullable=True)

    mime: Mapped[str] = mapped_column(String, nullable=False)

    size_b: Mapped[int] = mapped_column(BigInteger, nullable=False)

    # Relationships
    messages: Mapped[list["Message"]] = relationship(
        "Message", secondary="message_assets", back_populates="assets"
    )
