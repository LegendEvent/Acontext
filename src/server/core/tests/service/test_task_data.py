import pytest
import uuid
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession
from acontext_core.service.data.task import (
    fetch_current_tasks,
    update_task,
    insert_task,
    delete_task,
)
from acontext_core.schema.orm import Task, Project, Space, Session
from acontext_core.schema.result import Result
from acontext_core.schema.error_code import Code
from acontext_core.infra.db import DatabaseClient


class TestFetchCurrentTasks:
    @pytest.mark.asyncio
    async def test_fetch_all_tasks_success(self):
        """Test fetching all tasks for a session"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac", secret_key_hash_phc="test_key_hash"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            # Create sample tasks
            tasks_data = [
                {
                    "session_id": test_session.id,
                    "task_order": 1,
                    "task_data": {"name": "task1", "description": "First task"},
                    "task_status": "pending",
                },
                {
                    "session_id": test_session.id,
                    "task_order": 2,
                    "task_data": {"name": "task2", "description": "Second task"},
                    "task_status": "running",
                },
                {
                    "session_id": test_session.id,
                    "task_order": 3,
                    "task_data": {"name": "task3", "description": "Third task"},
                    "task_status": "success",
                },
            ]

            for task_data in tasks_data:
                task = Task(**task_data)
                session.add(task)

            await session.flush()

            # Test the function
            result = await fetch_current_tasks(session, test_session.id)

            assert isinstance(result, Result)
            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert len(data) == 3

            # Check if tasks are ordered by task_order
            assert data[0].task_order == 1
            assert data[1].task_order == 2
            assert data[2].task_order == 3

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_fetch_tasks_with_status_filter(self):
        """Test fetching tasks with status filter"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac2", secret_key_hash_phc="test_key_hash2"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            # Create sample tasks with different statuses
            task1 = Task(
                session_id=test_session.id,
                task_order=1,
                task_data={"name": "task1"},
                task_status="pending",
            )
            task2 = Task(
                session_id=test_session.id,
                task_order=2,
                task_data={"name": "task2"},
                task_status="running",
            )
            session.add_all([task1, task2])
            await session.flush()

            # Test filtering by status
            result = await fetch_current_tasks(
                session, test_session.id, status="pending"
            )

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert len(data) == 1
            assert data[0].task_status == "pending"

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_fetch_tasks_no_results(self):
        """Test fetching tasks for non-existent session"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            non_existent_session_id = uuid.uuid4()

            result = await fetch_current_tasks(session, non_existent_session_id)

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert len(data) == 0


class TestUpdateTask:
    @pytest.mark.asyncio
    async def test_update_task_status_success(self):
        """Test updating task status"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac3", secret_key_hash_phc="test_key_hash3"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            task = Task(
                session_id=test_session.id,
                task_order=1,
                task_data={"name": "test_task"},
                task_status="pending",
            )
            session.add(task)
            await session.flush()

            original_status = task.task_status

            result = await update_task(session, task.id, status="success")

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert data.task_status == "success"
            assert data.task_status != original_status

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_update_task_order_success(self):
        """Test updating task order"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac4", secret_key_hash_phc="test_key_hash4"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            task = Task(
                session_id=test_session.id,
                task_order=1,
                task_data={"name": "test_task"},
                task_status="pending",
            )
            session.add(task)
            await session.flush()

            original_order = task.task_order
            new_order = 10

            result = await update_task(session, task.id, order=new_order)

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert data.task_order == new_order
            assert data.task_order != original_order

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_update_task_data_success(self):
        """Test updating task data"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac5", secret_key_hash_phc="test_key_hash5"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            task = Task(
                session_id=test_session.id,
                task_order=1,
                task_data={"name": "original_task"},
                task_status="pending",
            )
            session.add(task)
            await session.flush()

            new_data = {"name": "updated_task", "priority": "high"}

            result = await update_task(session, task.id, data=new_data)

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert data.task_data == new_data

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_update_multiple_fields(self):
        """Test updating multiple task fields at once"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac6", secret_key_hash_phc="test_key_hash6"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            task = Task(
                session_id=test_session.id,
                task_order=1,
                task_data={"name": "original_task"},
                task_status="pending",
            )
            session.add(task)
            await session.flush()

            new_status = "running"
            new_order = 5
            new_data = {"name": "multi_update", "stage": "testing"}

            result = await update_task(
                session, task.id, status=new_status, order=new_order, data=new_data
            )

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert data.task_status == new_status
            assert data.task_order == new_order
            assert data.task_data == new_data

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_update_nonexistent_task(self):
        """Test updating a task that doesn't exist"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            non_existent_task_id = uuid.uuid4()

            result = await update_task(session, non_existent_task_id, status="success")

            data, error = result.unpack()
            assert data is None
            assert error is not None
            assert f"Task {non_existent_task_id} not found" in error.errmsg

    @pytest.mark.asyncio
    async def test_update_task_with_none_values(self):
        """Test updating task with None values (should not change anything)"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac7", secret_key_hash_phc="test_key_hash7"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            task = Task(
                session_id=test_session.id,
                task_order=1,
                task_data={"name": "test_task"},
                task_status="pending",
            )
            session.add(task)
            await session.flush()

            original_status = task.task_status
            original_order = task.task_order
            original_data = task.task_data

            result = await update_task(
                session, task.id, status=None, order=None, data=None
            )

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert data.task_status == original_status
            assert data.task_order == original_order
            assert data.task_data == original_data

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_update_task_patch_data_success(self):
        """Test updating task using patch_data for partial updates"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac_patch",
                secret_key_hash_phc="test_key_hash_patch",
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            # Create task with initial complex data
            initial_data = {
                "task_description": "original task",
                "priority": "medium",
                "metadata": {"tags": ["backend", "api"], "config": {"timeout": 300}},
                "status_details": {"attempts": 0},
            }
            task = Task(
                session_id=test_session.id,
                task_order=1,
                task_data=initial_data,
                task_status="pending",
            )
            session.add(task)
            await session.flush()

            # Test 1: Simple patch_data update
            patch_data = {"task_description": "updated task description"}
            result = await update_task(session, task.id, patch_data=patch_data)

            data, error = result.unpack()
            assert error is None
            assert data is not None

            # Verify the patched field is updated
            assert data.task_data["task_description"] == "updated task description"
            # Verify other fields remain unchanged
            assert data.task_data["priority"] == "medium"
            assert data.task_data["metadata"] == initial_data["metadata"]
            assert data.task_data["status_details"] == initial_data["status_details"]

            # Test 2: Add new fields via patch_data
            patch_data = {"new_field": "new_value", "assigned_to": "user123"}
            result = await update_task(session, task.id, patch_data=patch_data)

            data, error = result.unpack()
            assert error is None
            assert data is not None

            # Verify new fields are added
            assert data.task_data["new_field"] == "new_value"
            assert data.task_data["assigned_to"] == "user123"
            # Verify existing fields remain unchanged
            assert data.task_data["task_description"] == "updated task description"
            assert data.task_data["priority"] == "medium"

            # Test 3: Update multiple existing fields
            patch_data = {
                "priority": "high",
                "status_details": {"attempts": 1, "last_error": None},
            }
            result = await update_task(session, task.id, patch_data=patch_data)

            data, error = result.unpack()
            assert error is None
            assert data is not None

            # Verify updated fields
            assert data.task_data["priority"] == "high"
            assert data.task_data["status_details"] == {
                "attempts": 1,
                "last_error": None,
            }
            # Verify other fields remain unchanged
            assert data.task_data["task_description"] == "updated task description"
            assert data.task_data["new_field"] == "new_value"

            # Test 4: Verify data parameter takes precedence over patch_data
            complete_new_data = {"completely": "new", "data": "structure"}
            patch_data = {"should": "be_ignored"}

            result = await update_task(
                session, task.id, data=complete_new_data, patch_data=patch_data
            )

            data, error = result.unpack()
            assert error is None
            assert data is not None

            # Verify data parameter took precedence
            assert data.task_data == complete_new_data
            assert "should" not in data.task_data

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_update_task_patch_data_with_status_and_order(self):
        """Test updating task using patch_data combined with status and order updates"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac_patch2",
                secret_key_hash_phc="test_key_hash_patch2",
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            task = Task(
                session_id=test_session.id,
                task_order=1,
                task_data={"task_description": "original", "step": "init"},
                task_status="pending",
            )
            session.add(task)
            await session.flush()

            # Update multiple aspects of the task
            patch_data = {
                "task_description": "patched description",
                "step": "processing",
            }
            result = await update_task(
                session, task.id, status="running", order=5, patch_data=patch_data
            )

            data, error = result.unpack()
            assert error is None
            assert data is not None

            # Verify all updates were applied
            assert data.task_status == "running"
            assert data.task_order == 5
            assert data.task_data["task_description"] == "patched description"
            assert data.task_data["step"] == "processing"

            await session.delete(project)


class TestInsertTask:
    @pytest.mark.asyncio
    async def test_insert_task_success(self):
        """Test inserting a new task"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac8", secret_key_hash_phc="test_key_hash8"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            task_data = {"name": "new_task", "description": "A new task"}
            after_order = 0  # Insert after position 0 (will become position 1)

            result = await insert_task(session, test_session.id, after_order, task_data)

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert isinstance(data, Task)  # Should return Task object, not UUID
            assert data.task_order == 1  # Should be at position 1
            assert data.task_data == task_data

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_insert_task_with_custom_status(self):
        """Test inserting a task with custom status"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac9", secret_key_hash_phc="test_key_hash9"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            task_data = {"name": "custom_status_task"}
            after_order = 1  # Insert after position 1 (will become position 2)
            custom_status = "running"

            result = await insert_task(
                session, test_session.id, after_order, task_data, status=custom_status
            )

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert isinstance(data, Task)
            assert data.task_status == custom_status
            assert data.task_order == 2  # Should be at position 2
            assert data.task_data == task_data

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_insert_task_default_status(self):
        """Test inserting a task with default status"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac10", secret_key_hash_phc="test_key_hash10"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            task_data = {"name": "default_status_task"}
            after_order = 2  # Insert after position 2 (will become position 3)

            result = await insert_task(session, test_session.id, after_order, task_data)

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert isinstance(data, Task)
            assert data.task_status == "pending"  # Should have default status
            assert data.task_order == 3  # Should be at position 3

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_insert_task_complex_data(self):
        """Test inserting a task with complex JSON data"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac11", secret_key_hash_phc="test_key_hash11"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            complex_data = {
                "name": "complex_task",
                "metadata": {
                    "priority": "high",
                    "tags": ["urgent", "backend"],
                    "config": {"timeout": 300, "retries": 3},
                },
                "steps": [
                    {"action": "validate", "params": {"strict": True}},
                    {"action": "process", "params": {"batch_size": 100}},
                ],
            }

            result = await insert_task(session, test_session.id, 0, complex_data)

            data, error = result.unpack()
            assert error is None
            assert data is not None
            assert isinstance(data, Task)
            assert data.task_data == complex_data

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_insert_task_order_increment(self):
        """Test that inserting a task increments subsequent task orders"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac_order",
                secret_key_hash_phc="test_key_hash_order",
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            # Create initial tasks with orders 1, 2, 3
            task1 = Task(
                session_id=test_session.id,
                task_order=1,
                task_data={"name": "task1"},
                task_status="pending",
            )
            task2 = Task(
                session_id=test_session.id,
                task_order=2,
                task_data={"name": "task2"},
                task_status="pending",
            )
            task3 = Task(
                session_id=test_session.id,
                task_order=3,
                task_data={"name": "task3"},
                task_status="pending",
            )
            session.add_all([task1, task2, task3])
            await session.flush()

            # Insert a new task after position 1 (should become position 2)
            new_task_data = {"name": "inserted_task"}
            result = await insert_task(session, test_session.id, 1, new_task_data)

            new_task, error = result.unpack()
            assert error is None
            assert new_task.task_order == 2

            # Fetch all tasks and verify the new ordering
            fetch_result = await fetch_current_tasks(session, test_session.id)
            all_tasks, _ = fetch_result.unpack()

            assert len(all_tasks) == 4

            # Verify the new order: task1(1), inserted_task(2), task2(3), task3(4)
            assert all_tasks[0].task_data["name"] == "task1"
            assert all_tasks[0].task_order == 1

            assert all_tasks[1].task_data["name"] == "inserted_task"
            assert all_tasks[1].task_order == 2

            assert all_tasks[2].task_data["name"] == "task2"
            assert all_tasks[2].task_order == 3  # Was 2, now 3

            assert all_tasks[3].task_data["name"] == "task3"
            assert all_tasks[3].task_order == 4  # Was 3, now 4

            await session.delete(project)


