import { expect, test } from "@playwright/test";

test.use({
  extraHTTPHeaders: {
    "x-user-id": "playwright-user-1",
    "x-user-email": "playwright-user-1@example.com",
    "x-user-name": "Playwright User",
  },
});

test("repository import flow selects multiple repos and imports", async ({ page }) => {
  await page.route("**/api/v1/organizations", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: [{ id: "11111111-1111-1111-1111-111111111111", name: "Acme", slug: "acme", role: "owner" }],
      }),
    });
  });

  await page.route("**/api/v1/github/repositories", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          repositories: [
            {
              githubRepoId: "123",
              ownerLogin: "octocat",
              name: "repo-memory",
              fullName: "octocat/repo-memory",
              private: true,
              defaultBranch: "main",
              htmlUrl: "https://github.com/octocat/repo-memory",
              description: "Internal tools",
            },
            {
              githubRepoId: "456",
              ownerLogin: "octocat",
              name: "docs",
              fullName: "octocat/docs",
              private: false,
              defaultBranch: "main",
              htmlUrl: "https://github.com/octocat/docs",
            },
          ],
        },
      }),
    });
  });

  await page.route("**/api/v1/github/repositories/import", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          importedRepositories: [
            {
              id: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
              organizationId: "11111111-1111-1111-1111-111111111111",
              githubRepoId: "123",
              ownerLogin: "octocat",
              name: "repo-memory",
              fullName: "octocat/repo-memory",
              private: true,
              defaultBranch: "main",
              htmlUrl: "https://github.com/octocat/repo-memory",
              importedAt: "2026-03-07T12:00:00Z",
            },
            {
              id: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
              organizationId: "11111111-1111-1111-1111-111111111111",
              githubRepoId: "456",
              ownerLogin: "octocat",
              name: "docs",
              fullName: "octocat/docs",
              private: false,
              defaultBranch: "main",
              htmlUrl: "https://github.com/octocat/docs",
              importedAt: "2026-03-07T12:00:00Z",
            },
          ],
        },
      }),
    });
  });

  await page.goto("/onboarding/repositories");
  await page.selectOption("#org-select", "11111111-1111-1111-1111-111111111111");

  await page.getByLabel("Select octocat/repo-memory").check();
  await page.getByLabel("Select octocat/docs").check();

  await expect(page.getByText("Selected: 2")).toBeVisible();
  await page.getByRole("button", { name: "Import selected (2)" }).click();

  await expect(page.getByText("Imported repositories")).toBeVisible();
  await expect(page.getByRole("cell", { name: "octocat/repo-memory" })).toBeVisible();
  await expect(page.getByRole("cell", { name: "octocat/docs" })).toBeVisible();
});
