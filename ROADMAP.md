# Acontext Roadmap

current version: v0.0

## Integrations

We're always welcome to integrations PRs:

- If your integrations involve SDK or cli changes, pull requests in this repo.
- If your integrations are combining Acontext SDK and other frameworks, pull requests to https://github.com/memodb-io/Acontext-Examples where your templates can be downloaded through `acontext-cli`: `acontext create my-proj --template-path "LANGUAGE/YOUR-TEMPLATE"`

## v0.0

Chore

- Telemetry：log detailed callings and latency

Prompt

- Prune prompts to lower cost, reduce thinking output
- Optimize task agent prompt to better reserve conditions of tasks
- Optimize experience agent prompt to act 

Text Match

- Use `pg_trim` to support `grep` and `glob` in Disks
- Use `pg_trim` to support keyword-matching `grep` and `glob` in Spaces

Session - Context Engineering

- Session - Count tokens
- Session - message version control
- Session - Context Compression based on Tasks
- Session - Context Offloading based on Disks

Dashboard

- Optimize task viewer to show description, progress and preferences

Core

- Fix bugs for long-handing MQ disconnection.

## v0.1

Disk - more agentic interface

- Disk: file/dir sharing UI Component.

- Disk:  support get artifact with line number and offset
- Disk: SDK, prepare agent tools and schema to navigate and edit artifacts

Space

- Space: export use_when as system prompt
- Integrate Claude Skill 

  - Space: integrate Claude skill into Space

  - Sandbox with Artifacts：Simple Code Sandbox
