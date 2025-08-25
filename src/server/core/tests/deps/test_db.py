import pytest
from sqlalchemy import select
from acontext_core.infra.db import DB_CLIENT, init_database
from acontext_core.schema.orm import Project, Space, Session


@pytest.mark.asyncio
async def test_db():
    await init_database()

    await DB_CLIENT.health_check()
    print(DB_CLIENT.get_pool_status())
    async with DB_CLIENT.get_session_context() as session:
        p = Project(configs={"name": "Test Project"})
        session.add(p)
        await session.flush()

        s = Space(configs={"name": "asdasd"})
        s.project = p
        session.add(s)
        await session.flush()

        se = Session(configs={"name": "asdasd"}, project_id=p.id)
        se.space_id = s.id
        session.add(se)
        await session.commit()

        pid = p.id
        sid = s.id
        seid = se.id

    async with DB_CLIENT.get_session_context() as session:
        # Use select() instead of session.query()
        se_result = await session.get(Session, (pid, seid))
        print(se_result)
        p = Session.validate_data(configs=se.configs)
        print(p.unpack())

        s_result = await session.get(Space, (pid, sid))
        print(s_result)

        p_result = await session.get(Project, pid)
        print(p_result)
