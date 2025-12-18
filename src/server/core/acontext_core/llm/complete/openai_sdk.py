import json
from time import perf_counter
from typing import Any, Optional

from openai.types.chat import ChatCompletion
from openai.types.chat import ChatCompletionMessageToolCall

from .clients import get_openai_async_client_instance
from ..copilot_auth import COPILOT_DEFAULT_HEADERS, get_copilot_access_token
from ...env import DEFAULT_CORE_CONFIG, LOG
from ...schema.llm import LLMResponse


def convert_openai_tool_to_llm_tool(tool_body: ChatCompletionMessageToolCall) -> dict:
    return {
        "id": tool_body.id,
        "type": tool_body.type,
        "function": {
            "name": tool_body.function.name,
            "arguments": json.loads(tool_body.function.arguments),
        },
    }


def _is_agent_call(messages: list[dict]) -> bool:
    # Must match opencode-copilot-auth: agent if any message role is tool/assistant
    for msg in messages:
        role = (msg or {}).get("role")
        if role in ("tool", "assistant"):
            return True
    return False


def _is_vision_request(messages: list[dict]) -> bool:
    # Must match opencode-copilot-auth: any message has content list with part.type == image_url
    for msg in messages:
        content = (msg or {}).get("content")
        if isinstance(content, list):
            for part in content:
                if isinstance(part, dict) and part.get("type") == "image_url":
                    return True
    return False


async def _build_extra_headers(messages: list[dict]) -> Optional[dict[str, str]]:
    if not (
        DEFAULT_CORE_CONFIG.copilot_enabled
        and not (DEFAULT_CORE_CONFIG.llm_api_key or "").strip()
    ):
        return None

    token = await get_copilot_access_token()

    headers: dict[str, str] = {
        **COPILOT_DEFAULT_HEADERS,
        "Authorization": f"Bearer {token}",
        "Openai-Intent": "conversation-edits",
        "X-Initiator": "agent" if _is_agent_call(messages) else "user",
    }
    if _is_vision_request(messages):
        headers["Copilot-Vision-Request"] = "true"

    # mirror plugin behavior: delete these if present in request headers.
    # In Python we only control extra_headers, but ensure we never add them.
    headers.pop("x-api-key", None)
    headers.pop("authorization", None)

    return headers


async def openai_complete(
    prompt=None,
    model=None,
    system_prompt=None,
    history_messages=[],
    json_mode=False,
    max_tokens=1024,
    prompt_kwargs: Optional[dict] = None,
    tools=None,
    **kwargs,
) -> LLMResponse:
    prompt_kwargs = prompt_kwargs or {}
    prompt_id = prompt_kwargs.get("prompt_id", "...")

    openai_async_client = get_openai_async_client_instance()

    if json_mode:
        kwargs["response_format"] = {"type": "json_object"}

    messages: list[dict[str, Any]] = []
    if system_prompt:
        messages.append({"role": "system", "content": system_prompt})
    messages.extend(history_messages)
    if prompt:
        messages.append({"role": "user", "content": prompt})

    if not messages:
        raise ValueError("No messages provided")

    extra_headers = await _build_extra_headers(messages)

    _start_s = perf_counter()
    response: ChatCompletion = await openai_async_client.chat.completions.create(
        model=model,
        messages=messages,
        timeout=DEFAULT_CORE_CONFIG.llm_response_timeout,
        max_tokens=max_tokens,
        tools=tools,
        extra_headers=extra_headers,
        **DEFAULT_CORE_CONFIG.llm_openai_completion_kwargs,
        **kwargs,
    )
    _end_s = perf_counter()
    cached_tokens = getattr(response.usage.prompt_tokens_details, "cached_tokens", None)
    LOG.info(
        f"LLM Complete: {prompt_id} {model}. "
        f"cached {cached_tokens}, input {response.usage.prompt_tokens}, total {response.usage.total_tokens}, "
        f"time {_end_s - _start_s:.4f}s"
    )

    _tu = (
        [
            convert_openai_tool_to_llm_tool(tool)
            for tool in response.choices[0].message.tool_calls
        ]
        if response.choices[0].message.tool_calls
        else None
    )

    llm_response = LLMResponse(
        role=response.choices[0].message.role,
        raw_response=response,
        content=response.choices[0].message.content,
        tool_calls=_tu,
    )

    if json_mode:
        try:
            json_content = json.loads(response.choices[0].message.content)
        except json.JSONDecodeError:
            LOG.error(f"JSON decode error: {response.choices[0].message.content}")
            json_content = None
        llm_response.json_content = json_content

    return llm_response
