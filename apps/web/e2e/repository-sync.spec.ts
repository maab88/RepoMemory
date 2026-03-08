import { expect, test } from "@playwright/test";

test.use({
  extraHTTPHeaders: {
    "x-user-id": "playwright-user-1",
    "x-user-email": "playwright-user-1@example.com",
    "x-user-name": "Playwright User",
  },
});

test("repository detail sync flow shows queued/running/succeeded and refreshes persisted values", async ({ page }) => {
  const repoId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
  let detailCalls = 0;
  let jobCalls = 0;

  await page.route(`**/api/v1/repositories/${repoId}`, async (route) => {
    detailCalls += 1;
    const synced = detailCalls > 1;
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
            lastSyncStatus: synced ? "succeeded" : "not yet synced",
            lastSyncTime: synced ? "2026-03-07T12:30:00Z" : null,
            pullRequestCount: synced ? 12 : 0,
            issueCount: synced ? 8 : 0,
            memoryEntryCount: 0,
          },
        },
      }),
    });
  });

  await page.route("**/api/v1/repositories", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ data: { repositories: [] } }),
    });
  });

  await page.route(`**/api/v1/repositories/${repoId}/sync`, async (route) => {
    await route.fulfill({
      status: 202,
      contentType: "application/json",
      body: JSON.stringify({ data: { jobId: "job-123", status: "queued" } }),
    });
  });

  await page.route("**/api/v1/jobs/job-123", async (route) => {
    jobCalls += 1;
    const status = jobCalls === 1 ? "queued" : jobCalls === 2 ? "running" : "succeeded";
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          job: {
            id: "job-123",
            jobType: "repo.initial_sync",
            status,
            queueName: "default",
            attempts: 1,
            lastError: null,
            payload: {
              repositoryId: repoId,
              organizationId: "11111111-1111-1111-1111-111111111111",
              triggeredByUserId: "22222222-2222-2222-2222-222222222222",
            },
            startedAt: null,
            finishedAt: null,
            createdAt: "2026-03-07T12:00:00Z",
            updatedAt: "2026-03-07T12:00:00Z",
          },
        },
      }),
    });
  });

  await page.goto(`/repositories/${repoId}`);

  const prCard = page.locator("article").filter({ hasText: "Pull requests" });
  const issueCard = page.locator("article").filter({ hasText: "Issues" });
  await expect(prCard).toBeVisible();
  await expect(issueCard).toBeVisible();
  await expect(prCard.getByText("0")).toBeVisible();
  await expect(issueCard.getByText("0")).toBeVisible();

  await page.getByRole("button", { name: "Trigger initial sync" }).click();
  await expect(page.getByText("queued")).toBeVisible();
  await expect(page.getByText("running")).toBeVisible();
  await expect(page.getByText("succeeded")).toBeVisible();

  await expect(prCard.getByText("12")).toBeVisible();
  await expect(issueCard.getByText("8")).toBeVisible();
});
