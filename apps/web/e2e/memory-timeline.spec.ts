import { expect, test } from "@playwright/test";

test.use({
  extraHTTPHeaders: {
    "x-user-id": "playwright-user-1",
    "x-user-email": "playwright-user-1@example.com",
    "x-user-name": "Playwright User",
  },
});

test("browse repository memory timeline and open detail drawer", async ({ page }) => {
  const repoId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
  const memoryId = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb";

  await page.route(`**/api/v1/repositories/${repoId}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          repository: {
            id: repoId,
            organizationId: "11111111-1111-1111-1111-111111111111",
            githubRepoId: "123",
            ownerLogin: "octocat",
            name: "repo-memory",
            fullName: "octocat/repo-memory",
            private: true,
            defaultBranch: "main",
            htmlUrl: "https://github.com/octocat/repo-memory",
            description: "Internal tools",
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

  await page.route(`**/api/v1/repositories/${repoId}/memory`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          memoryEntries: [
            {
              id: memoryId,
              repositoryId: repoId,
              organizationId: "11111111-1111-1111-1111-111111111111",
              type: "pr_summary",
              title: "Refactored invoice retry flow",
              summary: "Moved retry scheduling into worker service.",
              whyItMatters: "Reduces duplicate retries.",
              impactedAreas: ["billing", "workers"],
              risks: ["retry timing drift"],
              followUps: ["monitor failed payment retry rate"],
              generatedBy: "deterministic",
              sourceUrl: "https://github.com/org/repo/pull/123",
              createdAt: "2026-03-07T12:00:00Z",
            },
          ],
        },
      }),
    });
  });

  await page.route(`**/api/v1/repositories/${repoId}/memory/${memoryId}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          memoryEntry: {
            id: memoryId,
            repositoryId: repoId,
            organizationId: "11111111-1111-1111-1111-111111111111",
            type: "pr_summary",
            title: "Refactored invoice retry flow",
            summary: "Moved retry scheduling into worker service.",
            whyItMatters: "Reduces duplicate retries.",
            impactedAreas: ["billing", "workers"],
            risks: ["retry timing drift"],
            followUps: ["monitor failed payment retry rate"],
            generatedBy: "deterministic",
            sourceUrl: "https://github.com/org/repo/pull/123",
            createdAt: "2026-03-07T12:00:00Z",
            sources: [{ sourceType: "pull_request", sourceUrl: "https://github.com/org/repo/pull/123", displayLabel: "PR #123" }],
          },
        },
      }),
    });
  });

  await page.goto(`/repositories/${repoId}`);
  await page.getByRole("link", { name: "Open memory timeline" }).click();
  await expect(page).toHaveURL(`/repositories/${repoId}/memory`);
  await expect(page.getByText("Refactored invoice retry flow")).toBeVisible();

  await page.getByRole("button", { name: "Open detail" }).click();
  await expect(page.getByRole("dialog", { name: "Memory detail" })).toBeVisible();
  await expect(page.getByText("PR #123")).toBeVisible();
});
