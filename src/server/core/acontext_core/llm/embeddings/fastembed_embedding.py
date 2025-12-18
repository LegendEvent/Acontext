from __future__ import annotations

import asyncio
import time
from typing import Literal

import numpy as np

from ...env import DEFAULT_CORE_CONFIG, LOG
from ...schema.embedding import EmbeddingReturn

_model_cache: dict[tuple[str, str | None], object] = {}
_model_lock = asyncio.Lock()


async def _get_text_embedding_model(model_name: str):
    """Lazily construct and cache a FastEmbed TextEmbedding model.

    FastEmbed downloads models on first use and caches them locally.
    """

    cache_dir = DEFAULT_CORE_CONFIG.block_embedding_fastembed_cache_dir
    cache_key = (model_name, cache_dir)

    async with _model_lock:
        cached = _model_cache.get(cache_key)
        if cached is not None:
            return cached

        try:
            from fastembed import TextEmbedding  # local import; keeps import cheap
        except Exception as e:  # pragma: no cover
            raise RuntimeError(
                "fastembed is not installed. Add 'fastembed' to dependencies to use local embeddings."
            ) from e

        def _construct():
            if cache_dir:
                return TextEmbedding(model_name=model_name, cache_dir=cache_dir)
            return TextEmbedding(model_name=model_name)

        model = await asyncio.to_thread(_construct)
        _model_cache[cache_key] = model
        return model


def _maybe_add_prefix(
    texts: list[str], phase: Literal["query", "document"]
) -> list[str]:
    if not DEFAULT_CORE_CONFIG.block_embedding_fastembed_add_prefix:
        return texts

    prefix = "query: " if phase == "query" else "passage: "
    return [f"{prefix}{t}" for t in texts]


async def fastembed_embedding(
    model: str, texts: list[str], phase: Literal["query", "document"] = "document"
) -> EmbeddingReturn:
    """Local CPU embeddings via FastEmbed (ONNX Runtime).

    NOTES:
    - This is a synchronous model; we run it in a thread to avoid blocking the event loop.
    - Token usage is not available locally; we report 0 for prompt/total tokens.
    - Ensure `block_embedding_dim` matches the chosen FastEmbed model dimension
      (e.g. BAAI/bge-small-en-v1.5 -> 384).
    """

    embedding_model = await _get_text_embedding_model(model)
    input_texts = _maybe_add_prefix(texts, phase)

    _start_s = time.perf_counter()

    def _embed_sync() -> np.ndarray:
        # fastembed returns an iterable of numpy arrays
        vectors = list(getattr(embedding_model, "embed")(input_texts))
        if not vectors:
            return np.empty(
                (0, DEFAULT_CORE_CONFIG.block_embedding_dim), dtype=np.float32
            )

        raw = np.stack([np.asarray(v, dtype=np.float32) for v in vectors])

        # Ensure output matches configured dimension to keep DB schema stable.
        target_dim = int(DEFAULT_CORE_CONFIG.block_embedding_dim)
        cur_dim = int(raw.shape[-1])
        if cur_dim == target_dim:
            return raw

        if cur_dim > target_dim:
            return raw[:, :target_dim]

        # pad with zeros
        pad_width = target_dim - cur_dim
        pad = np.zeros((raw.shape[0], pad_width), dtype=np.float32)
        return np.concatenate([raw, pad], axis=1)

    embedding = await asyncio.to_thread(_embed_sync)

    _end_s = time.perf_counter()
    LOG.info(
        f"FastEmbed embedding, {model}, {phase}, batch={len(texts)}, time {_end_s - _start_s:.4f}s"
    )

    return EmbeddingReturn(
        embedding=embedding,
        prompt_tokens=0,
        total_tokens=0,
    )
