import { RequestError } from "@/lib/api-client";

export type MappedApiError = {
  title: string;
  message: string;
  retryable: boolean;
  code: string;
  requestId?: string;
};

export function mapApiError(error: unknown): MappedApiError {
  if (error instanceof RequestError) {
    const code = normalizeCode(error.code);
    const base: MappedApiError = {
      title: "Something went wrong",
      message: "The request failed. Please try again.",
      retryable: true,
      code,
      requestId: error.requestId,
    };

    switch (code) {
      case "GITHUB_RECONNECT_REQUIRED":
        return {
          ...base,
          title: "Reconnect GitHub required",
          message: "Your GitHub connection expired or is missing. Reconnect GitHub and retry.",
          retryable: false,
        };
      case "GITHUB_RATE_LIMITED":
        return {
          ...base,
          title: "GitHub rate limit reached",
          message: "GitHub is rate limiting requests right now. Please retry in a few minutes.",
          retryable: true,
        };
      case "OAUTH_CALLBACK_FAILED":
        return {
          ...base,
          title: "GitHub callback failed",
          message: "We couldn't complete the GitHub callback. Please start the connect flow again.",
          retryable: true,
        };
      case "JOB_FAILED":
        return {
          ...base,
          title: "Background job failed",
          message: "The job failed before completion. Review details and retry.",
          retryable: true,
        };
      case "FORBIDDEN":
        return {
          ...base,
          title: "Access denied",
          message: "You don't have permission for this action.",
          retryable: false,
        };
      case "NOT_FOUND":
        return {
          ...base,
          title: "Not found",
          message: "The requested resource could not be found.",
          retryable: false,
        };
      case "VALIDATION_ERROR":
        return {
          ...base,
          title: "Invalid request",
          message: error.message || "Please check the input and try again.",
          retryable: false,
        };
      default:
        return {
          ...base,
          message: error.message || base.message,
        };
    }
  }

  if (error instanceof Error) {
    return {
      title: "Unexpected error",
      message: error.message || "An unexpected error occurred.",
      retryable: true,
      code: "UNKNOWN",
    };
  }

  return {
    title: "Unexpected error",
    message: "An unexpected error occurred.",
    retryable: true,
    code: "UNKNOWN",
  };
}

function normalizeCode(code: string): string {
  const normalized = code.trim().toUpperCase();
  switch (normalized) {
    case "GITHUB_NOT_CONNECTED":
      return "GITHUB_RECONNECT_REQUIRED";
    case "FORBIDDEN":
    case "NOT_FOUND":
    case "VALIDATION_ERROR":
    case "JOB_FAILED":
    case "GITHUB_RECONNECT_REQUIRED":
    case "GITHUB_RATE_LIMITED":
    case "OAUTH_CALLBACK_FAILED":
      return normalized;
    default:
      return normalized;
  }
}
