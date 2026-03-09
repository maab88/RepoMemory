import { expect, test } from "@playwright/test";

test.use({
  extraHTTPHeaders: {
    "x-user-id": "playwright-user-1",
    "x-user-email": "playwright-user-1@example.com",
    "x-user-name": "Playwright User",
  },
});

test("expired GitHub token shows reconnect-required UI without white-screen", async ({ page }) => {
  await page.route("**/api/v1/organizations", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: [
          {
            id: "11111111-1111-1111-1111-111111111111",
            name: "Acme Engineering",
            slug: "acme-engineering",
            role: "OWNER",
          },
        ],
      }),
    });
  });

  await page.route("**/api/v1/github/repositories", async (route) => {
    await route.fulfill({
      status: 401,
      contentType: "application/json",
      body: JSON.stringify({
        error: {
          code: "GITHUB_RECONNECT_REQUIRED",
          message: "GitHub access token expired. Reconnect required.",
          requestId: "req-failure-path-1",
        },
      }),
    });
  });

  await page.goto("/onboarding/repositories");

  await expect(page.getByRole("heading", { name: "Reconnect GitHub required" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Open GitHub integrations" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Import selected (0)" })).toBeDisabled();
});
