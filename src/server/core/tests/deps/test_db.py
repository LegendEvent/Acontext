import pytest
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from acontext_core.infra.db import DatabaseClient
from acontext_core.schema.orm import Project, Space, Session

FAKE_KEY = "a" * 32


@pytest.mark.asyncio
async def test_db():
    db_client = DatabaseClient()
    await db_client.create_tables()

    await db_client.health_check()
    print(db_client.get_pool_status())

    async with db_client.get_session_context() as session:
        # check if same p exist
        p_result = await session.execute(
            select(Project).where(Project.secret_key_hmac == FAKE_KEY)
        )
        before_p = p_result.scalars().first()
        if before_p:
            await session.delete(before_p)
            await session.flush()

        p = Project(secret_key_hmac=FAKE_KEY, secret_key_hash_phc=FAKE_KEY)
        session.add(p)
        await session.flush()

        s = Space(project_id=p.id)
        session.add(s)
        await session.flush()

        se = Session(project_id=p.id)
        se.space = s
        session.add(se)
        await session.commit()

        pid = p.id
        sid = s.id
        seid = se.id
    print(pid, sid, seid)
    async with db_client.get_session_context() as session:
        # Use select() with selectinload for session and its space relationship
        se_query = await session.execute(
            select(Session)
            .options(selectinload(Session.space))
            .where(Session.id == seid)
        )
        se_result = se_query.scalar_one()
        print(se_result.id)
        print(se_result.space.id)  # Now this will work without greenlet error
        assert se_result.space_id == sid
        assert se_result.project_id == pid

        s_result = await session.get(Space, sid)
        print(s_result.id)
        assert s_result.project_id == pid

        # Use select() with selectinload for project and its relationships
        p_query = await session.execute(
            select(Project)
            .options(selectinload(Project.sessions), selectinload(Project.spaces))
            .where(Project.id == pid)
        )
        p_result = p_query.scalar_one()
        print(len(p_result.sessions), len(p_result.spaces))
        assert p_result.sessions[0].id == seid
        assert p_result.spaces[0].id == sid
