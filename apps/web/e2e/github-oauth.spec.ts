import { expect, test } from "@playwright/test";

test.use({
  extraHTTPHeaders: {
    "x-user-id": "playwright-user-1",
    "x-user-email": "playwright-user-1@example.com",
    "x-user-name": "Playwright User",
  },
});

test("github oauth success flow shows connected UI", async ({ page }) => {
  await page.route("**/api/v1/github/connect/start", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          redirectUrl: "http://127.0.0.1:3000/integrations/github/callback?code=oauth-code-123&state=oauth-state-123",
        },
      }),
    });
  });

  await page.route("**/api/v1/github/callback?*", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          connected: true,
          account: {
            id: "11111111-1111-1111-1111-111111111111",
            githubLogin: "octocat",
            githubUserId: "583231",
            connectedAt: "2026-03-07T12:00:00Z",
          },
        },
      }),
    });
  });

  await page.goto("/settings/integrations/github");
  await page.getByRole("button", { name: "Connect GitHub" }).click();

  await expect(page).toHaveURL(/\/integrations\/github\/callback\?/);
  await expect(page.getByRole("heading", { name: "GitHub connected" })).toBeVisible();
  await expect(page.getByText("octocat")).toBeVisible();

  const content = await page.content();
  expect(content).not.toContain("gho_");
  expect(content).not.toContain("access_token");
});