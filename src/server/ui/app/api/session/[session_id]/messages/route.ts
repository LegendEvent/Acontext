import { createApiResponse, createApiError } from "@/lib/api-response";
import { GetMessagesResp } from "@/types";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ session_id: string }> }
) {
  const session_id = (await params).session_id;
  if (!session_id) {
    return createApiError("session_id is required");
  }

  const { searchParams } = new URL(request.url);
  const limit = searchParams.get("limit") || "20";
  const cursor = searchParams.get("cursor") || "";
  const with_asset_public_url =
    searchParams.get("with_asset_public_url") || "true";

  const getMessages = new Promise<GetMessagesResp>(async (resolve, reject) => {
    try {
      const params = new URLSearchParams({
        limit,
        with_asset_public_url,
      });
      if (cursor) {
        params.append("cursor", cursor);
      }

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_SERVER_URL}/api/v1/session/${session_id}/messages?${params.toString()}`,
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
    const res = await getMessages;
    return createApiResponse(res);
  } catch (error) {
    console.error(error);
    return createApiError("Internal Server Error");
  }
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ session_id: string }> }
) {
  const session_id = (await params).session_id;
  if (!session_id) {
    return createApiError("session_id is required");
  }

  const sendMessage = new Promise<null>(async (resolve, reject) => {
    try {
      const contentType = request.headers.get("content-type") || "";

      let fetchOptions: RequestInit;

      // 判断是否是 multipart/form-data
      if (contentType.includes("multipart/form-data")) {
        // 处理文件上传
        const formData = await request.formData();

        // 直接转发 FormData 给后端
        fetchOptions = {
          method: "POST",
          headers: {
            Authorization: `Bearer sk-ac-${process.env.ROOT_API_BEARER_TOKEN}`,
            // 不设置 Content-Type，让浏览器自动设置 multipart/form-data boundary
          },
          body: formData,
        };
      } else {
        // 处理 JSON 数据
        const body = await request.json();

        fetchOptions = {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer sk-ac-${process.env.ROOT_API_BEARER_TOKEN}`,
          },
          body: JSON.stringify(body),
        };
      }

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_SERVER_URL}/api/v1/session/${session_id}/messages`,
        fetchOptions
      );

      if (response.status !== 201) {
        console.log(response.status);
        console.log(await response.json());
        reject(new Error("Internal Server Error"));
      }

      const result = await response.json();
      if (result.code !== 0) {
        reject(new Error(result.message));
      }
      resolve(null);
    } catch (error) {
      console.error("Send message error:", error);
      reject(new Error("Internal Server Error"));
    }
  });

  try {
    await sendMessage;
    return createApiResponse(null);
  } catch (error) {
    console.error(error);
    return createApiError("Internal Server Error");
  }
}

