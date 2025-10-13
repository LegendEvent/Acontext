from ...schema.orm import Session
from ...schema.utils import asUUID
from ...schema.result import Result
from ...infra.db import AsyncSession


async def fetch_session(
    db_session: AsyncSession, session_id: asUUID
) -> Result[Session]:
    session = await db_session.get(Session, session_id)
    if session is None:
        return Result.reject(f"Session {session_id} not found")
    return Result.resolve(session)
