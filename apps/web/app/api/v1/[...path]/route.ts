import { NextRequest } from "next/server";

const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

function buildMockHeaders(request: NextRequest): Headers {
  const headers = new Headers();
  const userID = request.headers.get("x-user-id") ?? process.env.MOCK_USER_ID ?? "";
  const userEmail = request.headers.get("x-user-email") ?? process.env.MOCK_USER_EMAIL ?? "";
  const userName = request.headers.get("x-user-name") ?? process.env.MOCK_USER_NAME ?? "";

  if (userID) {
    headers.set("x-user-id", userID);
  }
  if (userEmail) {
    headers.set("x-user-email", userEmail);
  }
  if (userName) {
    headers.set("x-user-name", userName);
  }

  return headers;
}

async function forward(request: NextRequest, path: string[]) {
  const url = `${apiBaseUrl}/v1/${path.join("/")}`;
  const body = request.method === "GET" ? undefined : await request.text();

  const upstream = await fetch(url, {
    method: request.method,
    headers: buildMockHeaders(request),
    body,
    cache: "no-store",
  });

  const text = await upstream.text();
  return new Response(text, {
    status: upstream.status,
    headers: {
      "Content-Type": upstream.headers.get("Content-Type") ?? "application/json",
    },
  });
}

export async function GET(request: NextRequest, context: { params: Promise<{ path: string[] }> }) {
  const { path } = await context.params;
  return forward(request, path);
}

export async function POST(request: NextRequest, context: { params: Promise<{ path: string[] }> }) {
  const { path } = await context.params;
  return forward(request, path);
}