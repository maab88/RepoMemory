export type ApiError = {
  code: string;
  message: string;
};

export type ApiEnvelope<T> = {
  data?: T;
  error?: ApiError;
};

export type CurrentUser = {
  id: string;
  email?: string;
  displayName: string;
  avatarUrl?: string;
};

export type Organization = {
  id: string;
  name: string;
  slug: string;
  role: "owner" | "member";
};