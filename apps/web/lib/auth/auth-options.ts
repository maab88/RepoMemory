import type { NextAuthOptions } from "next-auth";
import GitHubProvider from "next-auth/providers/github";
import CredentialsProvider from "next-auth/providers/credentials";

const enableDevCredentials = process.env.AUTH_ENABLE_DEV_CREDENTIALS === "true";
const githubProvider = configuredGitHubProvider();

function configuredGitHubProvider() {
  const clientId = process.env.AUTH_GITHUB_ID ?? "";
  const clientSecret = process.env.AUTH_GITHUB_SECRET ?? "";
  if (!clientId || !clientSecret) {
    return null;
  }

  return GitHubProvider({
    clientId,
    clientSecret,
  });
}

export const authOptions: NextAuthOptions = {
  session: {
    strategy: "jwt",
  },
  pages: {
    signIn: "/sign-in",
  },
  providers: [
    ...(githubProvider ? [githubProvider] : []),
    ...(enableDevCredentials
      ? [
          CredentialsProvider({
            id: "dev-credentials",
            name: "Dev credentials",
            credentials: {
              email: { label: "Email", type: "email" },
              password: { label: "Password", type: "password" },
            },
            async authorize(credentials) {
              const expectedEmail = process.env.AUTH_DEV_EMAIL ?? "dev@example.com";
              const expectedPassword = process.env.AUTH_DEV_PASSWORD ?? "dev-password";
              const expectedName = process.env.AUTH_DEV_NAME ?? "Dev User";

              if (!credentials?.email || !credentials.password) {
                return null;
              }
              if (credentials.email !== expectedEmail || credentials.password !== expectedPassword) {
                return null;
              }

              return {
                id: `dev:${expectedEmail}`,
                email: expectedEmail,
                name: expectedName,
                image: null,
              };
            },
          }),
        ]
      : []),
  ],
  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.sub = user.id;
        token.email = user.email;
        token.name = user.name;
        token.picture = user.image;
      }
      return token;
    },
    async session({ session, token }) {
      if (session.user && token.sub) {
        session.user.id = token.sub;
        session.user.email = token.email;
        session.user.name = token.name;
        session.user.image = token.picture;
      }
      return session;
    },
  },
};
