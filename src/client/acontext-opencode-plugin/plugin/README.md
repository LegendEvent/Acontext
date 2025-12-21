# OpenCode Acontext Plugin

File: `~/.config/opencode/plugin/50-acontext.ts`

This plugin connects OpenCode to an Acontext server. It:

- Creates/uses an Acontext **space** per worktree (`opencode:<repoName>`)
- Creates an Acontext **session** per OpenCode session
- Stores user messages to Acontext
- Runs `experience_search` on each user message and injects the most relevant “skill blocks” into the prompt

## Configuration

Configuration is loaded (in this order):

1. Plugin defaults (see `DEFAULT_CONFIG` in `50-acontext.ts`)
2. User config: `~/.config/opencode/acontext.json`
3. Project config: `<repo>/.opencode/acontext.json`
4. Environment variables (override everything)

### `acontext.json` fields

```json
{
  "baseUrl": "http://127.0.0.1:8029/api/v1",
  "apiKey": "Bearer sk-ac-your-root-api-bearer-token",

  "searchBaseUrl": "https://acontext.example.com:8443/api/v1",
  "searchApiKey": "Bearer <proxy-bearer-token>",

  "mode": "fast",
  "limit": 5,
  "maxIterations": 4,
  "maxDistance": 0.8,
  "injectHeader": "SKILLS REFERENCES:"
}
```

Notes:
- `apiKey`/`searchApiKey` can be either `Bearer <token>` or just `<token>`; the plugin will normalize it to `Bearer <token>`.
- If you don’t set `searchBaseUrl`/`searchApiKey`, search uses `baseUrl`/`apiKey`.
- `maxIterations` is sent to Acontext as `max_iterations` (only relevant in `agentic` mode).

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

The plugin caches created IDs in:

- `~/.config/opencode/acontext-state.json`

This maps:

- worktree hash → `spaceId`
- OpenCode session ID → `acontextSessionId`

## Debug logging

Enable debug logging:

- `ACONTEXT_PLUGIN_DEBUG=1` (or `OPENCODE_PLUGIN_DEBUG=1`)

Optional log file path:

- `ACONTEXT_PLUGIN_DEBUG_FILE=/tmp/opencode-acontext.log` (or `OPENCODE_PLUGIN_DEBUG_FILE=...`)

Logs are JSON lines and attempt to redact bearer tokens and API keys.

## Network / Firewall notes (ACME + IP allowlist)

A common deployment is:

- Public entry: Caddy on `:8443` (TLS + bearer token + path allowlist)
- ACME HTTP-01 validation: Caddy on `:80`
- Internal Acontext API: `127.0.0.1:8029`
- Optional nftables allowlist for `:8443` via set `inet acontext_fw dynamic_allowed4`

### Manually allow an IP (nftables)

Add an IPv4 address to the allowlist set:

```bash
sudo nft add element inet acontext_fw dynamic_allowed4 { 1.2.3.4 timeout 1d }
```

Use the set’s default timeout:

```bash
sudo nft add element inet acontext_fw dynamic_allowed4 { 1.2.3.4 }
```

Remove an IP again:

```bash
sudo nft delete element inet acontext_fw dynamic_allowed4 { 1.2.3.4 }
```

Show current allowlisted IPs:

```bash
sudo nft list set inet acontext_fw dynamic_allowed4
```

## Allowed HTTP paths (Caddy)

If you’re using the recommended proxy setup described above, the Caddyfile typically allowlists only:

- `GET /api/v1/ping`
- `GET /api/v1/space/<spaceId>/experience_search`

`/` will return `404` by design.
