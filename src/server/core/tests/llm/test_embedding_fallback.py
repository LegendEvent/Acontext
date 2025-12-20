import numpy as np
import pytest

from acontext_core.env import DEFAULT_CORE_CONFIG
from acontext_core.llm import embeddings
from acontext_core.schema.embedding import EmbeddingReturn


@pytest.mark.asyncio
async def test_get_embedding_falls_back_to_fastembed_when_no_key(monkeypatch):
    monkeypatch.setattr(DEFAULT_CORE_CONFIG, "block_embedding_provider", "openai")
    monkeypatch.setattr(DEFAULT_CORE_CONFIG, "block_embedding_api_key", None)
    monkeypatch.setattr(DEFAULT_CORE_CONFIG, "llm_api_key", None)
    monkeypatch.setattr(
        DEFAULT_CORE_CONFIG, "block_embedding_model", "text-embedding-3-small"
    )
    monkeypatch.setattr(DEFAULT_CORE_CONFIG, "block_embedding_dim", 1536)

    calls: list[tuple[str, list[str], str]] = []

    async def _fake_fastembed(model: str, texts: list[str], phase: str = "document"):
        calls.append((model, texts, phase))
        return EmbeddingReturn(
            embedding=np.zeros(
                (len(texts), DEFAULT_CORE_CONFIG.block_embedding_dim), dtype=np.float32
            ),
            prompt_tokens=0,
            total_tokens=0,
        )

    monkeypatch.setitem(embeddings.FACTORIES, "fastembed", _fake_fastembed)

    r = await embeddings.get_embedding(["hello"], phase="query")
    assert r.ok()

    assert len(calls) == 1
    model, texts, phase = calls[0]
    assert model == "BAAI/bge-small-en-v1.5"
    assert texts == ["hello"]
    assert phase == "query"


@pytest.mark.asyncio
async def test_get_embedding_uses_fastembed_model_if_explicit_hf_model(monkeypatch):
    monkeypatch.setattr(DEFAULT_CORE_CONFIG, "block_embedding_provider", "openai")
    monkeypatch.setattr(DEFAULT_CORE_CONFIG, "block_embedding_api_key", None)
    monkeypatch.setattr(DEFAULT_CORE_CONFIG, "llm_api_key", None)
    monkeypatch.setattr(DEFAULT_CORE_CONFIG, "block_embedding_dim", 1536)

    calls: list[str] = []

    async def _fake_fastembed(model: str, texts: list[str], phase: str = "document"):
        calls.append(model)
        return EmbeddingReturn(
            embedding=np.zeros(
                (len(texts), DEFAULT_CORE_CONFIG.block_embedding_dim), dtype=np.float32
            ),
            prompt_tokens=0,
            total_tokens=0,
        )

    monkeypatch.setitem(embeddings.FACTORIES, "fastembed", _fake_fastembed)

    r = await embeddings.get_embedding(["hello"], model="BAAI/bge-small-en-v1.5")
    assert r.ok()
    assert calls == ["BAAI/bge-small-en-v1.5"]
