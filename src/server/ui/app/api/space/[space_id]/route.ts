import { createApiResponse, createApiError } from "@/lib/api-response";

export async function DELETE(
  req: Request,
  { params }: { params: Promise<{ space_id: string }> }
) {
  const space_id = (await params).space_id;
  if (!space_id) {
    return createApiError("space_id is required");
  }

  const deleteSpace = new Promise<null>(async (resolve, reject) => {
    try {
      const response = await fetch(
        `${process.env.API_SERVER_URL}/api/v1/space/${space_id}`,
        {
          method: "DELETE",
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
      resolve(null);
    } catch {
      reject(new Error("Internal Server Error"));
    }
  });

  try {
    await deleteSpace;
    return createApiResponse(null);
  } catch (error) {
    console.error(error);
    return createApiError("Internal Server Error");
  }
}

