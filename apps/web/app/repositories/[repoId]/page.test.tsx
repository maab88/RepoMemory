import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import RepositoryDetailPage from "@/app/repositories/[repoId]/page";

const useRepositoryDetailMock = vi.fn();
const triggerSyncMutateAsyncMock = vi.fn();
const generateMemoryMutateAsyncMock = vi.fn();
const useJobStatusMock = vi.fn();
const useParamsMock = vi.fn();

vi.mock("@/lib/hooks/use-repository-detail", () => ({
  useRepositoryDetail: (repoId: string) => useRepositoryDetailMock(repoId),
}));

vi.mock("@/lib/hooks/use-trigger-sync", () => ({
  useTriggerSync: () => ({
    isPending: false,
    mutateAsync: (repoId: string) => triggerSyncMutateAsyncMock(repoId),
  }),
}));

vi.mock("@/lib/hooks/use-generate-memory", () => ({
  useGenerateMemory: () => ({
    isPending: false,
    mutateAsync: (repoId: string) => generateMemoryMutateAsyncMock(repoId),
  }),
}));

vi.mock("@/lib/hooks/use-job-status", () => ({
  useJobStatus: (jobId: string | null) => useJobStatusMock(jobId),
}));

vi.mock("next/navigation", () => ({
  useParams: () => useParamsMock(),
}));

describe("RepositoryDetailPage", () => {
  beforeEach(() => {
    useParamsMock.mockReturnValue({ repoId: "repo-1" });
    useRepositoryDetailMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: {
        repository: {
          id: "repo-1",
          organizationId: "org-1",
          githubRepoId: "123",
          ownerLogin: "octocat",
          name: "repo-memory",
          fullName: "octocat/repo-memory",
          private: true,
          defaultBranch: "main",
          htmlUrl: "https://github.com/octocat/repo-memory",
          description: "Internal tools",
          importedAt: "2026-03-07T12:00:00Z",
          lastSyncStatus: "queued",
          lastSyncTime: null,
          pullRequestCount: 0,
          issueCount: 0,
          memoryEntryCount: 0,
        },
      },
    });
    useJobStatusMock.mockReturnValue({
      data: null,
    });
    triggerSyncMutateAsyncMock.mockResolvedValue({ jobId: "job-1", status: "queued" });
    generateMemoryMutateAsyncMock.mockResolvedValue({ jobId: "job-2", status: "queued" });
  });

  it("renders persisted repository details", () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <RepositoryDetailPage />
      </QueryClientProvider>
    );
    expect(screen.getByText("octocat/repo-memory")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Trigger initial sync" })).toBeInTheDocument();
    expect(screen.getByText("Pull requests")).toBeInTheDocument();
    expect(screen.getByText("Issues")).toBeInTheDocument();
  });

  it("triggers sync and requests job status", async () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <RepositoryDetailPage />
      </QueryClientProvider>
    );
    fireEvent.click(screen.getByRole("button", { name: "Trigger initial sync" }));

    await waitFor(() => {
      expect(triggerSyncMutateAsyncMock).toHaveBeenCalledWith("repo-1");
    });
  });

  it("triggers memory generation", async () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <RepositoryDetailPage />
      </QueryClientProvider>
    );

    fireEvent.click(screen.getByRole("button", { name: "Generate memory" }));

    await waitFor(() => {
      expect(generateMemoryMutateAsyncMock).toHaveBeenCalledWith("repo-1");
    });
  });
});
