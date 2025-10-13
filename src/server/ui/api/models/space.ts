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
export const getSessions = async (
  spaceId?: string,
  notConnected?: boolean
): Promise<Res<Session[]>> => {
  const params = new URLSearchParams();
  if (spaceId) {
    params.append("space_id", spaceId);
  }
  if (notConnected !== undefined) {
    params.append("not_connected", notConnected.toString());
  }
  const queryString = params.toString();
  return await service.get(
    `/api/session${queryString ? `?${queryString}` : ""}`
  );
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
  parts: MessagePartIn[],
  files?: Record<string, File>
): Promise<Res<null>> => {
  // 判断是否有文件需要上传
  const hasFiles = files && Object.keys(files).length > 0;

  if (hasFiles) {
    // 使用 multipart/form-data
    const formData = new FormData();

    // 添加 payload 字段（JSON 字符串）
    formData.append("payload", JSON.stringify({ role, parts }));

    // 添加文件
    for (const [fieldName, file] of Object.entries(files!)) {
      formData.append(fieldName, file);
    }

    // FormData 会自动设置 Content-Type 为 multipart/form-data
    return await service.post(`/api/session/${session_id}/messages`, formData);
  } else {
    // 使用 JSON 方式
    return await service.post(`/api/session/${session_id}/messages`, {
      role,
      parts,
    });
  }
};

