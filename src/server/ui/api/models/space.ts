import service, { Res } from "../http";
import { Space, Session, GetMessagesResp } from "@/types";

// Space APIs
export const getSpaces = async (): Promise<Res<Space[]>> => {
  return await service.get("/api/space");
};

export const createSpace = async (
  configs?: Record<string, unknown>
): Promise<Res<Space>> => {
  return await service.post("/api/space", { configs: configs || {} });
};

export const deleteSpace = async (space_id: string): Promise<Res<null>> => {
  return await service.delete(`/api/space/${space_id}`);
};

export const getSpaceConfigs = async (space_id: string): Promise<Res<Space>> => {
  return await service.get(`/api/space/${space_id}/configs`);
};

export const updateSpaceConfigs = async (
  space_id: string,
  configs: Record<string, unknown>
): Promise<Res<null>> => {
  return await service.put(`/api/space/${space_id}/configs`, { configs });
};

// Session APIs
export const getSessions = async (): Promise<Res<Session[]>> => {
  return await service.get("/api/session");
};

export const createSession = async (
  space_id?: string,
  configs?: Record<string, unknown>
): Promise<Res<Session>> => {
  return await service.post("/api/session", {
    space_id: space_id || "",
    configs: configs || {},
  });
};

export const deleteSession = async (session_id: string): Promise<Res<null>> => {
  return await service.delete(`/api/session/${session_id}`);
};

export const getSessionConfigs = async (
  session_id: string
): Promise<Res<Session>> => {
  return await service.get(`/api/session/${session_id}/configs`);
};

export const updateSessionConfigs = async (
  session_id: string,
  configs: Record<string, unknown>
): Promise<Res<null>> => {
  return await service.put(`/api/session/${session_id}/configs`, { configs });
};

export const connectSessionToSpace = async (
  session_id: string,
  space_id: string
): Promise<Res<null>> => {
  return await service.post(`/api/session/${session_id}/connect_to_space`, {
    space_id,
  });
};

// Message APIs
export const getMessages = async (
  session_id: string,
  limit: number = 20,
  cursor?: string,
  with_asset_public_url: boolean = true
): Promise<Res<GetMessagesResp>> => {
  const params = new URLSearchParams({
    limit: limit.toString(),
    with_asset_public_url: with_asset_public_url.toString(),
  });
  if (cursor) {
    params.append("cursor", cursor);
  }
  return await service.get(
    `/api/session/${session_id}/messages?${params.toString()}`
  );
};

export interface MessagePartIn {
  type: "text" | "image" | "audio" | "video" | "file" | "tool-call" | "tool-result" | "data";
  text?: string;
  file_field?: string;
  meta?: Record<string, unknown>;
}

export const sendMessage = async (
  session_id: string,
  role: "user" | "assistant" | "system" | "tool" | "function",
  parts: MessagePartIn[]
): Promise<Res<null>> => {
  return await service.post(`/api/session/${session_id}/messages`, {
    role,
    parts,
  });
};

