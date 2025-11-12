/**
 * Tool endpoints.
 */

import { RequesterProtocol } from '../client-types';
import {
  ToolRenameItem,
  ToolReferenceData,
  FlagResponse,
  FlagResponseSchema,
  ToolReferenceDataSchema,
} from '../types/tool';

export class ToolsAPI {
  constructor(private requester: RequesterProtocol) {}

  async renameToolName(options: {
    rename: ToolRenameItem[];
  }): Promise<FlagResponse> {
    const payload = { rename: options.rename };
    const data = await this.requester.request('PUT', '/tool/name', {
      jsonData: payload,
    });
    return FlagResponseSchema.parse(data);
  }

  async getToolName(): Promise<ToolReferenceData[]> {
    const data = await this.requester.request('GET', '/tool/name');
    return Array.isArray(data)
      ? data.map((item) => ToolReferenceDataSchema.parse(item))
      : [];
  }
}

