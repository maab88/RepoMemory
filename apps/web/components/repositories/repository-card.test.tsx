import React from "react";
import { render, screen } from "@testing-library/react";

import { RepositoryCard } from "@/components/repositories/repository-card";

describe("RepositoryCard", () => {
  it("renders persisted repository summary", () => {
    render(
      <RepositoryCard
        repository={{
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
        }}
      />
    );

    expect(screen.getByText("octocat/repo-memory")).toBeInTheDocument();
    expect(screen.getByText("Sync: queued")).toBeInTheDocument();
  });
});

