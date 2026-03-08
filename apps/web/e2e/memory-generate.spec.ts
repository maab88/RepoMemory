import { expect, test } from "@playwright/test";

test.use({
  extraHTTPHeaders: {
    "x-user-id": "playwright-user-1",
    "x-user-email": "playwright-user-1@example.com",
    "x-user-name": "Playwright User",
  },
});

test("generate memory from repository detail and see persisted timeline entries", async ({ page }) => {
  const repoId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
  let jobCalls = 0;
  let generationSucceeded = false;

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
            memoryEntryCount: generationSucceeded ? 1 : 0,
          },
        },
      }),
    });
  });

  await page.route(`**/api/v1/repositories/${repoId}/memory/generate`, async (route) => {
    await route.fulfill({
      status: 202,
      contentType: "application/json",
      body: JSON.stringify({ data: { jobId: "job-memory-1", status: "queued" } }),
    });
  });

  await page.route("**/api/v1/jobs/job-memory-1", async (route) => {
    jobCalls += 1;
    const status = jobCalls === 1 ? "queued" : jobCalls === 2 ? "running" : "succeeded";
    if (status === "succeeded") {
      generationSucceeded = true;
    }
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          job: {
            id: "job-memory-1",
            jobType: "repo.generate_memory",
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

  await page.route(`**/api/v1/repositories/${repoId}/memory`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          memoryEntries: generationSucceeded
            ? [
                {
                  id: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
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
              ]
            : [],
        },
      }),
    });
  });

  await page.goto(`/repositories/${repoId}`);
  await page.getByRole("button", { name: "Generate memory" }).click();
  await expect(page.getByText("Memory job status: queued")).toBeVisible();
  await expect(page.getByText("Memory job status: running")).toBeVisible();
  await expect(page.getByText("Memory job status: succeeded")).toBeVisible();

  await page.locator(`a[href="/repositories/${repoId}/memory"]`).click();
  await expect(page).toHaveURL(`/repositories/${repoId}/memory`);
  await expect(page.getByText("Refactored invoice retry flow")).toBeVisible();
});
