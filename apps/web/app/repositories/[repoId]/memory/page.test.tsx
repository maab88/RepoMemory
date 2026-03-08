import React from "react";
import { fireEvent, render, screen } from "@testing-library/react";

import RepositoryMemoryPage from "@/app/repositories/[repoId]/memory/page";

const useParamsMock = vi.fn();
const useRepositoryDetailMock = vi.fn();
const useRepositoryMemoryMock = vi.fn();
const useMemoryDetailMock = vi.fn();

vi.mock("next/navigation", () => ({
  useParams: () => useParamsMock(),
}));

vi.mock("@/lib/hooks/use-repository-detail", () => ({
  useRepositoryDetail: (repoId: string) => useRepositoryDetailMock(repoId),
}));

vi.mock("@/lib/hooks/use-repository-memory", () => ({
  useRepositoryMemory: (repoId: string) => useRepositoryMemoryMock(repoId),
}));

vi.mock("@/lib/hooks/use-memory-detail", () => ({
  useMemoryDetail: (repoId: string, memoryId: string | null) => useMemoryDetailMock(repoId, memoryId),
}));

describe("RepositoryMemoryPage", () => {
  beforeEach(() => {
    useParamsMock.mockReturnValue({ repoId: "repo-1" });
    useRepositoryDetailMock.mockReturnValue({
      data: {
        repository: {
          id: "repo-1",
          fullName: "octocat/repo-memory",
          htmlUrl: "https://github.com/octocat/repo-memory",
        },
      },
    });
    useMemoryDetailMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: null,
    });
  });

  it("renders empty state", () => {
    useRepositoryMemoryMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: { memoryEntries: [] },
    });

    render(<RepositoryMemoryPage />);
    expect(screen.getByText("No memory entries yet")).toBeInTheDocument();
  });

  it("renders timeline entries and filter state", () => {
    useRepositoryMemoryMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: {
        memoryEntries: [
          {
            id: "mem-1",
            repositoryId: "repo-1",
            organizationId: "org-1",
            type: "pr_summary",
            title: "PR memory",
            summary: "Summary",
            whyItMatters: "Matters",
            impactedAreas: ["sync"],
            risks: [],
            followUps: [],
            generatedBy: "deterministic",
            sourceUrl: "https://github.com/org/repo/pull/1",
            createdAt: "2026-03-07T12:00:00Z",
          },
        ],
      },
    });

    render(<RepositoryMemoryPage />);
    expect(screen.getByText("PR memory")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "pr summary" }));
    expect(screen.getByText("PR memory")).toBeInTheDocument();
  });

  it("opens detail drawer from memory card", () => {
    useRepositoryMemoryMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: {
        memoryEntries: [
          {
            id: "mem-1",
            repositoryId: "repo-1",
            organizationId: "org-1",
            type: "pr_summary",
            title: "PR memory",
            summary: "Summary",
            whyItMatters: "Matters",
            impactedAreas: ["sync"],
            risks: [],
            followUps: [],
            generatedBy: "deterministic",
            sourceUrl: "https://github.com/org/repo/pull/1",
            createdAt: "2026-03-07T12:00:00Z",
          },
        ],
      },
    });

    useMemoryDetailMock.mockImplementation((_repoId: string, memoryId: string | null) => ({
      isLoading: false,
      error: null,
      data: memoryId
        ? {
            memoryEntry: {
              id: "mem-1",
              repositoryId: "repo-1",
              organizationId: "org-1",
              type: "pr_summary",
              title: "PR memory",
              summary: "Summary",
              whyItMatters: "Matters",
              impactedAreas: ["sync"],
              risks: [],
              followUps: [],
              generatedBy: "deterministic",
              sourceUrl: "https://github.com/org/repo/pull/1",
              createdAt: "2026-03-07T12:00:00Z",
              sources: [{ sourceType: "pull_request", sourceUrl: "https://github.com/org/repo/pull/1", displayLabel: "PR #1" }],
            },
          }
        : null,
    }));

    render(<RepositoryMemoryPage />);
    fireEvent.click(screen.getByRole("button", { name: "Open detail" }));
    expect(screen.getByRole("dialog", { name: "Memory detail" })).toBeInTheDocument();
    expect(screen.getByText("PR #1")).toBeInTheDocument();
  });
});
