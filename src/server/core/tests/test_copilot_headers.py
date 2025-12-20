import pytest


@pytest.mark.asyncio
async def test_copilot_headers_x_initiator_agent(monkeypatch):
    from acontext_core.llm.complete import openai_sdk

    async def fake_get_token():
        return "copilot_token"

    monkeypatch.setattr(openai_sdk, "get_copilot_access_token", fake_get_token)
    monkeypatch.setattr(openai_sdk.DEFAULT_CORE_CONFIG, "copilot_enabled", True)
    monkeypatch.setattr(openai_sdk.DEFAULT_CORE_CONFIG, "llm_api_key", None)

    messages = [
        {"role": "user", "content": "hello"},
        {"role": "assistant", "content": "hi"},
    ]

    headers = await openai_sdk._build_extra_headers(messages)
    assert headers is not None
    assert headers["Authorization"] == "Bearer copilot_token"
    assert headers["Openai-Intent"] == "conversation-edits"
    assert headers["X-Initiator"] == "agent"


@pytest.mark.asyncio
async def test_copilot_headers_x_initiator_user(monkeypatch):
    from acontext_core.llm.complete import openai_sdk

    async def fake_get_token():
        return "copilot_token"

    monkeypatch.setattr(openai_sdk, "get_copilot_access_token", fake_get_token)
    monkeypatch.setattr(openai_sdk.DEFAULT_CORE_CONFIG, "copilot_enabled", True)
    monkeypatch.setattr(openai_sdk.DEFAULT_CORE_CONFIG, "llm_api_key", "")

    messages = [
        {"role": "user", "content": "hello"},
        {"role": "user", "content": "second"},
    ]

    headers = await openai_sdk._build_extra_headers(messages)
    assert headers is not None
    assert headers["X-Initiator"] == "user"


@pytest.mark.asyncio
async def test_copilot_headers_vision_flag(monkeypatch):
    from acontext_core.llm.complete import openai_sdk

    async def fake_get_token():
        return "copilot_token"

    monkeypatch.setattr(openai_sdk, "get_copilot_access_token", fake_get_token)
    monkeypatch.setattr(openai_sdk.DEFAULT_CORE_CONFIG, "copilot_enabled", True)
    monkeypatch.setattr(openai_sdk.DEFAULT_CORE_CONFIG, "llm_api_key", None)

    messages = [
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "what is this?"},
                {"type": "image_url", "image_url": {"url": "https://x/y.png"}},
            ],
        }
    ]

    headers = await openai_sdk._build_extra_headers(messages)
    assert headers is not None
    assert headers["Copilot-Vision-Request"] == "true"


@pytest.mark.asyncio
async def test_copilot_headers_disabled_when_llm_api_key_present(monkeypatch):
    from acontext_core.llm.complete import openai_sdk

    async def fake_get_token():
        raise AssertionError("Should not be called")

    monkeypatch.setattr(openai_sdk, "get_copilot_access_token", fake_get_token)
    monkeypatch.setattr(openai_sdk.DEFAULT_CORE_CONFIG, "copilot_enabled", True)
    monkeypatch.setattr(openai_sdk.DEFAULT_CORE_CONFIG, "llm_api_key", "sk-test")

    headers = await openai_sdk._build_extra_headers([{"role": "user", "content": "x"}])
    assert headers is None
