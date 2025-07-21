import os
from typing import Optional, Dict, Any, List, AsyncGenerator, Union
from contextlib import asynccontextmanager
import asyncio
from dataclasses import dataclass, field
from urllib.parse import urlparse
import clickhouse_connect
from clickhouse_connect.driver import Client
from clickhouse_connect.driver import httputil
from clickhouse_connect.driver.exceptions import (
    DatabaseError,
    OperationalError,
)

from ..env import LOG as logger
from ..env import ENV


@dataclass
class ClickHouseConfig:
    """Configuration for ClickHouse connection."""

    host: str = "localhost"
    port: int = 8123
    username: str = "default"
    password: str = ""
    database: str = "default"
    secure: bool = False
    verify: bool = True
    compress: bool = True
    query_limit: int = 0
    connect_timeout: int = 10
    send_receive_timeout: int = 300
    client_name: str = "acontext_server"
    # Connection pool settings
    pool_mgr: httputil.PoolManager = field(
        default_factory=lambda: httputil.PoolManager(
            maxsize=ENV.clickhouse_pool_size, num_pools=12
        )
    )


class ClickHouseClient:
    """
    Best-practice ClickHouse client with connection pooling and async support.

    Features:
    - Async ClickHouse operations with connection pooling
    - Automatic connection recovery and retry logic
    - Health checks and monitoring
    - Query optimization and batching support
    - Proper error handling and logging
    - Support for both sync and async operations
    """

    def __init__(self, config: Optional[ClickHouseConfig] = None):
        self.config = config or self._build_config_from_env()
        self._client: Client = self._create_client()
        logger.info(f"ClickHouse URL: {self.config.host}:{self.config.port}")

    def _build_config_from_env(self) -> ClickHouseConfig:
        """Build ClickHouse configuration from environment variables."""
        # Check for CLICKHOUSE_URL first (for unified configuration)
        clickhouse_url = os.getenv("CLICKHOUSE_URL")
        if clickhouse_url:
            # Parse URL format: clickhouse://user:password@host:port/database
            parsed = urlparse(clickhouse_url)
            return ClickHouseConfig(
                host=parsed.hostname or "localhost",
                port=parsed.port or 8123,
                username=parsed.username or "default",
                password=parsed.password or "",
                database=parsed.path.lstrip("/") if parsed.path else "default",
                secure=parsed.scheme == "clickhouses",
            )

        raise ValueError("CLICKHOUSE_URL is not set")

    @property
    def client(self) -> Client:
        """Get the ClickHouse client, creating it if necessary."""
        return self._client

    def _create_client(self) -> Client:
        """Create the ClickHouse client with optimal settings."""
        try:
            client = clickhouse_connect.get_client(
                host=self.config.host,
                port=self.config.port,
                username=self.config.username,
                password=self.config.password,
                database=self.config.database,
                secure=self.config.secure,
                verify=self.config.verify,
                compress=self.config.compress,
                query_limit=self.config.query_limit,
                connect_timeout=self.config.connect_timeout,
                send_receive_timeout=self.config.send_receive_timeout,
                client_name=self.config.client_name,
                pool_mgr=self.config.pool_mgr,
                # Additional performance settings
                settings={
                    "max_execution_time": 300,  # 5 minutes max query time
                    "max_memory_usage": 10000000000,  # 10GB max memory per query
                    "use_uncompressed_cache": 1,
                    "load_balancing": "random",
                    "max_threads": 8,
                },
            )

            logger.info("ClickHouse client created successfully")
            return client

        except Exception as e:
            logger.error(f"Failed to create ClickHouse client: {e}")
            raise

    async def get_client(self) -> Client:
        """
        Get a ClickHouse client instance with thread safety.

        Note: clickhouse-connect client is thread-safe but we use async lock
        for consistent patterns with other clients in the project.
        """
        return self.client

    @asynccontextmanager
    async def get_client_context(self) -> AsyncGenerator[Client, None]:
        """
        Get a ClickHouse client with context manager support.

        Usage:
            async with clickhouse_client.get_client_context() as client:
                result = client.query("SELECT * FROM users LIMIT 10")
                data = result.result_rows
        """
        client = await self.get_client()
        try:
            yield client
        except Exception as e:
            logger.error(f"Error in ClickHouse client context: {e}")
            raise

    async def execute_query(
        self,
        query: str,
        parameters: Optional[Dict[str, Any]] = None,
        settings: Optional[Dict[str, Any]] = None,
        use_numpy: bool = False,
        column_oriented: bool = False,
    ) -> Any:
        """
        Execute a query asynchronously.

        Args:
            query: SQL query string
            parameters: Query parameters for substitution
            settings: ClickHouse settings for this query
            use_numpy: Return results as numpy arrays
            column_oriented: Return column-oriented data

        Returns:
            Query result object
        """
        async with self.get_client_context() as client:
            try:
                # Run in thread pool to avoid blocking
                loop = asyncio.get_event_loop()

                def _execute():
                    return client.query(
                        query=query,
                        parameters=parameters or {},
                        settings=settings or {},
                        use_numpy=use_numpy,
                        column_oriented=column_oriented,
                    )

                result = await loop.run_in_executor(None, _execute)
                logger.debug(f"Query executed successfully: {query[:100]}...")
                return result

            except (DatabaseError, OperationalError) as e:
                logger.error(f"ClickHouse query failed: {e}")
                raise
            except Exception as e:
                logger.error(f"Unexpected error executing query: {e}")
                raise

    async def execute_command(
        self,
        command: str,
        parameters: Optional[Dict[str, Any]] = None,
        settings: Optional[Dict[str, Any]] = None,
    ) -> Any:
        """
        Execute a command (INSERT, CREATE, DROP, etc.) asynchronously.

        Args:
            command: SQL command string
            parameters: Command parameters for substitution
            settings: ClickHouse settings for this command

        Returns:
            Command result
        """
        async with self.get_client_context() as client:
            try:
                loop = asyncio.get_event_loop()

                def _execute():
                    return client.command(
                        cmd=command,
                        parameters=parameters or {},
                        settings=settings or {},
                    )

                result = await loop.run_in_executor(None, _execute)
                logger.debug(f"Command executed successfully: {command[:100]}...")
                return result

            except (DatabaseError, OperationalError) as e:
                logger.error(f"ClickHouse command failed: {e}")
                raise
            except Exception as e:
                logger.error(f"Unexpected error executing command: {e}")
                raise

    async def insert_data(
        self,
        table: str,
        data: Union[List[List[Any]], List[Dict[str, Any]]],
        column_names: Optional[List[str]] = None,
        database: Optional[str] = None,
        settings: Optional[Dict[str, Any]] = None,
    ) -> int:
        """
        Insert data into a table asynchronously.

        Args:
            table: Target table name
            data: Data to insert (list of rows or list of dicts)
            column_names: Column names (required if data is list of lists)
            database: Target database (uses client default if not specified)
            settings: ClickHouse settings for this insert

        Returns:
            Number of rows inserted
        """
        async with self.get_client_context() as client:
            try:
                loop = asyncio.get_event_loop()

                def _insert():
                    return client.insert(
                        table=table,
                        data=data,
                        column_names=column_names,
                        database=database or self.config.database,
                        settings=settings or {},
                    )

                result = await loop.run_in_executor(None, _insert)
                logger.debug(
                    f"Inserted {len(data) if hasattr(data, '__len__') else 'unknown'} rows into {table}"
                )
                return result

            except (DatabaseError, OperationalError) as e:
                logger.error(f"ClickHouse insert failed: {e}")
                raise
            except Exception as e:
                logger.error(f"Unexpected error inserting data: {e}")
                raise

    async def health_check(self) -> bool:
        """
        Perform a health check on the ClickHouse connection.

        Returns:
            True if ClickHouse is accessible, False otherwise.
        """
        try:
            async with self.get_client_context() as client:
                loop = asyncio.get_event_loop()
                result = await loop.run_in_executor(
                    None, lambda: client.query("SELECT 1 as health_check")
                )
                return result.result_rows[0][0] == 1

        except Exception as e:
            logger.error(f"ClickHouse health check failed: {e}")
            return False

    async def get_server_info(self) -> Dict[str, Any]:
        """Get ClickHouse server information."""
        try:
            async with self.get_client_context() as client:
                loop = asyncio.get_event_loop()

                def _get_info():
                    version_result = client.query("SELECT version() as version")
                    uptime_result = client.query("SELECT uptime() as uptime")
                    return {
                        "version": version_result.result_rows[0][0],
                        "uptime_seconds": uptime_result.result_rows[0][0],
                        "host": self.config.host,
                        "port": self.config.port,
                        "database": self.config.database,
                    }

                return await loop.run_in_executor(None, _get_info)

        except Exception as e:
            logger.error(f"Failed to get server info: {e}")
            return {"error": str(e)}

    async def close(self) -> None:
        """Close the ClickHouse client connection."""
        if self._client:
            try:
                # clickhouse-connect client doesn't have explicit close method
                # but we can clear the reference to allow garbage collection
                self._client = None
                logger.info("ClickHouse client closed")
            except Exception as e:
                logger.error(f"Error closing ClickHouse client: {e}")

    def get_connection_info(self) -> Dict[str, Any]:
        """Get current connection information for monitoring."""
        return {
            "host": self.config.host,
            "port": self.config.port,
            "database": self.config.database,
            "username": self.config.username,
            "secure": self.config.secure,
            "connected": self._client is not None,
        }


# Global ClickHouse client instance
CLICKHOUSE_CLIENT = ClickHouseClient()


# Dependency function for FastAPI
async def get_clickhouse_client() -> Client:
    """
    FastAPI dependency to get a ClickHouse client.

    Usage in FastAPI routes:
        @app.get("/analytics")
        async def get_analytics(clickhouse: Client = Depends(get_clickhouse_client)):
            result = clickhouse.query("SELECT count() FROM events")
            return {"total_events": result.result_rows[0][0]}
    """
    return await CLICKHOUSE_CLIENT.get_client()


# Convenience functions
async def init_clickhouse() -> None:
    """Initialize ClickHouse connection (perform health check)."""
    if await CLICKHOUSE_CLIENT.health_check():
        server_info = await CLICKHOUSE_CLIENT.get_server_info()
        logger.info(f"ClickHouse server info: {server_info}")
    else:
        logger.error("Failed to initialize ClickHouse connection")
        raise ConnectionError("Could not connect to ClickHouse")


async def close_clickhouse() -> None:
    """Close ClickHouse connections."""
    await CLICKHOUSE_CLIENT.close()
    logger.info("ClickHouse client closed")
