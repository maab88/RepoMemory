import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 30_000,
  use: {
    baseURL: "http://127.0.0.1:3000",
    trace: "retain-on-failure",
  },
  webServer: {
    command: "corepack pnpm exec next dev --hostname 127.0.0.1 --port 3000",
    port: 3000,
    timeout: 120_000,
    reuseExistingServer: false,
    env: {
      NEXTAUTH_URL: "http://127.0.0.1:3000",
      AUTH_SECRET: "playwright-auth-secret",
      API_AUTH_JWT_SECRET: "playwright-auth-secret",
      API_AUTH_JWT_ISSUER: "repomemory-web",
      API_AUTH_JWT_AUDIENCE: "repomemory-api",
      AUTH_ENABLE_DEV_CREDENTIALS: "true",
      AUTH_DEV_EMAIL: "dev@example.com",
      AUTH_DEV_PASSWORD: "dev-password",
      AUTH_DEV_NAME: "Playwright User",
    },
  },
});
