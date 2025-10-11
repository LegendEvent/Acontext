import service, { Res } from "../http";
import { Artifact, ListFilesResp } from "@/types";

export const getArtifacts = async (): Promise<Res<Artifact[]>> => {
  return await service.get("/api/artifact");
};

export const getListFiles = async (
  artifact_id: string,
  path: string
): Promise<Res<ListFilesResp>> => {
  return await service.get(`/api/artifact/${artifact_id}/file/ls?path=${path}`);
};
