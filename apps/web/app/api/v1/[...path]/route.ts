import { NextRequest } from "next/server";
import { getServerSession } from "next-auth/next";
import { authOptions } from "@/lib/auth/auth-options";
import { signAPIToken } from "@/lib/auth/api-token";

const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

function unauthorizedResponse() {
  return new Response(JSON.stringify({ error: { code: "unauthorized", message: "authentication required" } }), {
    status: 401,
    headers: {
      "Content-Type": "application/json",
    },
  });
}

function buildForwardHeaders(request: NextRequest, bearerToken: string): Headers {
  const headers = new Headers();
  headers.set("Authorization", `Bearer ${bearerToken}`);

  const contentType = request.headers.get("content-type");
  if (contentType) {
    headers.set("content-type", contentType);
  }

  return headers;
}

async function forward(request: NextRequest, path: string[]) {
  const session = await getServerSession(authOptions);
  const userID = session?.user?.id;
  if (!userID) {
    return unauthorizedResponse();
  }

  const apiToken = await signAPIToken({
    subject: userID,
    email: session?.user?.email,
    name: session?.user?.name,
    image: session?.user?.image,
  });

  const query = request.nextUrl.search ?? "";
  const url = `${apiBaseUrl}/v1/${path.join("/")}${query}`;
  const body = request.method === "GET" ? undefined : await request.text();

  const upstream = await fetch(url, {
    method: request.method,
    headers: buildForwardHeaders(request, apiToken),
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
