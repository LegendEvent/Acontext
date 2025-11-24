# Acontext Roadmap

current version: v0.0

## Integrations

We're always welcome to integrations PRs:

- If your integrations involve SDK or cli changes, pull requests in this repo.
- If your integrations are combining Acontext SDK and other frameworks, pull requests to https://github.com/memodb-io/Acontext-Examples where your templates can be downloaded through `acontext-cli`: `acontext create my-proj --template-path "LANGUAGE/YOUR-TEMPLATE"`

## v0.0

Session - Context Engineering

- Session - Tool Result reduction/artifact offloading
- Session - Count token
- Session - Context LLM compression

Chore

- Telemetry：log detailed callings and latency

## v0.1

Disk - more agentic interface

- Disk: file/dir sharing UI Component.

- Disk:  support get artifact with line number and offset
- Disk: SDK, prepare agent tools and schema to navigate and edit artifacts

Space

- Space: export use_when as system prompt

- Claude Skill integrate

  - Space: integrate Claude skill into Space

  - Sandbox with Artifacts：Simple Code Sandbox
