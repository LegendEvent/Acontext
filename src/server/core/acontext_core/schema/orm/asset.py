import uuid
from dataclasses import dataclass, field
from sqlalchemy import String, BigInteger, Index, Column
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.postgresql import UUID
from typing import TYPE_CHECKING, List, Optional
from .base import ORM_BASE, CommonMixin

if TYPE_CHECKING:
    from .message import Message


@ORM_BASE.mapped
@dataclass
class Asset(CommonMixin):
    __tablename__ = "assets"

    __table_args__ = (Index("u_bucket_key", "bucket", "s3_key", unique=True),)

    bucket: str = field(metadata={"db": Column(String, nullable=False)})

    s3_key: str = field(metadata={"db": Column(String, nullable=False)})

    mime: str = field(metadata={"db": Column(String, nullable=False)})

    size_b: int = field(metadata={"db": Column(BigInteger, nullable=False)})

    etag: Optional[str] = field(
        default=None, metadata={"db": Column(String, nullable=True)}
    )

    sha256: Optional[str] = field(
        default=None, metadata={"db": Column(String, nullable=True)}
    )

    # Relationships
    messages: List["Message"] = field(
        default_factory=list,
        metadata={
            "db": relationship(
                "Message", secondary="message_assets", back_populates="assets"
            )
        },
    )
