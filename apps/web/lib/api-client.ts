import { ApiEnvelope } from "@/lib/types";

export class RequestError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string
  ) {
    super(message);
  }
}

async function parseResponse<T>(response: Response): Promise<T> {
  const payload = (await response.json()) as ApiEnvelope<T>;

  if (!response.ok || payload.error) {
    const code = payload.error?.code ?? "request_failed";
    const message = payload.error?.message ?? `request failed with status ${response.status}`;
    throw new RequestError(response.status, code, message);
  }

  if (!payload.data) {
    throw new RequestError(response.status, "invalid_response", "missing response data");
  }

  return payload.data;
}

export async function apiRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/v1${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    cache: "no-store",
  });

  return parseResponse<T>(response);
}