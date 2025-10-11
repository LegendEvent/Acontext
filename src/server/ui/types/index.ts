export interface Artifact {
  id: string;
  project_id: string;
  created_at: string;
  updated_at: string;
}

export interface File {
  artifact_id: string;
  path: string;
  filename: string;
  meta: {
    __file_info__: {
      filename: string;
      mime: string;
      path: string;
      size: number;
    };
    [key: string]: unknown;
  };
  created_at: string;
  updated_at: string;
}

export interface ListFilesResp {
  files: File[];
  directories: string[];
}
