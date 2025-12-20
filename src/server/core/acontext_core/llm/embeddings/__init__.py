from typing import Literal
from traceback import format_exc
from ...env import LOG, DEFAULT_CORE_CONFIG
from ...schema.result import Result
from ...schema.embedding import EmbeddingReturn
from ...telemetry.otel import instrument_llm_embedding
from .jina_embedding import jina_embedding
from .openai_embedding import openai_embedding
from .fastembed_embedding import fastembed_embedding

FACTORIES = {
    "openai": openai_embedding,
    "jina": jina_embedding,
    "fastembed": fastembed_embedding,
}
assert DEFAULT_CORE_CONFIG.block_embedding_provider in FACTORIES, (
    f"Unsupported embedding provider: {DEFAULT_CORE_CONFIG.block_embedding_provider}"
)


async def embedding_sanity_check():
    r = await get_embedding(["Hello, world!"])
    if not r.ok():
        raise ValueError(
            "Embedding provider check failed! Verify embedding provider configuration (API key/base_url for remote providers, or local model availability for fastembed)."
        )
    d = r.data
    embedding_dim = d.embedding.shape[-1]
    if embedding_dim != DEFAULT_CORE_CONFIG.block_embedding_dim:
        raise ValueError(
            f"Embedding dimension mismatch! Expected {DEFAULT_CORE_CONFIG.block_embedding_dim}, got {embedding_dim}."
        )
    LOG.info(f"Embedding dimension matched with Config: {embedding_dim}")


@instrument_llm_embedding
async def get_embedding(
    texts: list[str],
    phase: Literal["query", "document"] = "document",
    model: str | None = None,
) -> Result[EmbeddingReturn]:
    # Prefer explicit args, otherwise use config.
    requested_model = model or DEFAULT_CORE_CONFIG.block_embedding_model
    requested_provider = DEFAULT_CORE_CONFIG.block_embedding_provider

    # Backward-compatible no-key fallback:
    # If user keeps provider="openai" but no embedding/LLM API key is configured
    # (e.g. they rely on Copilot device-flow for chat completions), embeddings would fail.
    # In that case, transparently fall back to local CPU embeddings.
    effective_provider = requested_provider
    effective_model = requested_model

    if requested_provider == "openai":
        has_openai_key = bool(
            (DEFAULT_CORE_CONFIG.block_embedding_api_key or "").strip()
            or (DEFAULT_CORE_CONFIG.llm_api_key or "").strip()
        )
        if not has_openai_key:
            effective_provider = "fastembed"
            # If the caller passed an OpenAI embedding model name (e.g. "text-embedding-3-small"),
            # FastEmbed won't recognize it. Default to a known HuggingFace model unless the
            # caller already provided a HF-style model id.
            if model is None or "/" not in str(model):
                effective_model = "BAAI/bge-small-en-v1.5"

            LOG.info(
                "No embedding API key configured for provider 'openai'; falling back to local 'fastembed' embeddings. "
                "(Set BLOCK_EMBEDDING_PROVIDER to control this behavior.)"
            )

    try:
        results = await FACTORIES[effective_provider](effective_model, texts, phase)
    except Exception as e:
        LOG.error(f"Error in get_embedding: {e} {format_exc()}")
        return Result.reject(f"Error in get_embedding: {e}")
    return Result.resolve(results)
