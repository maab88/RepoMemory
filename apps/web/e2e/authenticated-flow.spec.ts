import { expect, test } from "@playwright/test";
import { encode } from "next-auth/jwt";

test("authenticated session can access protected route and sign out", async ({ page, context }) => {
  const sessionToken = await encode({
    secret: "playwright-auth-secret",
    token: {
      sub: "dev:playwright@example.com",
      email: "playwright@example.com",
      name: "Playwright User",
    },
    maxAge: 60 * 60,
  });

  await context.addCookies([
    {
      name: "next-auth.session-token",
      value: sessionToken,
      url: "http://localhost:3000",
      httpOnly: true,
      sameSite: "Lax",
    },
  ]);

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
            role: "owner",
          },
        ],
      }),
    });
  });

  await page.goto("/organizations");
  await expect(page).toHaveURL(/\/organizations$/);
  await expect(page.getByRole("heading", { name: "Your organizations" })).toBeVisible();
  await expect(page.getByText("Acme Engineering")).toBeVisible();

  await page.getByRole("button", { name: "Sign out" }).click();
  await expect(page).toHaveURL(/\/sign-in/);
});
