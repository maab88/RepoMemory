import React from "react";
import { render, screen } from "@testing-library/react";

import RepositoriesPage from "@/app/repositories/page";

const useRepositoriesMock = vi.fn();

vi.mock("@/lib/hooks/use-repositories", () => ({
  useRepositories: () => useRepositoriesMock(),
}));

describe("RepositoriesPage", () => {
  it("renders dashboard list", () => {
    useRepositoriesMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: {
        repositories: [
          {
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
            lastSyncStatus: "succeeded",
            lastSyncTime: "2026-03-07T12:30:00Z",
            pullRequestCount: 12,
            issueCount: 8,
            memoryEntryCount: 0,
          },
        ],
      },
    });

    render(<RepositoriesPage />);
    expect(screen.getByText("Repository Dashboard")).toBeInTheDocument();
    expect(screen.getByText("octocat/repo-memory")).toBeInTheDocument();
    expect(screen.getByText("PRs: 12")).toBeInTheDocument();
    expect(screen.getByText("Issues: 8")).toBeInTheDocument();
  });

  it("renders empty state", () => {
    useRepositoriesMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: { repositories: [] },
    });

    render(<RepositoriesPage />);
    expect(screen.getByText("No repositories imported yet")).toBeInTheDocument();
  });
});

