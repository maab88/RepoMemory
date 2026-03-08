import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import RepositoryDigestsPage from "@/app/repositories/[repoId]/digests/page";

const useParamsMock = vi.fn();
const useRepositoryDetailMock = vi.fn();
const useRepositoryDigestsMock = vi.fn();
const useGenerateDigestMock = vi.fn();
const useJobStatusMock = vi.fn();

vi.mock("next/navigation", () => ({
  useParams: () => useParamsMock(),
}));

vi.mock("@/lib/hooks/use-repository-detail", () => ({
  useRepositoryDetail: (repoId: string) => useRepositoryDetailMock(repoId),
}));

vi.mock("@/lib/hooks/use-repository-digests", () => ({
  useRepositoryDigests: (repoId: string) => useRepositoryDigestsMock(repoId),
}));

vi.mock("@/lib/hooks/use-generate-digest", () => ({
  useGenerateDigest: () => useGenerateDigestMock(),
}));

vi.mock("@/lib/hooks/use-job-status", () => ({
  useJobStatus: (jobId: string | null) => useJobStatusMock(jobId),
}));

describe("RepositoryDigestsPage", () => {
  beforeEach(() => {
    useParamsMock.mockReturnValue({ repoId: "repo-1" });
    useRepositoryDetailMock.mockReturnValue({
      data: {
        repository: {
          id: "repo-1",
          fullName: "octocat/repo-memory",
        },
      },
    });
    useGenerateDigestMock.mockReturnValue({
      isPending: false,
      mutateAsync: vi.fn().mockResolvedValue({ jobId: "job-1", status: "queued" }),
    });
    useJobStatusMock.mockReturnValue({
      data: null,
    });
  });

  it("renders empty state", () => {
    useRepositoryDigestsMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: { digests: [] },
    });

    render(
      <QueryClientProvider client={new QueryClient()}>
        <RepositoryDigestsPage />
      </QueryClientProvider>
    );

    expect(screen.getByText("No weekly digests yet")).toBeInTheDocument();
  });

  it("renders digest list and detail", () => {
    useRepositoryDigestsMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: {
        digests: [
          {
            id: "d-1",
            repositoryId: "repo-1",
            periodStart: "2026-03-02T00:00:00Z",
            periodEnd: "2026-03-08T23:59:59Z",
            title: "Weekly Digest: Mar 2 - Mar 8",
            summary: "3 merged PRs.",
            bodyMarkdown: "## Highlights",
            generatedBy: "deterministic",
            createdAt: "2026-03-08T20:00:00Z",
          },
        ],
      },
    });

    render(
      <QueryClientProvider client={new QueryClient()}>
        <RepositoryDigestsPage />
      </QueryClientProvider>
    );

    expect(screen.getAllByText("Weekly Digest: Mar 2 - Mar 8").length).toBeGreaterThan(0);
    expect(screen.getByText("## Highlights")).toBeInTheDocument();
  });

  it("queues digest generation", async () => {
    const mutateAsync = vi.fn().mockResolvedValue({ jobId: "job-1", status: "queued" });
    useGenerateDigestMock.mockReturnValue({
      isPending: false,
      mutateAsync,
    });
    useRepositoryDigestsMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: { digests: [] },
    });

    render(
      <QueryClientProvider client={new QueryClient()}>
        <RepositoryDigestsPage />
      </QueryClientProvider>
    );

    fireEvent.click(screen.getAllByRole("button", { name: "Generate weekly digest" })[0]);
    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith("repo-1");
    });
  });
});
