from .base import SandboxRuntime
from .registry import get_runtime_for_sandbox, register_runtime_backend

# Register runtime implementations
# from ..lifecycle.enums import SandboxBackend
# from ..lifecycle.models import SandboxInfo


# Auto-register Docker runtime
# register_runtime_backend(SandboxBackend.DOCKER, _create_docker_runtime)

__all__ = [
    "SandboxRuntime",
    "get_runtime_for_sandbox",
    "register_runtime_backend",
]
