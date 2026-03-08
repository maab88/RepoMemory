import { expect, test } from "@playwright/test";

function makeUUID() {
  return "11111111-1111-4111-8111-111111111111".replace(/1/g, () => Math.floor(Math.random() * 16).toString(16));
}

test.use({
  extraHTTPHeaders: {
    "x-user-id": "playwright-user-1",
    "x-user-email": "playwright-user-1@example.com",
    "x-user-name": "Playwright User",
  },
});

test("onboarding creates organization and lands on detail page", async ({ page }) => {
  const organizations: Array<{ id: string; name: string; slug: string; role: "owner" | "member" }> = [];

  await page.route("**/api/v1/organizations", async (route) => {
    const request = route.request();

    if (request.method() === "GET") {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ data: organizations }),
      });
      return;
    }

    if (request.method() === "POST") {
      const payload = request.postDataJSON() as { name: string };
      const id = makeUUID();
      const slug = payload.name.toLowerCase().replace(/\s+/g, "-");
      const created = { id, name: payload.name, slug, role: "owner" as const };
      organizations.push(created);

      await route.fulfill({
        status: 201,
        contentType: "application/json",
        body: JSON.stringify({ data: created }),
      });
      return;
    }

    await route.continue();
  });

  await page.route("**/api/v1/organizations/*", async (route) => {
    const id = route.request().url().split("/").pop() ?? "";
    const found = organizations.find((org) => org.id === id);

    if (!found) {
      await route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ error: { code: "not_found", message: "organization not found" } }),
      });
      return;
    }

    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ data: found }),
    });
  });

  const orgName = `Acme E2E ${Date.now()}`;

  await page.goto("/onboarding");
  await page.getByLabel("Organization name").fill(orgName);
  await page.getByRole("button", { name: "Create organization" }).click();

  await expect(page).toHaveURL(/\/organizations\//, { timeout: 15_000 });
  await expect(page.getByRole("heading", { name: orgName })).toBeVisible();

  await page.goto("/organizations");
  await expect(page.getByText(orgName)).toBeVisible();
});
