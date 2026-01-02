import type { Plugin } from "@opencode-ai/plugin";
import crypto from "node:crypto";
import { appendFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";

const PLUGIN_INSTANCE_ID = crypto.randomUUID();
const PLUGIN_LOADED_AT = new Date().toISOString();

type AcontextConfig = {
  baseUrl: string;
  apiKey: string;

  searchBaseUrl?: string;
  searchApiKey?: string;

  mode?: "fast" | "agentic";
  limit?: number;
  maxIterations?: number;
  maxDistance?: number;

  injectHeader?: string;
};

type StateFile = {
  spaces?: Record<string, string>;
  sessions?: Record<string, string>;
};

type TextPart = {
  type: "text";
  text: string;
};

type FilePart = {
  type: "file";
  mime: string;
  url: string;
  filename?: string;
};

type OutputPart = TextPart | FilePart | { type: string; [k: string]: unknown };

type AcontextMessage = {
  role: "user" | "assistant";
  parts: Array<{ type: string; text?: string; meta?: Record<string, unknown>; file_field?: string; [k: string]: unknown }>;
  meta?: Record<string, unknown>;
};

function isTextPart(p: unknown): p is TextPart {
  return Boolean(p) && typeof p === "object" && (p as any).type === "text" && typeof (p as any).text === "string";
}

function isFilePart(p: unknown): p is FilePart {
  return (
    Boolean(p) &&
    typeof p === "object" &&
    (p as any).type === "file" &&
    typeof (p as any).mime === "string" &&
    typeof (p as any).url === "string"
  );
}

function truncateText(input: string, maxChars: number): { text: string; truncated: boolean } {
  if (input.length <= maxChars) return { text: input, truncated: false };
  return {
    text: input.slice(0, maxChars) + `\n[TRUNCATED to ${maxChars} chars]`,
    truncated: true,
  };
}

function stableJson(value: unknown): string {
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

type DebugLogger = {
  enabled: boolean;
  filePath: string;
  log: (msg: string, extra?: Record<string, unknown>) => Promise<void>;
};

function env(name: string): string | undefined {
  const v = process.env[name];
  return v && v.trim() ? v.trim() : undefined;
}

function envFlag(name: string): boolean {
  const v = env(name);
  if (!v) return false;
  return ["1", "true", "yes", "on"].includes(v.toLowerCase());
}

function createDebugLogger(pluginName: string): DebugLogger {
  const enabled = envFlag("OPENCODE_PLUGIN_DEBUG") || envFlag("ACONTEXT_PLUGIN_DEBUG");
  const filePath =
    env("OPENCODE_PLUGIN_DEBUG_FILE") ??
    env("ACONTEXT_PLUGIN_DEBUG_FILE") ??
    path.join(os.tmpdir(), `${pluginName}.log`);

  return {
    enabled,
    filePath,
    async log(msg, extra) {
      if (!enabled) return;
      const line = {
        ts: new Date().toISOString(),
        plugin: pluginName,
        msg,
        ...(extra ? { extra } : {}),
      };
      await appendFile(filePath, JSON.stringify(line) + "\n").catch(() => {});
    },
  };
}

function isUuid(v: unknown): v is string {
  return typeof v === "string" && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(v);
}

function joinUrl(baseUrl: string, route: string): string {
  return baseUrl.replace(/\/$/, "") + "/" + route.replace(/^\//, "");
}

function addQuery(url: string, query: Record<string, string | number | boolean | undefined>): string {
  const u = new URL(url);
  for (const [k, v] of Object.entries(query)) {
    if (v === undefined) continue;
    u.searchParams.set(k, String(v));
  }
  return u.toString();
}

function normalizeBearerToken(token: string): string {
  const t = token.trim();
  return /^Bearer\s+/i.test(t) ? t : `Bearer ${t}`;
}

async function fileExists(filePath: string): Promise<boolean> {
  try {
    return await Bun.file(filePath).exists();
  } catch {
    return false;
  }
}

async function readJsonFile<T>(filePath: string): Promise<T | undefined> {
  if (!(await fileExists(filePath))) return undefined;
  try {
    const text = await Bun.file(filePath).text();
    return JSON.parse(text) as T;
  } catch {
    return undefined;
  }
}

async function writeJsonFile(filePath: string, data: unknown): Promise<void> {
  const text = JSON.stringify(data, null, 2) + "\n";
  await Bun.write(filePath, text);
}

function shaWorktree(worktree: string): string {
  return crypto.createHash("sha256").update(worktree).digest("hex").slice(0, 16);
}

class AcontextApi {
  constructor(
    private cfg: AcontextConfig,
    private debug?: DebugLogger,
  ) {}

  private async requestEnvelope<T>(
    method: string,
    route: string,
    opts?: { query?: Record<string, any>; body?: unknown },
    override?: { baseUrl?: string; apiKey?: string; timeoutMs?: number },
  ): Promise<{ code?: number; msg?: string; data: T }> {
    const baseUrl = override?.baseUrl ?? this.cfg.baseUrl;
    const apiKey = override?.apiKey ?? this.cfg.apiKey;
    const timeoutMs = override?.timeoutMs ?? 5_000;

    const url = addQuery(joinUrl(baseUrl, route), opts?.query ?? {});

    await this.debug?.log("http.request", {
      method,
      url,
      route,
      hasBody: Boolean(opts?.body),
      timeoutMs,
    });

    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), timeoutMs);

    let res: Response;
    let rawText = "";
    let parsedBody: unknown = undefined;

    try {
      res = await fetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
          Authorization: normalizeBearerToken(apiKey),
        },
        body: opts?.body ? JSON.stringify(opts.body) : undefined,
        signal: controller.signal,
      });

      try {
        parsedBody = await res.clone().json();
      } catch {
        rawText = await res.text().catch(() => "");
      }
    } finally {
      clearTimeout(timeout);
    }

    const bodyPreview =
      rawText !== ""
        ? rawText.slice(0, 500)
        : parsedBody !== undefined
          ? JSON.stringify(parsedBody).slice(0, 500)
          : "";

    await this.debug?.log("http.response", {
      method,
      url,
      route,
      status: res.status,
      ok: res.ok,
      bodyPreview,
    });

    if (!res.ok) {
      const extra = bodyPreview ? `\n${bodyPreview}` : "";
      throw new Error(`Acontext ${method} ${route} failed: ${res.status} ${res.statusText}${extra}`);
    }

    if (parsedBody && typeof parsedBody === "object" && "data" in (parsedBody as any)) {
      const env = parsedBody as any;
      return { code: env.code, msg: env.msg, data: env.data as T };
    }

    if (parsedBody !== undefined) {
      return { data: parsedBody as T };
    }

    return { data: (rawText as unknown as T) };
  }

  async listSpaces(limit = 200): Promise<any[]> {
    const r = await this.requestEnvelope<{ items?: any[] }>("GET", "/space", { query: { limit } }, { timeoutMs: 5_000 });
    return r.data?.items ?? [];
  }

  async createSpace(name: string): Promise<{ id: string }> {
    const r = await this.requestEnvelope<any>("POST", "/space", { body: { configs: { name } } }, { timeoutMs: 5_000 });
    if (!r.data?.id) throw new Error("Acontext createSpace: missing id in response");
    return { id: String(r.data.id) };
  }

  async createSession(spaceId: string): Promise<{ id: string }> {
    const r = await this.requestEnvelope<any>(
      "POST",
      "/session",
      { body: { space_id: spaceId, configs: { mode: "chat" } } },
      { timeoutMs: 5_000 },
    );
    if (!r.data?.id) throw new Error("Acontext createSession: missing id in response");
    return { id: String(r.data.id) };
  }

  async storeMessage(sessionId: string, blob: AcontextMessage): Promise<void> {
    await this.requestEnvelope(
      "POST",
      `/session/${sessionId}/messages`,
      { body: { format: "acontext", blob } },
      { timeoutMs: 5_000 },
    );
  }

  async experienceSearch(
    spaceId: string,
    queryText: string,
    mode: string,
    limit: number,
    maxIterations?: number,
  ): Promise<any> {
    const searchBaseUrl = this.cfg.searchBaseUrl ?? this.cfg.baseUrl;
    const searchApiKey = this.cfg.searchApiKey ?? this.cfg.apiKey;

    const timeoutMs = mode === "agentic" ? 15_000 : 3_000;

    const r = await this.requestEnvelope<any>(
      "GET",
      `/space/${spaceId}/experience_search`,
      {
        query: { query: queryText, mode, limit, max_iterations: maxIterations },
      },
      { baseUrl: searchBaseUrl, apiKey: searchApiKey, timeoutMs },
    );

    return r.data;
  }
}

