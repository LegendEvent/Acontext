import type { Plugin } from "@opencode-ai/plugin";
import crypto from "node:crypto";
import { appendFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";

type AcontextConfig = {
  /** Base URL for write APIs: space/session/messages */
  baseUrl: string;
  apiKey: string;

  /** Optional separate base URL + key for experience_search */
  searchBaseUrl?: string;
  searchApiKey?: string;

  mode?: "fast" | "agentic";
  limit?: number;
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

function isHttpUrl(url: string): boolean {
  return /^https?:\/\//i.test(url);
}

function parseBase64DataUrl(url: string): { mime: string; base64: string } | undefined {
  const m = /^data:([^;,]+);base64,(.+)$/i.exec(url);
  if (!m) return undefined;
  return { mime: m[1] ?? "application/octet-stream", base64: m[2] ?? "" };
}

function fileUrlToPath(url: string): string | undefined {
  if (!/^file:\/\//i.test(url)) return undefined;
  try {
    return decodeURIComponent(new URL(url).pathname);
  } catch {
    return undefined;
  }
}

async function getFileBase64(filePart: FilePart): Promise<{ mime: string; base64: string } | undefined> {
  const dataUrl = parseBase64DataUrl(filePart.url);
  if (dataUrl) return { mime: filePart.mime || dataUrl.mime, base64: dataUrl.base64 };

  const fsPath = fileUrlToPath(filePart.url);
  if (!fsPath) return undefined;

  try {
    const buf = await Bun.file(fsPath).arrayBuffer();
    return { mime: filePart.mime, base64: Buffer.from(buf).toString("base64") };
  } catch {
    return undefined;
  }
}

type OpenAIMessage = {
  role: "user" | "assistant" | "system";
  content:
    | string
    | Array<
        | { type: "text"; text: string }
        | { type: "image_url"; image_url: { url: string; detail?: "low" | "high" | "auto" } }
      >;
};

type AnthropicMessage = {
  role: "user" | "assistant";
  content: Array<
    | { type: "text"; text: string }
    | { type: "image"; source: { type: "base64"; media_type: string; data: string } }
    | { type: "document"; source: { type: "base64"; media_type: string; data: string }; title?: string }
  >;
};

async function buildOpenAIMessage(parts: OutputPart[], fallbackText: string): Promise<OpenAIMessage> {
  const hasFiles = parts.some(isFilePart);
  if (!hasFiles) {
    return { role: "user", content: fallbackText };
  }

  const content: Exclude<OpenAIMessage["content"], string> = [];

  for (const p of parts) {
    if (isTextPart(p)) {
      const t = p.text.trim();
      if (t) content.push({ type: "text", text: t });
      continue;
    }

    if (isFilePart(p)) {
      const isImage = p.mime.toLowerCase().startsWith("image/");
      if (isImage) {
        const dataUrl = parseBase64DataUrl(p.url);
        if (dataUrl) {
          content.push({ type: "image_url", image_url: { url: p.url } });
          continue;
        }

        const fsPath = fileUrlToPath(p.url);
        if (fsPath) {
          const base64 = await getFileBase64(p);
          if (base64?.base64) {
            content.push({ type: "image_url", image_url: { url: `data:${base64.mime};base64,${base64.base64}` } });
            continue;
          }
        }

        if (isHttpUrl(p.url)) {
          content.push({ type: "image_url", image_url: { url: p.url } });
          continue;
        }
      }

      const label = (p.filename ?? "").trim() || p.url;
      content.push({ type: "text", text: `[file] ${label}` });
    }
  }

  if (content.length === 0) content.push({ type: "text", text: fallbackText || "(no content)" });

  return { role: "user", content };
}

async function buildAnthropicMessage(parts: OutputPart[], fallbackText: string): Promise<AnthropicMessage> {
  const content: AnthropicMessage["content"] = [];

  for (const p of parts) {
    if (isTextPart(p)) {
      const t = p.text.trim();
      if (t) content.push({ type: "text", text: t });
      continue;
    }

    if (isFilePart(p)) {
      const mime = p.mime.toLowerCase();
      const isImage = mime.startsWith("image/");
      const isPdf = mime === "application/pdf";
      const base64 = await getFileBase64(p);

      if (base64?.base64 && isImage) {
        content.push({
          type: "image",
          source: { type: "base64", media_type: base64.mime, data: base64.base64 },
        });
        continue;
      }

      if (base64?.base64 && isPdf) {
        content.push({
          type: "document",
          title: p.filename,
          source: { type: "base64", media_type: base64.mime, data: base64.base64 },
        });
        continue;
      }

      const label = (p.filename ?? "").trim() || p.url;
      content.push({ type: "text", text: `[file] ${label}` });
    }
  }

  if (content.length === 0 && fallbackText) content.push({ type: "text", text: fallbackText });
  if (content.length === 0) content.push({ type: "text", text: "(no content)" });

  return { role: "user", content };
}

async function buildAcontextStorePayload(
  parts: OutputPart[],
  providerID: unknown,
  fallbackText: string,
): Promise<{ format: "openai" | "anthropic"; blob: OpenAIMessage | AnthropicMessage }> {
  if (providerID === "anthropic") {
    return { format: "anthropic", blob: await buildAnthropicMessage(parts, fallbackText) };
  }

  return { format: "openai", blob: await buildOpenAIMessage(parts, fallbackText) };
}

const DEFAULT_CONFIG: AcontextConfig = {
  baseUrl: "http://127.0.0.1:8029/api/v1",
  apiKey: "sk-ac-your-root-api-bearer-token",
  mode: "fast",
  limit: 5,
  maxDistance: 0.8,
  injectHeader: "SKILLS REFERENCES:",
};

function envFlag(name: string): boolean {
  const v = env(name);
  if (!v) return false;
  return ["1", "true", "yes", "on"].includes(v.toLowerCase());
}

function redactSecrets(s: string): string {
  return s
    .replace(/(Bearer\s+)([^\s]+)/gi, "$1[REDACTED]")
    .replace(/("apiKey"\s*:\s*")([^"]+)(")/gi, "$1[REDACTED]$3")
    .replace(/("searchApiKey"\s*:\s*")([^"]+)(")/gi, "$1[REDACTED]$3");
}

type DebugLogger = {
  enabled: boolean;
  filePath: string;
  log: (msg: string, extra?: Record<string, unknown>) => Promise<void>;
};

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
      await appendFile(filePath, redactSecrets(JSON.stringify(line)) + "\n").catch(() => {});
    },
  };
}

function env(name: string): string | undefined {
  const v = process.env[name];
  return v && v.trim() ? v.trim() : undefined;
}

function shaWorktree(worktree: string): string {
  return crypto.createHash("sha256").update(worktree).digest("hex").slice(0, 16);
}

function isUuid(v: unknown): v is string {
  return typeof v === "string" && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(v);
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
      queryKeys: opts?.query ? Object.keys(opts.query) : [],
      timeoutMs,
    });

    let res: Response;
    let rawText = "";
    let parsedBody: unknown = undefined;

    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), timeoutMs);

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
    } catch (e) {
      await this.debug?.log("http.fetch_error", {
        method,
        url,
        route,
        error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
      });
      throw e;
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

  async storeMessage(
    sessionId: string,
    payload: {
      format: "openai" | "anthropic";
      blob: unknown;
    },
  ): Promise<void> {
    await this.requestEnvelope(
      "POST",
      `/session/${sessionId}/messages`,
      { body: payload },
      { timeoutMs: 5_000 },
    );
  }

  async experienceSearch(spaceId: string, queryText: string, mode: string, limit: number): Promise<any> {
    const searchBaseUrl = this.cfg.searchBaseUrl ?? this.cfg.baseUrl;
    const searchApiKey = this.cfg.searchApiKey ?? this.cfg.apiKey;

    const timeoutMs = mode === "agentic" ? 15_000 : 3_000;

    const r = await this.requestEnvelope<any>(
      "GET",
      `/space/${spaceId}/experience_search`,
      {
        query: { query: queryText, mode, limit },
      },
      { baseUrl: searchBaseUrl, apiKey: searchApiKey, timeoutMs },
    );

    return r.data;
  }
}

