import { createApiResponse, createApiError } from "@/lib/api-response";
import { Space } from "@/types";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ space_id: string }> }
) {
  const space_id = (await params).space_id;
  if (!space_id) {
    return createApiError("space_id is required");
  }

  const getConfigs = new Promise<Space>(async (resolve, reject) => {
    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_SERVER_URL}/api/v1/space/${space_id}/configs`,
        {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer sk-ac-${process.env.ROOT_API_BEARER_TOKEN}`,
          },
        }
      );
      if (response.status !== 200) {
        reject(new Error("Internal Server Error"));
      }

      const result = await response.json();
      if (result.code !== 0) {
        reject(new Error(result.message));
      }
      resolve(result.data);
    } catch {
      reject(new Error("Internal Server Error"));
    }
  });

  try {
    const res = await getConfigs;
    return createApiResponse(res);
  } catch (error) {
    console.error(error);
    return createApiError("Internal Server Error");
  }
}

export async function PUT(
  request: Request,
  { params }: { params: Promise<{ space_id: string }> }
) {
  const space_id = (await params).space_id;
  if (!space_id) {
    return createApiError("space_id is required");
  }

  const body = await request.json();
  const updateConfigs = new Promise<null>(async (resolve, reject) => {
    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_SERVER_URL}/api/v1/space/${space_id}/configs`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer sk-ac-${process.env.ROOT_API_BEARER_TOKEN}`,
          },
          body: JSON.stringify(body),
        }
      );
      if (response.status !== 200) {
        reject(new Error("Internal Server Error"));
      }

      const result = await response.json();
      if (result.code !== 0) {
        reject(new Error(result.message));
      }
      resolve(null);
    } catch {
      reject(new Error("Internal Server Error"));
    }
  });

  try {
    await updateConfigs;
    return createApiResponse(null);
  } catch (error) {
    console.error(error);
    return createApiError("Internal Server Error");
  }
}

