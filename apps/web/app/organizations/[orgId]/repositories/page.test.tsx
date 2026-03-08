import React from "react";
import { render, screen } from "@testing-library/react";

import OrganizationRepositoriesPage from "@/app/organizations/[orgId]/repositories/page";

const useOrganizationRepositoriesMock = vi.fn();
const useParamsMock = vi.fn();

vi.mock("@/lib/hooks/use-organization-repositories", () => ({
  useOrganizationRepositories: (orgId: string) => useOrganizationRepositoriesMock(orgId),
}));

vi.mock("next/navigation", () => ({
  useParams: () => useParamsMock(),
}));

describe("OrganizationRepositoriesPage", () => {
  beforeEach(() => {
    useParamsMock.mockReturnValue({ orgId: "org-1" });
  });

  it("renders persisted repositories from API", () => {
    useOrganizationRepositoriesMock.mockReturnValue({
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
            lastSyncStatus: "queued",
            lastSyncTime: null,
            pullRequestCount: 0,
            issueCount: 0,
            memoryEntryCount: 0,
          },
        ],
      },
    });

    render(<OrganizationRepositoriesPage />);
    expect(screen.getByText("octocat/repo-memory")).toBeInTheDocument();
  });

  it("renders empty state", () => {
    useOrganizationRepositoriesMock.mockReturnValue({
      isLoading: false,
      error: null,
      data: { repositories: [] },
    });

    render(<OrganizationRepositoriesPage />);
    expect(screen.getByText("No repositories imported yet")).toBeInTheDocument();
  });
});
