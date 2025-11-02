import pytest
import uuid
from sqlalchemy.ext.asyncio import AsyncSession
from acontext_core.schema.orm import Block, Project, Space
from acontext_core.schema.orm.block import (
    BLOCK_TYPE_FOLDER,
    BLOCK_TYPE_PAGE,
    BLOCK_TYPE_SOP,
    BLOCK_TYPE_TEXT,
)
from acontext_core.infra.db import DatabaseClient
from acontext_core.service.data.block import create_new_path_block
from acontext_core.service.data.block_nav import list_paths_under_block


class TestListPathsUnderBlock:
    @pytest.mark.asyncio
    async def test_list_paths_empty_space(self):
        """Test listing paths in an empty space returns empty dict"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            # List paths at root with no blocks created
            r = await list_paths_under_block(session, space.id, depth=0)
            assert r.ok()
            assert r.data == {}

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_list_paths_root_level_depth_zero(self):
        """Test listing paths at root level with depth=0 (no recursion)"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            # Create some root-level pages and folders
            r = await create_new_path_block(
                session, space.id, "Page1", type=BLOCK_TYPE_PAGE
            )
            assert r.ok()
            page1_id = r.data.id

            r = await create_new_path_block(
                session, space.id, "Folder1", type=BLOCK_TYPE_FOLDER
            )
            assert r.ok()
            folder1_id = r.data.id

            r = await create_new_path_block(
                session, space.id, "Page2", type=BLOCK_TYPE_PAGE
            )
            assert r.ok()
            page2_id = r.data.id

            # List paths at root
            r = await list_paths_under_block(session, space.id, depth=0)
            assert r.ok()
            paths = r.data

            assert len(paths) == 3
            assert paths["Page1"] == page1_id
            assert paths["Folder1"] == folder1_id
            assert paths["Page2"] == page2_id

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_list_paths_with_nested_structure(self):
        """Test listing paths with nested folder structure and depth=1"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            # Create folder structure
            r = await create_new_path_block(
                session, space.id, "Docs", type=BLOCK_TYPE_FOLDER
            )
            assert r.ok()
            docs_id = r.data.id

            # Create pages inside Docs folder
            r = await create_new_path_block(
                session, space.id, "README", par_block_id=docs_id, type=BLOCK_TYPE_PAGE
            )
            assert r.ok()
            readme_id = r.data.id

            r = await create_new_path_block(
                session, space.id, "Guide", par_block_id=docs_id, type=BLOCK_TYPE_PAGE
            )
            assert r.ok()
            guide_id = r.data.id

            # List paths with depth=1
            r = await list_paths_under_block(session, space.id, depth=1)
            assert r.ok()
            paths = r.data

            assert len(paths) == 3
            assert paths["Docs"] == docs_id
            assert paths["Docs/README"] == readme_id
            assert paths["Docs/Guide"] == guide_id

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_list_paths_deep_nesting_with_depth_control(self):
        """Test listing paths with deep nesting and different depth parameters"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            # Create nested structure: Root/Level1/Level2/Page
            r = await create_new_path_block(
                session, space.id, "Root", type=BLOCK_TYPE_FOLDER
            )
            assert r.ok()
            root_id = r.data.id

            r = await create_new_path_block(
                session,
                space.id,
                "Level1",
                par_block_id=root_id,
                type=BLOCK_TYPE_FOLDER,
            )
            assert r.ok()
            level1_id = r.data.id

            r = await create_new_path_block(
                session,
                space.id,
                "Level2",
                par_block_id=level1_id,
                type=BLOCK_TYPE_FOLDER,
            )
            assert r.ok()
            level2_id = r.data.id

            r = await create_new_path_block(
                session,
                space.id,
                "DeepPage",
                par_block_id=level2_id,
                type=BLOCK_TYPE_PAGE,
            )
            assert r.ok()
            deep_page_id = r.data.id

            # Test with depth=0 (only root level)
            r = await list_paths_under_block(session, space.id, depth=0)
            assert r.ok()
            assert len(r.data) == 1
            assert "Root" in r.data

            # Test with depth=1 (shows Root/Level1)
            r = await list_paths_under_block(session, space.id, depth=1)
            assert r.ok()
            assert len(r.data) == 2
            assert "Root" in r.data
            assert "Root/Level1" in r.data

            # Test with depth=2 (shows Root/Level1/Level2)
            r = await list_paths_under_block(session, space.id, depth=2)
            assert r.ok()
            assert len(r.data) == 3
            assert "Root" in r.data
            assert "Root/Level1" in r.data
            assert "Root/Level1/Level2" in r.data

            # Test with depth=3 (shows all including DeepPage)
            r = await list_paths_under_block(session, space.id, depth=3)
            assert r.ok()
            assert len(r.data) == 4
            assert r.data["Root/Level1/Level2/DeepPage"] == deep_page_id

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_list_paths_under_specific_folder(self):
        """Test listing paths under a specific folder (not root)"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            # Create root folders
            r = await create_new_path_block(
                session, space.id, "Folder1", type=BLOCK_TYPE_FOLDER
            )
            folder1_id = r.data.id

            r = await create_new_path_block(
                session, space.id, "Folder2", type=BLOCK_TYPE_FOLDER
            )
            folder2_id = r.data.id

            # Add content to Folder1
            r = await create_new_path_block(
                session,
                space.id,
                "PageA",
                par_block_id=folder1_id,
                type=BLOCK_TYPE_PAGE,
            )
            page_a_id = r.data.id

            r = await create_new_path_block(
                session,
                space.id,
                "PageB",
                par_block_id=folder1_id,
                type=BLOCK_TYPE_PAGE,
            )
            page_b_id = r.data.id

            # Add content to Folder2
            r = await create_new_path_block(
                session,
                space.id,
                "PageC",
                par_block_id=folder2_id,
                type=BLOCK_TYPE_PAGE,
            )
            page_c_id = r.data.id

            # List paths under Folder1 only
            r = await list_paths_under_block(
                session, space.id, depth=0, block_id=folder1_id
            )
            assert r.ok()
            paths = r.data

            assert len(paths) == 2
            assert paths["PageA"] == page_a_id
            assert paths["PageB"] == page_b_id
            # PageC should not be included
            assert "PageC" not in paths

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_list_paths_excludes_sop_and_text_blocks(self):
        """Test that SOP and TEXT blocks are not included in path listing"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            # Create a page
            r = await create_new_path_block(
                session, space.id, "TestPage", type=BLOCK_TYPE_PAGE
            )
            assert r.ok()
            page_id = r.data.id

            # Create text block under page
            r = await create_new_path_block(
                session,
                space.id,
                "TextBlock",
                par_block_id=page_id,
                type=BLOCK_TYPE_TEXT,
                props={"preferences": "test"},
            )
            assert r.ok()

            # List paths - should only show the page, not the text block
            r = await list_paths_under_block(session, space.id, depth=0)
            assert r.ok()
            paths = r.data

            assert len(paths) == 1
            assert paths["TestPage"] == page_id

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_list_paths_invalid_block_id(self):
        """Test that listing paths with non-existent block_id fails"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            fake_block_id = uuid.uuid4()
            r = await list_paths_under_block(
                session, space.id, depth=0, block_id=fake_block_id
            )
            assert not r.ok()
            assert "not found" in r.error.errmsg.lower()

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_list_paths_with_page_block_id_fails(self):
        """Test that using a page as block_id fails (must be folder)"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            # Create a page
            r = await create_new_path_block(
                session, space.id, "TestPage", type=BLOCK_TYPE_PAGE
            )
            assert r.ok()
            page_id = r.data.id

            # Try to list paths under a page (should fail)
            r = await list_paths_under_block(
                session, space.id, depth=0, block_id=page_id
            )
            assert not r.ok()
            assert "not a folder" in r.error.errmsg.lower()

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_list_paths_mixed_content(self):
        """Test listing paths with mixed folders and pages at multiple levels"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            # Create structure:
            # - Projects (folder)
            #   - Project1 (folder)
            #     - README (page)
            #   - Project2 (page)
            # - Notes (page)

            r = await create_new_path_block(
                session, space.id, "Projects", type=BLOCK_TYPE_FOLDER
            )
            projects_id = r.data.id

            r = await create_new_path_block(
                session,
                space.id,
                "Project1",
                par_block_id=projects_id,
                type=BLOCK_TYPE_FOLDER,
            )
            project1_id = r.data.id

            r = await create_new_path_block(
                session,
                space.id,
                "README",
                par_block_id=project1_id,
                type=BLOCK_TYPE_PAGE,
            )
            readme_id = r.data.id

            r = await create_new_path_block(
                session,
                space.id,
                "Project2",
                par_block_id=projects_id,
                type=BLOCK_TYPE_PAGE,
            )
            project2_id = r.data.id

            r = await create_new_path_block(
                session, space.id, "Notes", type=BLOCK_TYPE_PAGE
            )
            notes_id = r.data.id

            # List all paths with depth=2
            r = await list_paths_under_block(session, space.id, depth=2)
            assert r.ok()
            paths = r.data

            assert len(paths) == 5
            assert paths["Projects"] == projects_id
            assert paths["Projects/Project1"] == project1_id
            assert paths["Projects/Project1/README"] == readme_id
            assert paths["Projects/Project2"] == project2_id
            assert paths["Notes"] == notes_id

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_list_paths_with_special_characters_in_title(self):
        """Test listing paths with special characters in titles"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            # Create blocks with special characters
            r = await create_new_path_block(
                session, space.id, "Folder-With-Dashes", type=BLOCK_TYPE_FOLDER
            )
            folder_id = r.data.id

            r = await create_new_path_block(
                session,
                space.id,
                "Page_With_Spaces",
                par_block_id=folder_id,
                type=BLOCK_TYPE_PAGE,
            )
            page_id = r.data.id

            # List paths
            r = await list_paths_under_block(session, space.id, depth=1)
            assert r.ok()
            paths = r.data

            assert "Folder-With-Dashes" in paths
            assert "Folder-With-Dashes/Page_With_Spaces" in paths
            assert paths["Folder-With-Dashes/Page_With_Spaces"] == page_id

            await session.delete(project)
