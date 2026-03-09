import {
  ApiError as GeneratedApiError,
  OpenAPI,
  type ErrorResponse,
} from "@repomemory/contracts";

OpenAPI.BASE = "/api";
OpenAPI.CREDENTIALS = "same-origin";
OpenAPI.WITH_CREDENTIALS = false;

export class RequestError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
    public readonly requestId?: string
  ) {
    super(message);
  }
}

function isErrorResponse(value: unknown): value is ErrorResponse {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const maybe = value as { error?: unknown };
  if (typeof maybe.error !== "object" || maybe.error === null) {
    return false;
  }

  const apiError = maybe.error as { code?: unknown; message?: unknown; requestId?: unknown };
  if (apiError.requestId !== undefined && typeof apiError.requestId !== "string") {
    return false;
  }
  return typeof apiError.code === "string" && typeof apiError.message === "string";
}

function toRequestError(error: unknown): RequestError {
  if (error instanceof GeneratedApiError) {
    if (isErrorResponse(error.body)) {
      return new RequestError(error.status, error.body.error.code, error.body.error.message, error.body.error.requestId);
    }

    return new RequestError(error.status, "request_failed", error.message);
  }

  if (error instanceof Error) {
    return new RequestError(500, "request_failed", error.message);
  }

  return new RequestError(500, "request_failed", "request failed");
}

type WithData<T> = { data: T };

export async function unwrapData<T>(request: Promise<WithData<T>>): Promise<T> {
  try {
    const payload = await request;
    return payload.data;
  } catch (error) {
    throw toRequestError(error);
  }
}
