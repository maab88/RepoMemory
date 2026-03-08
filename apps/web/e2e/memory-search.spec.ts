import { expect, test } from "@playwright/test";

test.use({
  extraHTTPHeaders: {
    "x-user-id": "playwright-user-1",
    "x-user-email": "playwright-user-1@example.com",
    "x-user-name": "Playwright User",
  },
});

test("search memory and open result timeline", async ({ page }) => {
  const organizationId = "11111111-1111-1111-1111-111111111111";
  const repositoryId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
  const memoryId = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb";

  await page.route("**/api/v1/organizations", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: [{ id: organizationId, name: "Acme Engineering", slug: "acme-engineering", role: "owner" }],
      }),
    });
  });

  await page.route("**/api/v1/repositories", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          repositories: [
            {
              id: repositoryId,
              organizationId,
              githubRepoId: "123",
              ownerLogin: "octocat",
              name: "repo-memory",
              fullName: "octocat/repo-memory",
              private: true,
              defaultBranch: "main",
              htmlUrl: "https://github.com/octocat/repo-memory",
              importedAt: "2026-03-07T12:00:00Z",
              lastSyncStatus: "succeeded",
              lastSyncTime: "2026-03-07T12:30:00Z",
              pullRequestCount: 12,
              issueCount: 8,
              memoryEntryCount: 1,
            },
          ],
        },
      }),
    });
  });

  await page.route("**/api/v1/memory/search**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          query: "retry",
          page: 1,
          pageSize: 20,
          total: 1,
          results: [
            {
              id: memoryId,
              repositoryId,
              repositoryName: "repo-memory",
              type: "pr_summary",
              title: "Refactored retry scheduling",
              summarySnippet: "Moved retry scheduling into the worker service.",
              sourceUrl: "https://github.com/org/repo/pull/123",
              createdAt: "2026-03-07T12:00:00Z",
            },
          ],
        },
      }),
    });
  });

  await page.route(`**/api/v1/repositories/${repositoryId}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          repository: {
            id: repositoryId,
            organizationId,
            githubRepoId: "123",
            ownerLogin: "octocat",
            name: "repo-memory",
            fullName: "octocat/repo-memory",
            private: true,
            defaultBranch: "main",
            htmlUrl: "https://github.com/octocat/repo-memory",
            importedAt: "2026-03-07T12:00:00Z",
            lastSyncStatus: "succeeded",
            lastSyncTime: "2026-03-07T12:30:00Z",
            pullRequestCount: 12,
            issueCount: 8,
            memoryEntryCount: 1,
          },
        },
      }),
    });
  });

  await page.route(`**/api/v1/repositories/${repositoryId}/memory`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          memoryEntries: [
            {
              id: memoryId,
              repositoryId,
              organizationId,
              type: "pr_summary",
              title: "Refactored retry scheduling",
              summary: "Moved retry scheduling into worker service.",
              whyItMatters: "Reduces duplicate retries.",
              impactedAreas: ["workers"],
              risks: ["retry drift"],
              followUps: ["watch job failures"],
              generatedBy: "deterministic",
              sourceUrl: "https://github.com/org/repo/pull/123",
              createdAt: "2026-03-07T12:00:00Z",
            },
          ],
        },
      }),
    });
  });

  await page.goto("/search");
  await page.getByLabel("Search Engineering Memory").fill("retry");
  await page.getByRole("button", { name: "Search" }).click();

  await expect(page.getByText("Refactored retry scheduling")).toBeVisible();
  const resultHref = await page
    .locator(`a[href="/repositories/${repositoryId}/memory?memoryId=${memoryId}"]`)
    .getAttribute("href");
  if (!resultHref) {
    throw new Error("expected result link href");
  }
  await page.goto(resultHref);
  await expect(page).toHaveURL(new RegExp(`/repositories/${repositoryId}/memory`));
  await expect(page.getByText("Refactored retry scheduling")).toBeVisible();
});