const DEFAULT_CONFIG: AcontextConfig = {
  baseUrl: "http://127.0.0.1:8029/api/v1",
  apiKey: "sk-ac-your-root-api-bearer-token",
  mode: "fast",
  limit: 5,
  maxIterations: 4,
  maxDistance: 0.8,
  injectHeader: "SKILLS REFERENCES:",
};

async function loadConfig(ctx: { worktree: string }): Promise<AcontextConfig> {
  const home = env("HOME") ?? "/home/legendevent";
  const userCfgPath = path.join(home, ".config", "opencode", "acontext.json");
  const projectCfgPath = path.join(ctx.worktree, ".opencode", "acontext.json");

  const userCfg = (await readJsonFile<Partial<AcontextConfig>>(userCfgPath)) ?? {};
  const projectCfg = (await readJsonFile<Partial<AcontextConfig>>(projectCfgPath)) ?? {};

  const cfg: AcontextConfig = {
    ...DEFAULT_CONFIG,
    ...userCfg,
    ...projectCfg,
  };

  cfg.baseUrl = (env("ACONTEXT_BASE_URL") ?? cfg.baseUrl).trim();
  cfg.apiKey = (env("ACONTEXT_API_KEY") ?? cfg.apiKey).trim();

  cfg.searchBaseUrl = (env("ACONTEXT_SEARCH_BASE_URL") ?? cfg.searchBaseUrl)?.trim();
  cfg.searchApiKey = (env("ACONTEXT_SEARCH_API_KEY") ?? cfg.searchApiKey)?.trim();

  cfg.searchBaseUrl = cfg.searchBaseUrl && cfg.searchBaseUrl.trim() ? cfg.searchBaseUrl.trim() : undefined;
  cfg.searchApiKey = cfg.searchApiKey && cfg.searchApiKey.trim() ? cfg.searchApiKey.trim() : undefined;

  cfg.mode = (env("ACONTEXT_MODE") as any) ?? cfg.mode;
  cfg.limit = env("ACONTEXT_LIMIT") ? Number(env("ACONTEXT_LIMIT")) : cfg.limit;
  cfg.maxIterations = env("ACONTEXT_MAX_ITERATIONS") ? Number(env("ACONTEXT_MAX_ITERATIONS")) : cfg.maxIterations;
  cfg.maxDistance = env("ACONTEXT_MAX_DISTANCE") ? Number(env("ACONTEXT_MAX_DISTANCE")) : cfg.maxDistance;

  return cfg;
}

