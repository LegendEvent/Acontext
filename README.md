<div align="center">
  <a href="https://discord.acontext.io">
      <img alt="Show Acontext header banner" src="./assets/Acontext-header-banner.png">
  </a>
  <p>
    <h3>Engineer Contexts, Learn Skills</h3>
  </p>
  <p align="center">
    <a href="https://pypi.org/project/acontext/"><img src="https://img.shields.io/pypi/v/acontext.svg"></a>
    <a href="https://www.npmjs.com/package/@acontext/acontext"><img src="https://img.shields.io/npm/v/@acontext/acontext.svg?logo=npm&logoColor=fff&style=flat&labelColor=2C2C2C&color=28CF8D"></a>
    <a href="https://github.com/memodb-io/acontext/actions/workflows/core-test.yaml"><img src="https://github.com/memodb-io/acontext/actions/workflows/core-test.yaml/badge.svg"></a>
    <a href="https://github.com/memodb-io/acontext/actions/workflows/api-test.yaml"><img src="https://github.com/memodb-io/acontext/actions/workflows/api-test.yaml/badge.svg"></a>
    <a href="https://github.com/memodb-io/acontext/actions/workflows/cli-test.yaml"><img src="https://github.com/memodb-io/acontext/actions/workflows/cli-test.yaml/badge.svg"></a>
  </p>
  <p align="center">
    <a href="https://x.com/acontext_io"><img src="https://img.shields.io/twitter/follow/acontext_io?style=social" alt="Twitter Follow"></a>
    <a href="https://discord.acontext.io"><img src="https://img.shields.io/badge/dynamic/json?label=Acontext&style=flat&query=approximate_member_count&url=https%3A%2F%2Fdiscord.com%2Fapi%2Fv10%2Finvites%2FSG9xJcqVBu%3Fwith_counts%3Dtrue&logo=discord&logoColor=white&suffix=+members&color=36393f&labelColor=5765F2" alt="Acontext Discord"></a>
  </p>
  <div align="center">
    <!-- Keep these links. Translations will automatically update with the README. -->
    <a href="./readme/de/README.md">Deutsch</a> | 
    <a href="./readme/es/README.md">Español</a> | 
    <a href="./readme/fr/README.md">Français</a> | 
    <a href="./readme/ja/README.md">日本語</a> | 
    <a href="./readme/ko/README.md">한국어</a> | 
    <a href="./readme/pt/README.md">Português</a> | 
    <a href="./readme/ru/README.md">Русский</a> | 
    <a href="./readme/zh/README.md">中文</a>
  </div>
  <br/>
</div>

Acontext is a context data platform for building cloud-native AI agents.

It helps you:
- Store sessions (messages) and artifacts
- Observe agent tasks and user feedback
- Distill reusable skills (SOPs) from completed work
- Explore everything in a web dashboard

## Core features

- [Session](https://docs.acontext.io/store/messages/multi-provider) — multi-modal message storage
  - [Task Agent](https://docs.acontext.io/observe/agent_tasks) — background TODO agent extracting tasks, progress, and preferences
  - [Context Editing](https://docs.acontext.io/store/editing) — edit/rewrite context in one call
- [Disk](https://docs.acontext.io/store/disk) — artifact filesystem (upload, download, list)
- [Space](https://docs.acontext.io/learn/skill-space) — Notion-like knowledge base for agents
  - [Experience Agent](https://docs.acontext.io/learn/advance/experience-agent) — distills skills from sessions and saves them to Spaces
- [Dashboard](https://docs.acontext.io/observe/dashboard) — view messages, artifacts, skills, and metrics

## Local quickstart

1) Install the CLI:

```bash
curl -fsSL https://install.acontext.io | sh
```

2) Start the backend locally (requires Docker with docker compose):

```bash
mkdir acontext_server && cd acontext_server
acontext docker up
```

3) Open:
- API base URL: http://localhost:8029/api/v1
- Dashboard: http://localhost:3050/

### Local LLM & embeddings (no vendor API keys)

This repo supports running Core without vendor API keys:
- LLM (chat/completions): falls back to GitHub Copilot via GitHub OAuth device flow when `LLM_API_KEY` is missing
- Embeddings: falls back to local CPU embeddings via `fastembed` when embedding keys are missing

If you provide an LLM API key (OpenAI/Anthropic-compatible), set `LLM_SIMPLE_MODEL` to your preferred model (default: `gpt-4.1`).

## Docs and SDKs

- Docs: https://docs.acontext.io/
- CLI: ./src/client/acontext-cli/README.md
- Python SDK: ./src/client/acontext-py/README.md
- TypeScript SDK: ./src/client/acontext-ts/README.md

## What’s next (v0.1 roadmap highlights)

From the roadmap (see [ROADMAP.md](./ROADMAP.md)):
- Message version control
- Session context offloading via Disks
- Session message labeling (like/dislike/feedback)
- Session metadata (JSONB) with query/filter support across API/SDK
- Observability dashboards (telemetry metrics, service chain traces, internal service health)
- Sandbox integration and resource monitoring
- Security & privacy: encrypt context data in S3 with project API keys

## Contributing

- Check [ROADMAP.md](./ROADMAP.md) first
- Read [CONTRIBUTING.md](./CONTRIBUTING.md)

## License

Apache 2.0 — see [LICENSE](./LICENSE).

## Badges

![Made with Acontext](./assets/badge-made-with-acontext.svg) ![Made with Acontext (dark)](./assets/badge-made-with-acontext-dark.svg)

```md
[![Made with Acontext](https://assets.memodb.io/Acontext/badge-made-with-acontext.svg)](https://acontext.io)

[![Made with Acontext](https://assets.memodb.io/Acontext/badge-made-with-acontext-dark.svg)](https://acontext.io)
```