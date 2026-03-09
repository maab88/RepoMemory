import { SignJWT } from "jose";

const encoder = new TextEncoder();

export type APITokenIdentity = {
  subject: string;
  email?: string | null;
  name?: string | null;
  image?: string | null;
};

export async function signAPIToken(identity: APITokenIdentity): Promise<string> {
  const secret = process.env.API_AUTH_JWT_SECRET ?? process.env.AUTH_SECRET;
  if (!secret) {
    throw new Error("API_AUTH_JWT_SECRET or AUTH_SECRET must be configured");
  }

  const issuer = process.env.API_AUTH_JWT_ISSUER ?? "repomemory-web";
  const audience = process.env.API_AUTH_JWT_AUDIENCE ?? "repomemory-api";

  return new SignJWT({
    email: identity.email ?? undefined,
    name: identity.name ?? undefined,
    picture: identity.image ?? undefined,
  })
    .setProtectedHeader({ alg: "HS256", typ: "JWT" })
    .setIssuer(issuer)
    .setAudience(audience)
    .setSubject(identity.subject)
    .setIssuedAt()
    .setExpirationTime("1h")
    .sign(encoder.encode(secret));
}
