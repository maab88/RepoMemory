import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { RepositoryImportCard } from "@/components/repositories/repository-import-card";

const listOrganizationsMock = vi.fn();
const mutateAsyncMock = vi.fn();

vi.mock("@/lib/organizations-api", () => ({
  listOrganizations: () => listOrganizationsMock(),
}));

vi.mock("@/lib/hooks/use-github-repositories", () => ({
  useGitHubRepositories: () => ({
    isPending: false,
    error: null,
    data: {
      repositories: [
        {
          githubRepoId: "123",
          ownerLogin: "octocat",
          name: "repo-memory",
          fullName: "octocat/repo-memory",
          private: true,
          defaultBranch: "main",
          htmlUrl: "https://github.com/octocat/repo-memory",
          description: "Internal tools",
        },
      ],
    },
  }),
}));

vi.mock("@/lib/hooks/use-import-repositories", () => ({
  useImportRepositories: () => ({
    isPending: false,
    error: null,
    mutateAsync: (input: unknown) => mutateAsyncMock(input),
  }),
}));

function renderCard() {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={client}>
      <RepositoryImportCard />
    </QueryClientProvider>
  );
}

describe("RepositoryImportCard", () => {
  beforeEach(() => {
    listOrganizationsMock.mockResolvedValue([{ id: "org-1", name: "Acme", slug: "acme", role: "owner" }]);
    mutateAsyncMock.mockReset();
  });

  it("imports repositories successfully", async () => {
    mutateAsyncMock.mockResolvedValue({
      importedRepositories: [
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
          importedAt: "2026-03-07T12:00:00Z",
        },
      ],
    });

    renderCard();

    await waitFor(() => expect(screen.getByLabelText("Target organization")).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText("Target organization"), { target: { value: "org-1" } });
    fireEvent.click(screen.getByLabelText("Select octocat/repo-memory"));
    fireEvent.click(screen.getByRole("button", { name: "Import selected (1)" }));

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalledTimes(1);
      expect(screen.getByText("Imported repositories")).toBeInTheDocument();
    });
  });

  it("shows import error path", async () => {
    mutateAsyncMock.mockRejectedValue(new Error("failed"));

    renderCard();

    await waitFor(() => expect(screen.getByLabelText("Target organization")).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText("Target organization"), { target: { value: "org-1" } });
    fireEvent.click(screen.getByLabelText("Select octocat/repo-memory"));
    fireEvent.click(screen.getByRole("button", { name: "Import selected (1)" }));

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalledTimes(1);
      expect(screen.getByText("Import failed. Please review selection and try again.")).toBeInTheDocument();
    });
  });
});