function formatSkillBlocks(citedBlocks: any[], maxDistance: number): string {
  const good = citedBlocks
    .filter((b) => typeof b?.distance === "number" ? b.distance <= maxDistance : true)
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

  cfg.baseUrl = cfg.baseUrl.trim();
  cfg.apiKey = cfg.apiKey.trim();

  cfg.searchBaseUrl = (env("ACONTEXT_SEARCH_BASE_URL") ?? cfg.searchBaseUrl)?.trim();
  cfg.searchApiKey = (env("ACONTEXT_SEARCH_API_KEY") ?? cfg.searchApiKey)?.trim();

  cfg.searchBaseUrl = cfg.searchBaseUrl && cfg.searchBaseUrl.trim() ? cfg.searchBaseUrl.trim() : undefined;
  cfg.searchApiKey = cfg.searchApiKey && cfg.searchApiKey.trim() ? cfg.searchApiKey.trim() : undefined;

  cfg.mode = (env("ACONTEXT_MODE") as any) ?? cfg.mode;
  cfg.limit = env("ACONTEXT_LIMIT") ? Number(env("ACONTEXT_LIMIT")) : cfg.limit;
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

const AcontextPlugin: Plugin = async (ctx) => {
  const debug = createDebugLogger("opencode-acontext");
  await debug.log("plugin.init", {
    worktree: ctx.worktree,
    debugFile: debug.filePath,
  });

  const config = await loadConfig({ worktree: ctx.worktree });
  await debug.log("config.loaded", {
    baseUrl: config.baseUrl,
    hasSearchBaseUrl: Boolean(config.searchBaseUrl),
    mode: config.mode,
    limit: config.limit,
    maxDistance: config.maxDistance,
    injectHeader: config.injectHeader,
  });

  const api = new AcontextApi(config, debug);

  const cache = new Map<string, { ts: number; cited_blocks: any[] }>();

  async function ensureSpaceId(): Promise<string> {
    const { path: statePath, state } = await loadState();
    const key = shaWorktree(ctx.worktree);
    const existing = state.spaces?.[key];
    if (isUuid(existing)) {
      await debug.log("space.cache_hit", { key, spaceId: existing });
      return existing;
    }

    const repoName = path.basename(ctx.worktree);
    const desiredName = `opencode:${repoName}`;
    await debug.log("space.ensure", { key, desiredName });

    try {
      const spaces = await api.listSpaces(200);
      const found = spaces.find((s) => s?.configs?.name === desiredName);
      if (found?.id) {
        state.spaces![key] = found.id;
        await writeJsonFile(statePath, state);
        await debug.log("space.found", { desiredName, spaceId: found.id });
        return found.id;
      }
    } catch (e) {
      await debug.log("space.list_failed", {
        error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
      });
    }

    const created = await api.createSpace(desiredName);
    state.spaces![key] = created.id;
    await writeJsonFile(statePath, state);
    await debug.log("space.created", { desiredName, spaceId: created.id });
    return created.id;
  }

  async function ensureAcontextSessionId(opencodeSessionId: string, spaceId: string): Promise<string> {
    const { path: statePath, state } = await loadState();
    const existing = state.sessions?.[opencodeSessionId];
    if (isUuid(existing)) {
      await debug.log("session.cache_hit", { opencodeSessionId, acontextSessionId: existing });
      return existing;
    }

    const created = await api.createSession(spaceId);
    state.sessions![opencodeSessionId] = created.id;
    await writeJsonFile(statePath, state);
    await debug.log("session.created", { opencodeSessionId, acontextSessionId: created.id, spaceId });
    return created.id;
  }

  return {
    "chat.message": async (input, output) => {
      await debug.log("chat.message.enter", {
        opencodeSessionId: input.sessionID,
        partsCount: Array.isArray(output.parts) ? output.parts.length : undefined,
      });

      const texts: string[] = [];
      output.parts.forEach((p: any) => {
        if (p?.type === "text" && typeof p.text === "string") texts.push(p.text);
      });

      if (texts.length === 0) {
        await debug.log("chat.message.no_text_parts", {});
        return;
      }

       const userText = texts.join("\n").trim();
       if (!userText) {
         await debug.log("chat.message.empty_text", {});
         return;
       }

       const shouldSkipInjection = userText.startsWith("[BACKGROUND TASK COMPLETED]");
       if (shouldSkipInjection) {
         await debug.log("inject.skip_background_task_completed", {});
       }

      const spaceId = await ensureSpaceId().catch(async (e) => {
        await debug.log("space.ensure_failed", {
          error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
        });
        return "";
      });
      if (!isUuid(spaceId)) {
        await debug.log("space.invalid_id", { spaceId });
        return;
      }

      const acontextSessionId = await ensureAcontextSessionId(input.sessionID, spaceId).catch(async (e) => {
        await debug.log("session.ensure_failed", {
          error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
        });
        return "";
      });
      if (!isUuid(acontextSessionId)) {
        await debug.log("session.invalid_id", { acontextSessionId });
        return;
      }

      const providerID = (output as any)?.message?.model?.providerID;
      const parts = (Array.isArray(output.parts) ? output.parts : []) as OutputPart[];

      const payload = await buildAcontextStorePayload(parts, providerID, userText).catch(async (e) => {
        await debug.log("message.build_payload_failed", {
          error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
        });
        return { format: "openai" as const, blob: { role: "user", content: userText } satisfies OpenAIMessage };
      });

      api.storeMessage(acontextSessionId, payload).catch(async (e) => {
        await debug.log("message.store_failed", {
          error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
          format: payload.format,
        });
      });

       if (shouldSkipInjection) return;

       const cacheKey = `${spaceId}|${userText}`;
       const now = Date.now();
       let citedBlocks: any[] = [];
       let wasCacheHit = false;
       const hit = cache.get(cacheKey);
       if (hit && now - hit.ts < 90_000) {
         citedBlocks = hit.cited_blocks;
         wasCacheHit = true;
         await debug.log("search.cache_hit", { ageMs: now - hit.ts, citedBlocks: citedBlocks.length });
       } else {
         try {
           const result = await api.experienceSearch(spaceId, userText, config.mode ?? "fast", config.limit ?? 5);
           citedBlocks = Array.isArray(result?.cited_blocks) ? result.cited_blocks : [];
           cache.set(cacheKey, { ts: now, cited_blocks: citedBlocks });
           await debug.log("search.ok", { citedBlocks: citedBlocks.length });
         } catch (e) {
           citedBlocks = [];
           await debug.log("search.failed", {
             error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
           });
         }
       }

       const skillsText = formatSkillBlocks(citedBlocks, config.maxDistance ?? 0.8);
       const usedBlocksCount = skillsText ? skillsText.split("\n---\n").filter(Boolean).length : 0;
       if (!skillsText) {
         await debug.log("inject.skip_no_skills", { citedBlocks: citedBlocks.length });
         return;
       }


      const marker = "<!-- opencode-acontext:v1 -->";

      const injectedSkills = `${marker}\n${config.injectHeader ?? "SKILLS REFERENCES:"}\n${skillsText}`;

      for (const p of output.parts as any[]) {
        if (p?.type === "text" && typeof p.text === "string" && p.text.includes(marker)) {
          await debug.log("inject.skip_already_injected", {});
          return;
        }
      }

      const newParts: any[] = [];
      let inserted = false;
      for (const p of output.parts as any[]) {
        if (!inserted && p?.type === "text" && typeof p.text === "string") {
          newParts.push({
            type: "text",
            text: injectedSkills,
            synthetic: true,
          });
          inserted = true;
        }
        newParts.push(p);
      }

      if (!inserted) return;

      output.parts = newParts;

      ctx.client.tui
        .showToast({
          body: {
            title: "Acontext",
            message: `Injected ${usedBlocksCount} skill${usedBlocksCount === 1 ? "" : "s"}${wasCacheHit ? " (cache)" : ""}`,
            variant: "success",
            duration: 1500,
          },
        })
        .catch(async (e) => {
          await debug.log("tui.toast_failed", {
            error: e instanceof Error ? { name: e.name, message: e.message, stack: e.stack } : String(e),
          });
        });

      await debug.log("inject.done", { injectedChars: injectedSkills.length, partsCount: newParts.length, usedBlocksCount, wasCacheHit });
    },
  };
};

export default AcontextPlugin;
