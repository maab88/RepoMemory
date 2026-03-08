import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";

import SearchPage from "@/app/search/page";

const useRepositoriesMock = vi.fn();
const useMemorySearchMock = vi.fn();
const listOrganizationsMock = vi.fn();

vi.mock("@/lib/hooks/use-repositories", () => ({
  useRepositories: () => useRepositoriesMock(),
}));

vi.mock("@/lib/hooks/use-memory-search", () => ({
  useMemorySearch: (input: unknown) => useMemorySearchMock(input),
}));

vi.mock("@/lib/organizations-api", () => ({
  listOrganizations: () => listOrganizationsMock(),
}));

describe("SearchPage", () => {
  const renderPage = () =>
    render(
      <QueryClientProvider client={new QueryClient()}>
        <SearchPage />
      </QueryClientProvider>
    );

  beforeEach(() => {
    listOrganizationsMock.mockResolvedValue([
      { id: "org-1", name: "Acme", slug: "acme", role: "owner" },
      { id: "org-2", name: "Beta", slug: "beta", role: "owner" },
    ]);
    useRepositoriesMock.mockReturnValue({
      data: { repositories: [] },
      isLoading: false,
      error: null,
    });
    useMemorySearchMock.mockReturnValue({
      isLoading: false,
      isFetching: false,
      error: null,
      data: undefined,
    });
  });

  it("renders pre-search state", async () => {
    renderPage();
    expect(await screen.findByText("Start with a search term")).toBeInTheDocument();
  });

  it("shows no-results state", async () => {
    useMemorySearchMock.mockReturnValue({
      isLoading: false,
      isFetching: false,
      error: null,
      data: {
        query: "retry",
        page: 1,
        pageSize: 20,
        total: 0,
        results: [],
      },
    });

    renderPage();
    fireEvent.change(await screen.findByLabelText("Search Engineering Memory"), { target: { value: "retry" } });
    fireEvent.click(screen.getByRole("button", { name: "Search" }));
    expect(await screen.findByText("No memory results found")).toBeInTheDocument();
  });

  it("shows results state", async () => {
    useMemorySearchMock.mockReturnValue({
      isLoading: false,
      isFetching: false,
      error: null,
      data: {
        query: "retry",
        page: 1,
        pageSize: 20,
        total: 1,
        results: [
          {
            id: "mem-1",
            repositoryId: "repo-1",
            repositoryName: "repo-memory",
            type: "pr_summary",
            title: "Retry refactor",
            summarySnippet: "Moved retry scheduling.",
            sourceUrl: "https://github.com/org/repo/pull/1",
            createdAt: "2026-03-07T12:00:00Z",
          },
        ],
      },
    });

    renderPage();
    fireEvent.change(await screen.findByLabelText("Search Engineering Memory"), { target: { value: "retry" } });
    fireEvent.click(screen.getByRole("button", { name: "Search" }));
    expect(await screen.findByText("Retry refactor")).toBeInTheDocument();
  });

  it("shows error state", async () => {
    useMemorySearchMock.mockReturnValue({
      isLoading: false,
      isFetching: false,
      error: new Error("failed"),
      data: undefined,
    });

    renderPage();
    fireEvent.change(await screen.findByLabelText("Search Engineering Memory"), { target: { value: "retry" } });
    fireEvent.click(screen.getByRole("button", { name: "Search" }));
    expect(await screen.findByText("Could not load memory search results. Please verify organization access and try again.")).toBeInTheDocument();
  });
});
