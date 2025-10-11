import { createApiResponse } from "@/lib/api-response";

/**
 * ping
 * @returns
 */
export async function GET() {
  return createApiResponse(null, "pong");
}