async function loadState(): Promise<{ path: string; state: StateFile }> {
  const home = env("HOME") ?? "/home/legendevent";
  const statePath = path.join(home, ".config", "opencode", "acontext-state.json");
  const state = (await readJsonFile<StateFile>(statePath)) ?? {};
  state.spaces ??= {};
  state.sessions ??= {};
  return { path: statePath, state };
}

function formatSkillBlocks(citedBlocks: any[], maxDistance: number): string {
  const good = citedBlocks
    .filter((b) => (typeof b?.distance === "number" ? b.distance <= maxDistance : true))
    .slice(0, 5);

  if (good.length === 0) return "";

  const parts: string[] = [];
  for (const b of good) {
    const title = String(b?.title ?? b?.props?.use_when ?? "").trim();
    const props = b?.props ?? {};
    const preferences = String(props?.preferences ?? "").trim();
    const toolSops: any[] = Array.isArray(props?.tool_sops) ? props.tool_sops : [];

    const lines: string[] = [];
    lines.push(`Use when: ${title || "(untitled)"}`);
    if (preferences) lines.push(`Preferences: ${preferences}`);

    if (toolSops.length > 0) {
      lines.push("Tool SOPs:");
      for (const step of toolSops) {
        const order = step?.order ?? "";
        const tool = step?.tool_name ?? "";
        const action = step?.action ?? "";
        lines.push(`- ${order ? `${order}. ` : ""}${tool}: ${action}`.trim());
      }
    }

    parts.push(lines.join("\n"));
  }

  return parts.join("\n---\n");
}

