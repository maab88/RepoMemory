import { expect, test } from "@playwright/test";

test.use({
  extraHTTPHeaders: {
    "x-user-id": "playwright-user-1",
    "x-user-email": "playwright-user-1@example.com",
    "x-user-name": "Playwright User",
  },
});

test("generate and browse repository digests", async ({ page }) => {
  const repoId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
  let jobCalls = 0;
  let digestReady = false;

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
            memoryEntryCount: 3,
          },
        },
      }),
    });
  });

  await page.route(`**/api/v1/repositories/${repoId}/digests/generate`, async (route) => {
    await route.fulfill({
      status: 202,
      contentType: "application/json",
      body: JSON.stringify({ data: { jobId: "job-digest-1", status: "queued" } }),
    });
  });

  await page.route("**/api/v1/jobs/job-digest-1", async (route) => {
    jobCalls += 1;
    const status = jobCalls === 1 ? "queued" : "succeeded";
    if (status === "succeeded") {
      digestReady = true;
    }
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          job: {
            id: "job-digest-1",
            jobType: "repo.generate_digest",
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

  await page.route(`**/api/v1/repositories/${repoId}/digests`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: {
          digests: digestReady
            ? [
                {
                  id: "dddddddd-dddd-dddd-dddd-dddddddddddd",
                  repositoryId: repoId,
                  periodStart: "2026-03-02T00:00:00Z",
                  periodEnd: "2026-03-08T23:59:59Z",
                  title: "Weekly Digest: Mar 2 - Mar 8",
                  summary: "3 merged PRs, 2 open issues, hotspots in sync and worker flows.",
                  bodyMarkdown: "## Highlights\n- 3 pull requests were merged this week.",
                  generatedBy: "deterministic",
                  createdAt: "2026-03-08T20:00:00Z",
                },
              ]
            : [],
        },
      }),
    });
  });

  await page.goto(`/repositories/${repoId}/digests`);
  await page.getByRole("button", { name: "Generate weekly digest" }).first().click();
  await expect(page.getByText("Digest job status: queued")).toBeVisible();
  await expect(page.getByText("Digest job status: succeeded")).toBeVisible();
  await expect(page.getByText("Weekly Digest: Mar 2 - Mar 8").first()).toBeVisible();
});
