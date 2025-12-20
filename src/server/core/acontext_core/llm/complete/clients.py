from openai import AsyncOpenAI
from anthropic import AsyncAnthropic

from ...env import DEFAULT_CORE_CONFIG

_global_openai_async_client: AsyncOpenAI | None = None
_global_anthropic_async_client = None


def get_openai_async_client_instance() -> AsyncOpenAI:
    global _global_openai_async_client
    if _global_openai_async_client is None:
        # If llm_api_key is missing/empty, we fall back to GitHub Copilot (OpenAI-compatible)
        use_api_key = DEFAULT_CORE_CONFIG.llm_api_key or ""
        use_base_url = DEFAULT_CORE_CONFIG.llm_base_url

        if DEFAULT_CORE_CONFIG.copilot_enabled and not use_api_key.strip():
            from ..copilot_auth import copilot_openai_base_url, COPILOT_DEFAULT_HEADERS

            use_api_key = ""  # Authorization is injected per-request via extra_headers
            use_base_url = copilot_openai_base_url()

            # Merge any user-provided headers with Copilot-required headers.
            extra_default_headers = {
                **(DEFAULT_CORE_CONFIG.llm_openai_default_header or {}),
                **COPILOT_DEFAULT_HEADERS,
            }
        else:
            extra_default_headers = DEFAULT_CORE_CONFIG.llm_openai_default_header

        _global_openai_async_client = AsyncOpenAI(
            base_url=use_base_url,
            api_key=use_api_key,
            default_query=DEFAULT_CORE_CONFIG.llm_openai_default_query,
            default_headers=extra_default_headers,
        )
    return _global_openai_async_client


def get_anthropic_async_client_instance() -> AsyncAnthropic:
    global _global_anthropic_async_client
    if _global_anthropic_async_client is None:
        _global_anthropic_async_client = AsyncAnthropic(
            api_key=DEFAULT_CORE_CONFIG.llm_api_key,
            base_url=DEFAULT_CORE_CONFIG.llm_base_url,
        )
    return _global_anthropic_async_client
