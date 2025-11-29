# Acontext Roadmap

current version: v0.0

## Integrations

We're always welcome to integrations PRs:

- If your integrations involve SDK or cli changes, pull requests in this repo.
- If your integrations are combining Acontext SDK and other frameworks, pull requests to https://github.com/memodb-io/Acontext-Examples where your templates can be downloaded through `acontext-cli`: `acontext create my-proj --template-path "LANGUAGE/YOUR-TEMPLATE"`

## v0.0

Prompt

- [ ] Prune prompts to lower cost, reduce thinking output
- [x] Optimize task agent prompt to better reserve conditions of tasks
- [ ] Optimize experience agent prompt to act 

Session - Context Engineering

- [x] Session - Count tokens

Dashboard

- [x] Add task viewer to show description, progress and preferences

Core

- [ ] Fix bugs for long-handing MQ disconnection.

SDK: Design `agent` interface: `tool_pool`

- [ ] Offer tool_schema for openai/anthropic can directly operate artifacts

Chore

- [ ] Telemetry：log detailed callings and latency

## v0.1

Disk - more agentic interface

- [ ] Disk: file/dir sharing UI Component.

- [ ] Disk:  support get artifact with line number and offset
- [ ] Disk: SDK, prepare agent tools and schema to navigate and edit artifacts

Space

- [ ] Space: export use_when as system prompt
- [ ] Text Match

  - Use `pg_trim` to support `grep` and `glob` in Disks
  - Use `pg_trim` to support keyword-matching `grep` and `glob` in Spaces

Session - Context Engineering

- [ ] Session - message version control
- [ ] Session - Context Offloading based on Disks

Sandbox

- [ ] Add sandbox resource in Acontext
- [ ] Integrate Claude Skill 

Sercurity&Privacy

- [ ] Use project api key to encrypt context data in S3
