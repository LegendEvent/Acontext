# Overview

OpenCode Acontext Plugin connects the OpenCode runtime to an Acontext backend to:

- persist agent interactions (user, assistant, tool calls/results) into Acontext sessions
- retrieve relevant “skill blocks” from Acontext via `experience_search`
- inject those skills into OpenCode’s *system prompt* (not into the user message)

This README is intentionally written to be readable by both humans and LLM agents.

## What it does (feature-level)

### 1) Space & session lifecycle (automatic)

- Creates/uses **one Acontext Space per worktree**.
  - Space name: `opencode:<repoName>`
  - Worktree keying: `sha256(worktree).slice(0, 16)`
- Creates/uses **one Acontext Session per OpenCode session**.
- Caches mappings locally to avoid re-creating resources every run.

### 2) Stores runtime data to Acontext

The plugin stores messages into the mapped Acontext session:

- **User messages** from OpenCode’s `chat.message` hook
- **Tool calls** from `tool.execute.before` (tool name + JSON arguments)
- **Tool results** from `tool.execute.after`
- **Assistant output** captured from OpenCode events:
  - buffers streaming text from `message.part.updated`
  - flushes once the assistant message completes (`message.updated` with `time.completed`) or when OpenCode becomes idle (`session.idle`)

Stored message format:
- Acontext API call: `POST /session/{sessionId}/messages`
- Plugin stores with `format: "acontext"` and a `blob` containing `{ role, parts, meta }`.

### 3) Skill search + system prompt injection

On each non-background user message, the plugin:

1. Calls Acontext `experience_search` for the current Space:
   - `GET /space/{spaceId}/experience_search?query=...&mode=...&limit=...&max_iterations=...`
2. Formats Acontext “cited blocks” into a compact skill reference text.
3. Caches the produced skills text per `(spaceId, sessionID)` for a short TTL.
4. Injects the skills into the OpenCode system prompt via:
   - hook: `experimental.chat.system.transform`
   - marker: `<!-- opencode-acontext:v1 -->`
   - header: configurable via `injectHeader`

Important note:
- The plugin does **not** mutate `output.parts` for the user message.
- Injection happens by appending a system block to `output.system`.

### Background task messages are not injected

If the user text starts with:

- `[BACKGROUND TASK COMPLETED]`

…the plugin stores the message as-is and skips experience search and injection. This prevents feedback loops from OpenCode agent status updates.

# Usage

## Install (OpenCode plugin file)

OpenCode loads plugins from:

- `~/.config/opencode/plugin/`

Copy the plugin file:

```bash
cp src/client/acontext-opencode-plugin/plugin/50-acontext.ts \
  ~/.config/opencode/plugin/50-acontext.ts
```

## Verify it’s working (required procedure)

After copying the plugin, run:

```bash
timeout 30s bash -lc 'echo "test" | opencode run'
```

If you have debug enabled (see below), you should see HTTP request/response logs and stored messages being sent to Acontext.

## Configuration

Configuration is loaded in this order (later overrides earlier):

1. Plugin defaults (`DEFAULT_CONFIG` in `50-acontext.ts`)
2. User config: `~/.config/opencode/acontext.json`
3. Project config: `<repo>/.opencode/acontext.json`
4. Environment variables

### `acontext.json` fields

```json
{
  "baseUrl": "http://127.0.0.1:8029/api/v1",
  "apiKey": "sk-ac-your-root-api-bearer-token",

  "searchBaseUrl": "http://127.0.0.1:8029/api/v1",
  "searchApiKey": "sk-ac-your-root-api-bearer-token",

  "mode": "fast",
  "limit": 5,
  "maxIterations": 4,
  "maxDistance": 0.8,

  "injectHeader": "SKILLS REFERENCES:"
}
```

Notes:
- `apiKey` / `searchApiKey`: can be either `Bearer <token>` or just `<token>`; the plugin normalizes to `Bearer <token>`.
- If `searchBaseUrl` / `searchApiKey` aren’t provided, search uses `baseUrl` / `apiKey`.

### Environment variables

- `ACONTEXT_BASE_URL`
- `ACONTEXT_API_KEY`
- `ACONTEXT_SEARCH_BASE_URL`
- `ACONTEXT_SEARCH_API_KEY`
- `ACONTEXT_MODE` (`fast` or `agentic`)
- `ACONTEXT_LIMIT`
- `ACONTEXT_MAX_ITERATIONS`
- `ACONTEXT_MAX_DISTANCE`

## State files

The plugin stores IDs in:

- `~/.config/opencode/acontext-state.json`

It contains two maps:

- `spaces`: worktree-hash → `spaceId`
- `sessions`: opencodeSessionId → `acontextSessionId`

## Debug logging

Enable debug logs:

- `ACONTEXT_PLUGIN_DEBUG=1` (or `OPENCODE_PLUGIN_DEBUG=1`)

Optional log file path:

- `ACONTEXT_PLUGIN_DEBUG_FILE=/tmp/opencode-acontext.log` (or `OPENCODE_PLUGIN_DEBUG_FILE=...`)

Logs are JSON Lines and include HTTP request/response metadata plus failure details.
