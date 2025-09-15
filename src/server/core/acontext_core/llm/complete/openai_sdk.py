import json
from typing import Optional
from .clients import get_openai_async_client_instance
from openai.types.chat import ChatCompletion
from ...env import LOG, CONFIG
from ...schema.llm import LLMResponse


async def openai_complete(
    prompt,
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

    messages = []
    if system_prompt:
        messages.append({"role": "system", "content": system_prompt})
    messages.extend(history_messages)
    messages.append({"role": "user", "content": prompt})

    response: ChatCompletion = await openai_async_client.chat.completions.create(
        model=model,
        messages=messages,
        timeout=CONFIG.llm_response_timeout,
        max_tokens=max_tokens,
        tools=tools,
        **kwargs,
    )
    cached_tokens = getattr(response.usage.prompt_tokens_details, "cached_tokens", None)
    LOG.info(
        f"LLM Complete: {prompt_id} {model}. "
        f"cached {cached_tokens}, input {response.usage.prompt_tokens}, total {response.usage.total_tokens}"
    )

    _fc = (
        response.choices[0].message.function_call.model_dump()
        if response.choices[0].message.function_call
        else None
    )
    _tu = (
        [tool.model_dump() for tool in response.choices[0].message.tool_calls]
        if response.choices[0].message.tool_calls
        else None
    )
    llm_response = LLMResponse(
        role=response.choices[0].message.role,
        content=response.choices[0].message.content,
        function_call=_fc,
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