function buildAcontextMessageFromOutputParts(
  role: AcontextMessage["role"],
  parts: OutputPart[],
  fallbackText: string,
  meta?: Record<string, unknown>,
): AcontextMessage {
  const outParts: AcontextMessage["parts"] = [];

  for (const p of parts) {
    if (isTextPart(p)) {
      const t = p.text.trim();
      if (!t) continue;
      outParts.push({ type: "text", text: t });
      continue;
    }

    if (isFilePart(p)) {
      // We do NOT upload files (multipart) here. Store as text placeholder.
      const label = (p.filename ?? "").trim() || p.url;
      outParts.push({ type: "text", text: `[file] ${label}` });
      continue;
    }
  }

  if (outParts.length === 0) {
    const t = fallbackText.trim() || "(no content)";
    outParts.push({ type: "text", text: t });
  }

  return { role, parts: outParts, ...(meta ? { meta } : {}) };
}

const AcontextPlugin: Plugin = async (ctx) => {
  const debug = createDebugLogger("opencode-acontext");
  await debug.log("plugin.init", { worktree: ctx.worktree, debugFile: debug.filePath, pluginInstanceID: PLUGIN_INSTANCE_ID });

  const config = await loadConfig({ worktree: ctx.worktree });
  await debug.log("config.loaded", {
    baseUrl: config.baseUrl,
    hasSearchBaseUrl: Boolean(config.searchBaseUrl),
    mode: config.mode,
    limit: config.limit,
    maxIterations: config.maxIterations,
    maxDistance: config.maxDistance,
    injectHeader: config.injectHeader,
  });

  const api = new AcontextApi(config, debug);

  const searchCache = new Map<string, { ts: number; cited_blocks: any[] }>();

  const systemSkillsCache = new Map<
    string,
    { ts: number; spaceId: string; sessionID: string; userText: string; skillsText: string; usedBlocksCount: number; wasCacheHit: boolean }
  >();

  const lastSessionForSpace = new Map<string, string>();

  const SYSTEM_SKILLS_TTL_MS = 60 * 1000;

  async function updateSystemSkillsForSession(input: { sessionID: string }, spaceId: string, userText: string): Promise<void> {
    const now = Date.now();
    const key = `${spaceId}|${input.sessionID}`;

    const cached = systemSkillsCache.get(key);
    if (cached && now - cached.ts < SYSTEM_SKILLS_TTL_MS) {
      lastSessionForSpace.set(spaceId, input.sessionID);
      return;
    }

    const cacheKey = `${spaceId}|${userText}`;
    let citedBlocks: any[] = [];
    let wasCacheHit = false;

    const hit = searchCache.get(cacheKey);
    if (hit && now - hit.ts < 90_000) {
      citedBlocks = hit.cited_blocks;
      wasCacheHit = true;
    } else {
      try {
        const result = await api.experienceSearch(
          spaceId,
          userText,
          config.mode ?? "fast",
          config.limit ?? 5,
          config.maxIterations,
        );
        citedBlocks = Array.isArray(result?.cited_blocks) ? result.cited_blocks : [];
        searchCache.set(cacheKey, { ts: now, cited_blocks: citedBlocks });
      } catch (e) {
        citedBlocks = [];
        await debug.log("search.failed", {
          error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
        });
      }
    }

    const skillsText = formatSkillBlocks(citedBlocks, config.maxDistance ?? 0.8);
    const usedBlocksCount = skillsText ? skillsText.split("\n---\n").filter(Boolean).length : 0;

    systemSkillsCache.set(key, {
      ts: now,
      spaceId,
      sessionID: input.sessionID,
      userText,
      skillsText,
      usedBlocksCount,
      wasCacheHit,
    });

    lastSessionForSpace.set(spaceId, input.sessionID);

    ctx.client.tui
      .showToast({
        body: {
          title: "Acontext",
          message: skillsText
            ? `Injected ${usedBlocksCount} skill${usedBlocksCount === 1 ? "" : "s"}${wasCacheHit ? " (cache)" : ""}`
            : "No skills injected",
          variant: "success",
          duration: 1500,
        },
      })
      .catch(async (e) => {
        await debug.log("tui.toast_failed", {
          error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
        });
      });
  }

  function buildSystemSkillsBlock(skillsText: string): string {
    if (!skillsText.trim()) return "";

    const marker = "<!-- opencode-acontext:v1 -->";
    const header = config.injectHeader ?? "SKILLS REFERENCES:";

    return `${marker}\n${header}\n${skillsText}`;
  }


  // For assistant response capture.
  const messageRoles = new Map<string, "user" | "assistant">();
  const pendingAssistantText = new Map<string, Map<string, string>>(); // sessionID -> (messageID -> latest text)

  async function ensureSpaceId(): Promise<string> {
    const { path: statePath, state } = await loadState();
    const key = shaWorktree(ctx.worktree);
    const existing = state.spaces?.[key];
    if (isUuid(existing)) {
      return existing;
    }

    const repoName = path.basename(ctx.worktree);
    const desiredName = `opencode:${repoName}`;

    try {
      const spaces = await api.listSpaces(200);
      const found = spaces.find((s) => s?.configs?.name === desiredName);
      if (found?.id) {
        state.spaces![key] = found.id;
        await writeJsonFile(statePath, state);
        return found.id;
      }
    } catch {
      // ignore
    }

    const created = await api.createSpace(desiredName);
    state.spaces![key] = created.id;
    await writeJsonFile(statePath, state);
    return created.id;
  }

  async function ensureAcontextSessionId(opencodeSessionId: string, spaceId: string): Promise<string> {
    const { path: statePath, state } = await loadState();
    const existing = state.sessions?.[opencodeSessionId];
    if (isUuid(existing)) return existing;

    const created = await api.createSession(spaceId);
    state.sessions![opencodeSessionId] = created.id;
    await writeJsonFile(statePath, state);
    return created.id;
  }

  async function storeAcontextMessage(opencodeSessionId: string, msg: AcontextMessage): Promise<void> {
    const spaceId = await ensureSpaceId().catch(async (e) => {
      await debug.log("space.ensure_failed", {
        error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
      });
      return "";
    });
    if (!isUuid(spaceId)) return;

    const acontextSessionId = await ensureAcontextSessionId(opencodeSessionId, spaceId).catch(async (e) => {
      await debug.log("session.ensure_failed", {
        opencodeSessionId,
        error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
      });
      return "";
    });
    if (!isUuid(acontextSessionId)) return;

    await api.storeMessage(acontextSessionId, msg).catch(async (e) => {
      await debug.log("message.store_failed", {
        opencodeSessionId,
        error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
      });
    });
  }

  const assistantStored = new Map<string, Set<string>>();
  const assistantInFlight = new Map<string, Set<string>>();

  function getOrInitSet(map: Map<string, Set<string>>, key: string): Set<string> {
    let s = map.get(key);
    if (!s) {
      s = new Set();
      map.set(key, s);
    }
    return s;
  }

  async function flushAssistantMessage(sessionID: string, messageID: string, reason: string): Promise<void> {
    if (!sessionID || !messageID) return;

    const stored = getOrInitSet(assistantStored, sessionID);
    if (stored.has(messageID)) return;

    const inFlight = getOrInitSet(assistantInFlight, sessionID);
    if (inFlight.has(messageID)) return;
    inFlight.add(messageID);

    try {
      const pending = pendingAssistantText.get(sessionID);
      const hasTextPart = Boolean(pending?.has(messageID));
      const text = pending?.get(messageID) ?? "";
      const trimmed = text.trim();

      if (!hasTextPart || trimmed.length === 0) {
        stored.add(messageID);

        if (pending?.has(messageID)) {
          pending.delete(messageID);
          if (pending.size === 0) pendingAssistantText.delete(sessionID);
        }

        return;
      }

      const { text: truncated } = truncateText(trimmed, 200_000);

      const msg: AcontextMessage = {
        role: "assistant",
        parts: [
          {
            type: "text",
            text: truncated,
            meta: {
              plugin_instance_id: PLUGIN_INSTANCE_ID,
              plugin_loaded_at: PLUGIN_LOADED_AT,
              opencode_session_id: sessionID,
              opencode_message_id: messageID,
              source: "opencode:event",
              flush_reason: reason,
            },
          },
        ],
      };

      await storeAcontextMessage(sessionID, msg);

      stored.add(messageID);

      if (pending?.has(messageID)) {
        pending.delete(messageID);
        if (pending.size === 0) pendingAssistantText.delete(sessionID);
      }
    } finally {
      inFlight.delete(messageID);
      if (inFlight.size === 0) assistantInFlight.delete(sessionID);
    }
  }

  async function flushAllPendingAssistant(sessionID: string, reason: string): Promise<void> {
    const pending = pendingAssistantText.get(sessionID);
    if (!pending || pending.size === 0) return;

    const messageIDs = Array.from(pending.keys());
    for (const messageID of messageIDs) {
      await flushAssistantMessage(sessionID, messageID, reason);
    }
  }

  return {
    // Capture assistant output & lifecycle.
    event: async ({ event }) => {
      try {
        if (event.type === "message.updated") {
          const info = (event as any).properties?.info;
          if (info?.id && (info.role === "user" || info.role === "assistant")) {
            messageRoles.set(String(info.id), info.role);
          }

          if (info?.role === "assistant" && info?.time?.completed && info?.sessionID) {
            const messageID = String(info.id ?? "");
            if (messageID) await flushAssistantMessage(String(info.sessionID), messageID, "message.completed");
          }

          return;
        }

        if (event.type === "message.part.updated") {
          const part = (event as any).properties?.part;
          if (!part || part.type !== "text") return;

          const messageID = String(part.messageID ?? "");
          const sessionID = String(part.sessionID ?? "");
          if (!messageID || !sessionID) return;

          const role = messageRoles.get(messageID);
          if (role !== "assistant") return;

          const text = typeof part.text === "string" ? part.text : "";

          let perSession = pendingAssistantText.get(sessionID);
          if (!perSession) {
            perSession = new Map();
            pendingAssistantText.set(sessionID, perSession);
          }
           perSession.set(messageID, text);

           return;

        }

        if (event.type === "session.idle") {
          const sessionID = String((event as any).properties?.sessionID ?? "");
          if (sessionID) await flushAllPendingAssistant(sessionID, "session.idle");
        }
      } catch (e) {
        await debug.log("event.handler_failed", {
          type: (event as any)?.type,
          error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
        });
      }
    },

    "tool.execute.before": async (input, output) => {
      function safeJsonString(value: unknown): string {
        try {
          return JSON.stringify(value ?? {}, null, 2);
        } catch {
          return JSON.stringify({ _non_json: String(value) }, null, 2);
        }
      }

      const argsJson = safeJsonString(output.args ?? {});
      const { text: argsText } = truncateText(argsJson, 200_000);

      const msg: AcontextMessage = {
        role: "assistant",
        parts: [
          {
            type: "tool-call",
            meta: {
              id: input.callID,
              name: input.tool,
              arguments: argsText,
              plugin_instance_id: PLUGIN_INSTANCE_ID,
              plugin_loaded_at: PLUGIN_LOADED_AT,
              opencode_session_id: input.sessionID,
              hook: "tool.execute.before",
            },
          },
        ],
      };

      await storeAcontextMessage(input.sessionID, msg);
    },

    "tool.execute.after": async (input, output) => {
      const { text: outText } = truncateText(typeof output.output === "string" ? output.output : stableJson(output.output), 200_000);

      const msg: AcontextMessage = {
        role: "assistant",
        parts: [
          {
            type: "tool-result",
            text: outText,
            meta: {
              tool_call_id: input.callID,
              plugin_instance_id: PLUGIN_INSTANCE_ID,
              plugin_loaded_at: PLUGIN_LOADED_AT,
              opencode_session_id: input.sessionID,
              hook: "tool.execute.after",
              title: output.title,
              metadata: output.metadata,
            },
          },
        ],
      };

      await storeAcontextMessage(input.sessionID, msg);
    },

    // User messages (and skill injection)
    "chat.message": async (input, output) => {
      const texts: string[] = [];
      output.parts.forEach((p: any) => {
        if (p?.type === "text" && typeof p.text === "string") texts.push(p.text);
      });

      if (texts.length === 0) return;

      const userText = texts.join("\n").trim();
      if (!userText) return;

      const shouldSkipInjection = userText.startsWith("[BACKGROUND TASK COMPLETED]");

      const spaceId = await ensureSpaceId().catch(async (e) => {
        await debug.log("space.ensure_failed", {
          error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
        });
        return "";
      });
      if (!isUuid(spaceId)) return;

      const parts = (Array.isArray(output.parts) ? output.parts : []) as OutputPart[];

      // For background task completion messages, store as-is (no injection).
      if (shouldSkipInjection) {
        const msg = buildAcontextMessageFromOutputParts("user", parts, userText, {
          plugin_instance_id: PLUGIN_INSTANCE_ID,
          plugin_loaded_at: PLUGIN_LOADED_AT,
          opencode_session_id: input.sessionID,
          hook: "chat.message",
          injection: "skipped_background_task_completed",
        });

        await storeAcontextMessage(input.sessionID, msg);
        return;
      }

      await updateSystemSkillsForSession(input, spaceId, userText);


      const storeParts = parts;

      const finalUserTextParts: string[] = [];
      for (const p of storeParts) {
        if (isTextPart(p)) finalUserTextParts.push(p.text);
      }
      const finalUserText = finalUserTextParts.join("\n").trim() || userText;

      // IMPORTANT: per your request, no dedupe here. If OpenCode emits multiple chat.message events,
      // they will be stored multiple times (makes debugging easier).
      const msg = buildAcontextMessageFromOutputParts("user", storeParts, finalUserText, {
        plugin_instance_id: PLUGIN_INSTANCE_ID,
        plugin_loaded_at: PLUGIN_LOADED_AT,
        opencode_session_id: input.sessionID,
        hook: "chat.message",
      });

      await storeAcontextMessage(input.sessionID, msg);

    },

    "experimental.chat.system.transform": async (_input, output) => {
      const spaceId = await ensureSpaceId().catch(async (e) => {
        await debug.log("space.ensure_failed", {
          error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
        });
        return "";
      });
      if (!isUuid(spaceId)) return;

      const sessionID = lastSessionForSpace.get(spaceId);
      if (!sessionID) return;

      const entry = systemSkillsCache.get(`${spaceId}|${sessionID}`);
      if (!entry) return;
      if (Date.now() - entry.ts >= SYSTEM_SKILLS_TTL_MS) return;

      const block = buildSystemSkillsBlock(entry.skillsText);
      if (!block) return;

      const alreadyInjected = output.system.some((s) => typeof s === "string" && s.includes("<!-- opencode-acontext:v1 -->"));
      if (alreadyInjected) return;

      output.system.push(block);
    },
  };
};


export default AcontextPlugin;