class TestDeleteTask:
    @pytest.mark.asyncio
    async def test_delete_task_success(self):
        """Test deleting an existing task"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac12", secret_key_hash_phc="test_key_hash12"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            task = Task(
                session_id=test_session.id,
                task_order=1,
                task_data={"name": "task_to_delete"},
                task_status="pending",
            )
            session.add(task)
            await session.flush()

            task_id = task.id

            result = await delete_task(session, task_id)

            data, error = result.unpack()
            assert error is None
            assert data is None

            # Verify the task was actually deleted
            query = select(Task).where(Task.id == task_id)
            db_result = await session.execute(query)
            deleted_task = db_result.scalars().first()

            assert deleted_task is None

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_delete_nonexistent_task(self):
        """Test deleting a task that doesn't exist (should not raise error)"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            non_existent_task_id = uuid.uuid4()

            result = await delete_task(session, non_existent_task_id)

            data, error = result.unpack()
            assert error is None
            assert data is None

    @pytest.mark.asyncio
    async def test_delete_task_cascade_behavior(self):
        """Test that deleting a task doesn't affect other tasks"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac13", secret_key_hash_phc="test_key_hash13"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            # Create multiple tasks
            task1 = Task(
                session_id=test_session.id,
                task_order=1,
                task_data={"name": "task1"},
                task_status="pending",
            )
            task2 = Task(
                session_id=test_session.id,
                task_order=2,
                task_data={"name": "task2"},
                task_status="running",
            )
            task3 = Task(
                session_id=test_session.id,
                task_order=3,
                task_data={"name": "task3"},
                task_status="success",
            )
            session.add_all([task1, task2, task3])
            await session.flush()

            initial_count = 3
            task_to_delete = task2  # Delete middle task

            result = await delete_task(session, task_to_delete.id)

            data, error = result.unpack()
            assert error is None

            # Verify other tasks still exist
            count_query = select(func.count(Task.id)).where(
                Task.session_id == task_to_delete.session_id
            )
            count_result = await session.execute(count_query)
            remaining_count = count_result.scalar()

            assert remaining_count == initial_count - 1

            await session.delete(project)


class TestIntegrationScenarios:
    @pytest.mark.asyncio
    async def test_full_task_lifecycle(self):
        """Test complete task lifecycle: create, update, fetch, delete"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac14", secret_key_hash_phc="test_key_hash14"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            # 1. Create a task
            initial_data = {"name": "lifecycle_task", "step": "created"}
            create_result = await insert_task(session, test_session.id, 0, initial_data)
            created_task, _ = create_result.unpack()
            assert created_task is not None
            task_id = created_task.id

            # 2. Update the task
            updated_data = {"name": "lifecycle_task", "step": "updated"}
            update_result = await update_task(
                session, task_id, data=updated_data, status="running"
            )
            updated_task, _ = update_result.unpack()
            assert updated_task.task_data == updated_data
            assert updated_task.task_status == "running"

            # 3. Fetch the task
            fetch_result = await fetch_current_tasks(session, test_session.id)
            tasks, _ = fetch_result.unpack()
            assert len(tasks) == 1
            assert tasks[0].id == task_id
            assert tasks[0].task_data == updated_data

            # 4. Delete the task
            delete_result = await delete_task(session, task_id)
            _, error = delete_result.unpack()
            assert error is None

            # 5. Verify task is gone
            final_fetch_result = await fetch_current_tasks(session, test_session.id)
            final_tasks, _ = final_fetch_result.unpack()
            assert len(final_tasks) == 0

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_multiple_sessions_isolation(self):
        """Test that tasks from different sessions are properly isolated"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create two different sessions
            project = Project(
                secret_key_hmac="test_key_hmac15", secret_key_hash_phc="test_key_hash15"
            )
            session.add(project)
            await session.flush()

            space1 = Space(project_id=project.id)
            space2 = Space(project_id=project.id)
            session.add_all([space1, space2])
            await session.flush()

            session1 = Session(project_id=project.id, space_id=space1.id)
            session2 = Session(project_id=project.id, space_id=space2.id)
            session.add_all([session1, session2])
            await session.flush()

            # Create tasks in each session
            await insert_task(session, session1.id, 0, {"session": "1", "task": "A"})
            await insert_task(session, session1.id, 1, {"session": "1", "task": "B"})
            await insert_task(session, session2.id, 0, {"session": "2", "task": "A"})

            # Fetch tasks for each session
            session1_tasks_result = await fetch_current_tasks(session, session1.id)
            session2_tasks_result = await fetch_current_tasks(session, session2.id)

            session1_tasks, _ = session1_tasks_result.unpack()
            session2_tasks, _ = session2_tasks_result.unpack()

            # Verify isolation
            assert len(session1_tasks) == 2
            assert len(session2_tasks) == 1
            assert all(task.session_id == session1.id for task in session1_tasks)
            assert all(task.session_id == session2.id for task in session2_tasks)

            await session.delete(project)

    @pytest.mark.asyncio
    async def test_task_ordering_after_updates(self):
        """Test that task ordering is maintained after updates and insertions"""
        db_client = DatabaseClient()
        await db_client.create_tables()

        async with db_client.get_session_context() as session:
            # Create test data
            project = Project(
                secret_key_hmac="test_key_hmac16", secret_key_hash_phc="test_key_hash16"
            )
            session.add(project)
            await session.flush()

            space = Space(project_id=project.id)
            session.add(space)
            await session.flush()

            test_session = Session(project_id=project.id, space_id=space.id)
            session.add(test_session)
            await session.flush()

            # Create initial tasks in order
            task1_result = await insert_task(
                session,
                test_session.id,
                0,
                {"name": "task1"},  # Insert after 0 -> position 1
            )
            task2_result = await insert_task(
                session,
                test_session.id,
                1,
                {"name": "task2"},  # Insert after 1 -> position 2
            )
            task3_result = await insert_task(
                session,
                test_session.id,
                2,
                {"name": "task3"},  # Insert after 2 -> position 3
            )

            task1, _ = task1_result.unpack()
            task2, _ = task2_result.unpack()
            task3, _ = task3_result.unpack()

            # Insert a new task in the middle (after position 1)
            middle_task_result = await insert_task(
                session,
                test_session.id,
                1,
                {"name": "middle_task"},  # Insert after 1 -> position 2
            )
            middle_task, _ = middle_task_result.unpack()

            # Fetch and verify the new ordering
            fetch_result = await fetch_current_tasks(session, test_session.id)
            tasks, _ = fetch_result.unpack()

            assert len(tasks) == 4

            # Expected order: task1(1), middle_task(2), task2(3), task3(4)
            assert tasks[0].id == task1.id
            assert tasks[0].task_order == 1

            assert tasks[1].id == middle_task.id
            assert tasks[1].task_order == 2

            assert tasks[2].id == task2.id
            assert tasks[2].task_order == 3  # Was 2, incremented to 3

            assert tasks[3].id == task3.id
            assert tasks[3].task_order == 4  # Was 3, incremented to 4

            # Now test manual order updates
            await update_task(session, task1.id, order=10)  # Move task1 to the end

            # Fetch again and verify
            fetch_result2 = await fetch_current_tasks(session, test_session.id)
            tasks2, _ = fetch_result2.unpack()

            # Note: This doesn't automatically reorder other tasks - that would need
            # additional logic. For now, just verify the update worked.
            task1_updated = next(t for t in tasks2 if t.id == task1.id)
            assert task1_updated.task_order == 10

            await session.delete(project)
